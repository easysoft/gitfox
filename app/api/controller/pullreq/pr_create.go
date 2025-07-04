// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pullreq

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/auth"
	pullreqevents "github.com/easysoft/gitfox/app/events/pullreq"
	"github.com/easysoft/gitfox/app/services/instrument"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/git/sha"
	"github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
)

type CreateInput struct {
	IsDraft bool `json:"is_draft"`

	Title       string `json:"title"`
	Description string `json:"description"`

	SourceRepoRef string `json:"source_repo_ref"`
	SourceBranch  string `json:"source_branch"`
	TargetBranch  string `json:"target_branch"`
}

func (in *CreateInput) Sanitize() error {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)

	if err := validateTitle(in.Title); err != nil {
		return err
	}

	if err := validateDescription(in.Description); err != nil {
		return err
	}

	return nil
}

// Create creates a new pull request.
func (c *Controller) Create(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	in *CreateInput,
) (*types.PullReq, error) {
	if err := in.Sanitize(); err != nil {
		return nil, err
	}

	targetRepo, err := c.getRepoCheckAccess(ctx, session, repoRef, enum.PermissionRepoPush)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire access to target repo: %w", err)
	}

	if targetRepo.Mirror {
		return nil, store.ErrReadOnlyMirrorRepo
	}

	sourceRepo := targetRepo
	if in.SourceRepoRef != "" {
		sourceRepo, err = c.getRepoCheckAccess(ctx, session, in.SourceRepoRef, enum.PermissionRepoPush)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire access to source repo: %w", err)
		}
		if sourceRepo.Mirror {
			return nil, store.ErrReadOnlyMirrorRepo
		}
	}

	if sourceRepo.ID == targetRepo.ID && in.TargetBranch == in.SourceBranch {
		return nil, usererror.BadRequest("target and source branch can't be the same")
	}

	var sourceSHA sha.SHA

	if sourceSHA, err = c.verifyBranchExistence(ctx, sourceRepo, in.SourceBranch); err != nil {
		return nil, err
	}

	if _, err = c.verifyBranchExistence(ctx, targetRepo, in.TargetBranch); err != nil {
		return nil, err
	}

	if err = c.checkIfAlreadyExists(ctx, targetRepo.ID, sourceRepo.ID, in.TargetBranch, in.SourceBranch); err != nil {
		return nil, err
	}

	mergeBaseResult, err := c.git.MergeBase(ctx, git.MergeBaseParams{
		ReadParams: git.ReadParams{RepoUID: sourceRepo.GitUID},
		Ref1:       in.SourceBranch,
		Ref2:       in.TargetBranch,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find merge base: %w", err)
	}

	mergeBaseSHA := mergeBaseResult.MergeBaseSHA

	if mergeBaseSHA == sourceSHA {
		return nil, usererror.BadRequest("The source branch doesn't contain any new commits")
	}

	prStats, err := c.git.DiffStats(ctx, &git.DiffParams{
		ReadParams: git.ReadParams{RepoUID: targetRepo.GitUID},
		BaseRef:    mergeBaseSHA.String(),
		HeadRef:    sourceSHA.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR diff stats: %w", err)
	}

	targetRepo, err = c.repoStore.UpdateOptLock(ctx, targetRepo, func(repo *types.Repository) error {
		repo.PullReqSeq++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to acquire PullReqSeq number: %w", err)
	}

	pr := newPullReq(session, targetRepo.PullReqSeq, sourceRepo, targetRepo, in, sourceSHA, mergeBaseSHA)
	pr.Stats = types.PullReqStats{
		DiffStats:       types.NewDiffStats(prStats.Commits, prStats.FilesChanged, prStats.Additions, prStats.Deletions),
		Conversations:   0,
		UnresolvedCount: 0,
	}

	err = c.pullreqStore.Create(ctx, pr)
	if err != nil {
		return nil, fmt.Errorf("pullreq creation failed: %w", err)
	}

	c.eventReporter.Created(ctx, &pullreqevents.CreatedPayload{
		Base:         eventBase(pr, &session.Principal),
		SourceBranch: in.SourceBranch,
		TargetBranch: in.TargetBranch,
		SourceSHA:    sourceSHA.String(),
	})

	if err = c.sseStreamer.Publish(ctx, targetRepo.ParentID, enum.SSETypePullRequestUpdated, pr); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to publish PR changed event")
	}

	c.notifyCodeOwners(ctx, session, targetRepo, pr)
	err = c.instrumentation.Track(ctx, instrument.Event{
		Type:      instrument.EventTypeCreatePullRequest,
		Principal: session.Principal.ToPrincipalInfo(),
		Path:      sourceRepo.Path,
		Properties: map[instrument.Property]any{
			instrument.PropertyRepositoryID:   sourceRepo.ID,
			instrument.PropertyRepositoryName: sourceRepo.Identifier,
			instrument.PropertyPullRequestID:  pr.Number,
		},
	})
	if err != nil {
		log.Ctx(ctx).Warn().Msgf("failed to insert instrumentation record for create pull request operation: %s", err)
	}

	return pr, nil
}

func (c *Controller) notifyCodeOwners(ctx context.Context, session *auth.Session, repo *types.Repository, pr *types.PullReq) {
	reviewers, err := c.reviewerStore.List(ctx, pr.ID)
	if err != nil {
		return
	}

	ownerEvaluation, err := c.codeOwners.Evaluate(ctx, repo, pr, reviewers)
	if err != nil {
		return
	}

	codeOwners := make(map[int64]types.PrincipalInfo)

	for _, entry := range ownerEvaluation.EvaluationEntries {
		for _, eva := range entry.OwnerEvaluations {
			codeOwners[eva.Owner.ID] = eva.Owner
		}
	}

	for _, owner := range codeOwners {
		c.eventReporter.ReviewerAdded(ctx, &pullreqevents.ReviewerAddedPayload{
			Base:       eventBase(pr, &session.Principal),
			ReviewerID: owner.ID,
		})
	}
}

// newPullReq creates new pull request object.
func newPullReq(
	session *auth.Session,
	number int64,
	sourceRepo *types.Repository,
	targetRepo *types.Repository,
	in *CreateInput,
	sourceSHA, mergeBaseSHA sha.SHA,
) *types.PullReq {
	now := time.Now().UnixMilli()
	return &types.PullReq{
		ID:                0, // the ID will be populated in the data layer
		Version:           0,
		Number:            number,
		CreatedBy:         session.Principal.ID,
		Created:           now,
		Updated:           now,
		Edited:            now,
		State:             enum.PullReqStateOpen,
		IsDraft:           in.IsDraft,
		Title:             in.Title,
		Description:       in.Description,
		SourceRepoID:      sourceRepo.ID,
		SourceBranch:      in.SourceBranch,
		SourceSHA:         sourceSHA.String(),
		TargetRepoID:      targetRepo.ID,
		TargetBranch:      in.TargetBranch,
		ActivitySeq:       0,
		MergedBy:          nil,
		Merged:            nil,
		MergeMethod:       nil,
		MergeBaseSHA:      mergeBaseSHA.String(),
		MergeCheckStatus:  enum.MergeCheckStatusUnchecked,
		RebaseCheckStatus: enum.MergeCheckStatusUnchecked,
		Author:            *session.Principal.ToPrincipalInfo(),
		Merger:            nil,
	}
}
