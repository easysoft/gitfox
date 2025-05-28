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

package githook

import (
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/bootstrap"
	events "github.com/easysoft/gitfox/app/events/git"
	pullreqevents "github.com/easysoft/gitfox/app/events/pullreq"
	repoevents "github.com/easysoft/gitfox/app/events/repo"
	"github.com/easysoft/gitfox/errors"
	"github.com/easysoft/gitfox/git"
	gitenum "github.com/easysoft/gitfox/git/enum"
	"github.com/easysoft/gitfox/git/hook"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/gotidy/ptr"
	"github.com/rs/zerolog/log"
)

const (
	// gitReferenceNamePrefixBranch is the prefix of references of type branch.
	gitReferenceNamePrefixBranch = "refs/heads/"

	// gitReferenceNamePrefixTag is the prefix of references of type tag.
	gitReferenceNamePrefixTag = "refs/tags/"

	// gitReferenceNamePrefixTag is the prefix of pull req references.
	gitReferenceNamePullReq = "refs/pullreq/"

	// AGit Flow

	// gitReferenceNameAgitPullReq special ref to create a pull request: refs/for/<targe-branch>/<topic-branch>
	// or refs/for/<targe-branch> -o topic='<topic-branch>'
)

// PostReceive executes the post-receive hook for a git repository.
func (c *Controller) PostReceive(
	ctx context.Context,
	rgit RestrictedGIT,
	session *auth.Session,
	in types.GithookPostReceiveInput,
) (hook.Output, error) {
	repo, err := c.getRepoCheckAccess(ctx, session, in.RepoID, enum.PermissionRepoPush)
	if err != nil {
		return hook.Output{}, err
	}
	// create output object and have following messages fill its messages
	out := hook.Output{}
	// update default branch based on ref update info on empty repos.
	// as the branch could be different than the configured default value.
	c.handleEmptyRepoPush(ctx, repo, in.PostReceiveInput, &out)

	// report ref events if repo is in an active state (best effort)
	if repo.State == enum.RepoStateActive {
		c.reportReferenceEvents(ctx, rgit, repo, in.PrincipalID, in.PostReceiveInput)
	}

	// handle branch updates related to PRs - best effort
	c.handlePRMessaging(ctx, repo, in.PrincipalID, in.PostReceiveInput, &out)

	err = c.postReceiveExtender.Extend(ctx, rgit, session, repo, in, &out)
	if err != nil {
		return hook.Output{}, fmt.Errorf("failed to extend post-receive hook: %w", err)
	}

	return out, nil
}

// reportReferenceEvents is reporting reference events to the event system.
// NOTE: keep best effort for now as it doesn't change the outcome of the git operation.
// TODO: in the future we might want to think about propagating errors so user is aware of events not being triggered.
func (c *Controller) reportReferenceEvents(
	ctx context.Context,
	rgit RestrictedGIT,
	repo *types.Repository,
	principalID int64,
	in hook.PostReceiveInput,
) {
	for _, refUpdate := range in.RefUpdates {
		switch {
		case strings.HasPrefix(refUpdate.Ref, gitReferenceNamePrefixBranch):
			c.reportBranchEvent(ctx, rgit, repo, principalID, in.Environment, refUpdate, in, enum.PullRequestFlowGithub)
		case strings.HasPrefix(refUpdate.Ref, gitReferenceNamePrefixTag):
			c.reportTagEvent(ctx, repo, principalID, refUpdate)
		default:
			// Ignore any other references in post-receive
		}
	}
}

func (c *Controller) reportBranchEvent(
	ctx context.Context,
	rgit RestrictedGIT,
	repo *types.Repository,
	principalID int64,
	env hook.Environment,
	branchUpdate hook.ReferenceUpdate,
	in hook.PostReceiveInput,
	flow enum.PullRequestFlow,
) {
	ref := branchUpdate.Ref
	log.Ctx(ctx).Debug().Msgf("post receive reporting branch event for ref %q(source: %q), %v", ref, branchUpdate.Ref, flow.String())
	switch {
	case branchUpdate.Old.IsNil():
		count, _ := c.pullreqStore.Count(ctx, &types.PullReqFilter{
			SourceRepoID: repo.ID,
			SourceBranch: strings.Replace(ref, "refs/heads/", "", 1),
			States:       []enum.PullReqState{enum.PullReqStateOpen},
		})
		if flow == enum.PullRequestFlowZentao && count == 1 {
			log.Ctx(ctx).Debug().Msgf("ztflow branch %q update", ref)
		} else {
			log.Ctx(ctx).Debug().Msgf("branch %q created", ref)
			c.gitReporter.BranchCreated(ctx, &events.BranchCreatedPayload{
				RepoID:      repo.ID,
				PrincipalID: principalID,
				Ref:         ref,
				SHA:         branchUpdate.New.String(),
			})
		}
	case branchUpdate.New.IsNil():
		log.Ctx(ctx).Debug().Msgf("branch %q deleted", ref)
		c.gitReporter.BranchDeleted(ctx, &events.BranchDeletedPayload{
			RepoID:      repo.ID,
			PrincipalID: principalID,
			Ref:         ref,
			SHA:         branchUpdate.Old.String(),
		})
	default:
		// A force update event might trigger some additional operations that aren't required
		// for ordinary updates (force pushes alter the commit history of a branch).
		forced, err := isForcePush(ctx, rgit, repo.GitUID, env.AlternateObjectDirs, branchUpdate)
		if err != nil {
			// In case of an error consider this a forced update. In post-update the branch has already been updated,
			// so there's less harm in declaring the update as forced.
			forced = true
			log.Ctx(ctx).Warn().Err(err).
				Str("ref", branchUpdate.Ref).
				Msg("failed to check ancestor")
		}

		c.gitReporter.BranchUpdated(ctx, &events.BranchUpdatedPayload{
			RepoID:      repo.ID,
			PrincipalID: principalID,
			Ref:         ref,
			OldSHA:      branchUpdate.Old.String(),
			NewSHA:      branchUpdate.New.String(),
			Forced:      forced,
		})
	}
}

func (c *Controller) reportTagEvent(
	ctx context.Context,
	repo *types.Repository,
	principalID int64,
	tagUpdate hook.ReferenceUpdate,
) {
	switch {
	case tagUpdate.Old.IsNil():
		c.gitReporter.TagCreated(ctx, &events.TagCreatedPayload{
			RepoID:      repo.ID,
			PrincipalID: principalID,
			Ref:         tagUpdate.Ref,
			SHA:         tagUpdate.New.String(),
		})
	case tagUpdate.New.IsNil():
		c.gitReporter.TagDeleted(ctx, &events.TagDeletedPayload{
			RepoID:      repo.ID,
			PrincipalID: principalID,
			Ref:         tagUpdate.Ref,
			SHA:         tagUpdate.Old.String(),
		})
	default:
		c.gitReporter.TagUpdated(ctx, &events.TagUpdatedPayload{
			RepoID:      repo.ID,
			PrincipalID: principalID,
			Ref:         tagUpdate.Ref,
			OldSHA:      tagUpdate.Old.String(),
			NewSHA:      tagUpdate.New.String(),
			// tags can only be force updated!
			Forced: true,
		})
	}
}

// handlePRMessaging checks any single branch push for pr information and returns an according response if needed.
// TODO: If it is a new branch, or an update on a branch without any PR, it also sends out an SSE for pr creation.
func (c *Controller) handlePRMessaging(
	ctx context.Context,
	repo *types.Repository,
	principalID int64,
	in hook.PostReceiveInput,
	out *hook.Output,
) {
	// skip anything that was a batch push / isn't branch related / isn't updating/creating a branch.
	if len(in.RefUpdates) != 1 || in.RefUpdates[0].New.IsNil() {
		return
	}

	// for now we only care about first branch that was pushed.
	branchName := in.RefUpdates[0].Ref[len(gitReferenceNamePrefixBranch):]
	branchPre := ""
	c.suggestPullRequest(ctx, repo, branchName, branchPre, principalID, out)

	// TODO: store latest pushed branch for user in cache and send out SSE
}

func (c *Controller) suggestPullRequest(
	ctx context.Context,
	repo *types.Repository,
	branchName, branchPre string,
	principalID int64,
	out *hook.Output,
) {
	if branchName == repo.DefaultBranch && len(branchPre) == 0 {
		// Don't suggest a pull request if this is a push to the default branch.
		return
	}

	if len(branchPre) > 0 {
		branchName = fmt.Sprintf("%s%s", branchPre, branchName)
	}

	// do we have a PR related to it?
	prs, err := c.pullreqStore.List(ctx, &types.PullReqFilter{
		Page: 1,
		// without forks we expect at most one PR (keep 2 to not break when forks are introduced)
		Size:         2,
		SourceRepoID: repo.ID,
		SourceBranch: strings.TrimPrefix(branchName, "refs/heads/"),
		// we only care about open PRs - merged/closed will lead to "create new PR" message
		States: []enum.PullReqState{enum.PullReqStateOpen},
		Order:  enum.OrderAsc,
		Sort:   enum.PullReqSortCreated,
		// don't care about the PR description, omit it from the response
		ExcludeDescription: true,
	})
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msgf(
			"failed to find pullrequests for branch '%s' originating from repo '%s'",
			branchName,
			repo.Path,
		)
		return
	}

	// for already existing PRs, print them to users terminal for easier access.
	if len(prs) > 0 {
		log.Ctx(ctx).Debug().Int("pr_count", len(prs)).Msgf("found PRs for branch %s", branchName)
		msgs := make([]string, 2*len(prs)+1)
		msgs[0] = fmt.Sprintf("Branch %q has open PRs:", branchName)
		for i, pr := range prs {
			msgs[2*i+1] = fmt.Sprintf("  (#%d) %s", pr.Number, pr.Title)
			msgs[2*i+2] = "    " + c.urlProvider.GenerateUIPRURL(ctx, repo.Path, pr.Number)
		}
		out.Messages = append(out.Messages, msgs...)
		return
	}
	log.Ctx(ctx).Debug().Msgf("create new %s PR", branchName)
	// this is a new PR!
	out.Messages = append(out.Messages,
		fmt.Sprintf("Create a pull request for %q by visiting:", branchName),
		"  "+c.urlProvider.GenerateUICompareURL(ctx, repo.Path, repo.DefaultBranch, branchName),
	)
}

func (c *Controller) checkIfAlreadyExists(ctx context.Context,
	targetRepoID, sourceRepoID int64, targetBranch, sourceBranch string,
) error {
	sourceBranch = strings.TrimPrefix(sourceBranch, "refs/heads/")
	existing, err := c.pullreqStore.List(ctx, &types.PullReqFilter{
		SourceRepoID: sourceRepoID,
		SourceBranch: sourceBranch,
		TargetRepoID: targetRepoID,
		TargetBranch: targetBranch,
		States:       []enum.PullReqState{enum.PullReqStateOpen},
		Size:         1,
		Sort:         enum.PullReqSortNumber,
		Order:        enum.OrderAsc,
	})
	if err != nil {
		return fmt.Errorf("failed to get existing pull requests: %w", err)
	}
	if len(existing) > 0 {
		return usererror.ConflictWithPayload(
			"a pull request for this target and source branch already exists",
			map[string]any{
				"type":   "pr already exists",
				"number": existing[0].Number,
				"title":  existing[0].Title,
			},
		)
	}

	return nil
}

func (c *Controller) verifyBranchExistence(ctx context.Context,
	repo *types.Repository, branch string,
) (string, error) {
	if branch == "" {
		return "", usererror.BadRequest("branch name can't be empty")
	}
	branch = strings.TrimPrefix(branch, "refs/heads/")
	ref, err := c.git.GetRef(ctx,
		git.GetRefParams{
			ReadParams: git.ReadParams{RepoUID: repo.GitUID},
			Name:       branch,
			Type:       gitenum.RefTypeBranch,
		})
	if errors.AsStatus(err) == errors.StatusNotFound {
		return "", usererror.BadRequest(
			fmt.Sprintf("branch %q does not exist in the repository %q", branch, repo.Identifier))
	}
	if err != nil {
		return "", fmt.Errorf(
			"failed to check existence of the branch %q in the repository %q: %w",
			branch, repo.Identifier, err)
	}

	return ref.SHA.String(), nil
}

// handleEmptyRepoPush updates repo default branch on empty repos if push contains branches.
func (c *Controller) handleEmptyRepoPush(
	ctx context.Context,
	repo *types.Repository,
	in hook.PostReceiveInput,
	out *hook.Output,
) {
	if !repo.IsEmpty {
		return
	}

	var newDefaultBranch string
	// update default branch if corresponding branch does not exist
	for _, refUpdate := range in.RefUpdates {
		if strings.HasPrefix(refUpdate.Ref, gitReferenceNamePrefixBranch) && !refUpdate.New.IsNil() {
			branchName := refUpdate.Ref[len(gitReferenceNamePrefixBranch):]
			if branchName == repo.DefaultBranch {
				newDefaultBranch = branchName
				break
			}
			// use the first pushed branch if default branch is not present.
			if newDefaultBranch == "" {
				newDefaultBranch = branchName
			}
		}
	}
	if newDefaultBranch == "" {
		out.Error = ptr.String(usererror.ErrEmptyRepoNeedsBranch.Error())
		return
	}

	oldName := repo.DefaultBranch
	var err error
	repo, err = c.repoStore.UpdateOptLock(ctx, repo, func(r *types.Repository) error {
		r.IsEmpty = false
		r.DefaultBranch = newDefaultBranch
		return nil
	})
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msgf("failed to update the repo default branch to %s and is_empty to false",
			newDefaultBranch)
		return
	}

	if repo.DefaultBranch != oldName {
		c.repoReporter.DefaultBranchUpdated(ctx, &repoevents.DefaultBranchUpdatedPayload{
			RepoID:      repo.ID,
			PrincipalID: bootstrap.NewSystemServiceSession().Principal.ID,
			OldName:     oldName,
			NewName:     repo.DefaultBranch,
		})
	}
}

func eventBase(pr *types.PullReq, principal *types.Principal) pullreqevents.Base {
	return pullreqevents.Base{
		PullReqID:    pr.ID,
		SourceRepoID: pr.SourceRepoID,
		TargetRepoID: pr.TargetRepoID,
		Number:       pr.Number,
		PrincipalID:  principal.ID,
	}
}
