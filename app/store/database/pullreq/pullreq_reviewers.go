// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pullreq

import (
	"context"
	"fmt"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var _ store.PullReqReviewerStore = (*ReviewerOrmStore)(nil)

const maxPullRequestReviewers = 100

// NewPullReqReviewerOrmStore returns a new PullReqReviewerOrmStore.
func NewPullReqReviewerOrmStore(db *gorm.DB,
	pCache store.PrincipalInfoCache) *ReviewerOrmStore {
	return &ReviewerOrmStore{
		db:     db,
		pCache: pCache,
	}
}

// ReviewerOrmStore implements store.PullReqReviewerStore backed by a relational database.
type ReviewerOrmStore struct {
	db     *gorm.DB
	pCache store.PrincipalInfoCache
}

// pullReqReviewer is used to fetch pull request reviewer data from the database.
type pullReqReviewer struct {
	PullReqID   int64 `gorm:"column:pullreq_reviewer_pullreq_id"`
	PrincipalID int64 `gorm:"column:pullreq_reviewer_principal_id"`
	CreatedBy   int64 `gorm:"column:pullreq_reviewer_created_by"`
	Created     int64 `gorm:"column:pullreq_reviewer_created"`
	Updated     int64 `gorm:"column:pullreq_reviewer_updated"`

	RepoID         int64                    `gorm:"column:pullreq_reviewer_repo_id"`
	Type           enum.PullReqReviewerType `gorm:"column:pullreq_reviewer_type"`
	LatestReviewID null.Int                 `gorm:"column:pullreq_reviewer_latest_review_id"`

	ReviewDecision enum.PullReqReviewDecision `gorm:"column:pullreq_reviewer_review_decision"`
	SHA            string                     `gorm:"column:pullreq_reviewer_sha"`
}

const (
	tableReviewer = "pullreq_reviewers"
)

// Find finds the pull request reviewer by pull request id and principal id.
func (s *ReviewerOrmStore) Find(ctx context.Context, prID, principalID int64) (*types.PullReqReviewer, error) {
	dst := &pullReqReviewer{}
	q := pullReqReviewer{PullReqID: prID, PrincipalID: principalID}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableReviewer).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pull request reviewer")
	}

	return s.mapPullReqReviewer(ctx, dst), nil
}

// Create creates a new pull request reviewer.
func (s *ReviewerOrmStore) Create(ctx context.Context, v *types.PullReqReviewer) error {
	dbObj := mapInternalPullReqReviewer(v)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableReviewer).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to insert pull request reviewer")
	}

	return nil
}

// Update updates the pull request reviewer.
func (s *ReviewerOrmStore) Update(ctx context.Context, v *types.PullReqReviewer) error {
	updatedAt := time.Now()

	dbv := mapInternalPullReqReviewer(v)
	dbv.Updated = updatedAt.UnixMilli()

	updatedFields := []string{"Updated", "LatestReviewID", "ReviewDecision", "SHA"}

	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableReviewer).
		Where(&pullReqReviewer{PullReqID: v.PullReqID, PrincipalID: v.PrincipalID}).
		Select(updatedFields).Updates(dbv)
	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update pull request activity")
	}

	v.Updated = dbv.Updated

	return nil
}

// Delete deletes the pull request reviewer.
func (s *ReviewerOrmStore) Delete(ctx context.Context, prID, reviewerID int64) error {
	q := pullReqReviewer{PullReqID: prID, PrincipalID: reviewerID}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableReviewer).Where(&q).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "delete reviewer query failed")
	}
	return nil
}

// List returns a list of pull reviewers for a pull request.
func (s *ReviewerOrmStore) List(ctx context.Context, prID int64) ([]*types.PullReqReviewer, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableReviewer).
		Where(&pullReqReviewer{PullReqID: prID}).
		Order("pullreq_reviewer_created asc").Limit(maxPullRequestReviewers)

	dst := make([]*pullReqReviewer, 0)

	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing pull request reviewer list query")
	}

	result, err := s.mapSlicePullReqReviewer(ctx, dst)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func mapPullReqReviewer(v *pullReqReviewer) *types.PullReqReviewer {
	m := &types.PullReqReviewer{
		PullReqID:      v.PullReqID,
		PrincipalID:    v.PrincipalID,
		CreatedBy:      v.CreatedBy,
		Created:        v.Created,
		Updated:        v.Updated,
		RepoID:         v.RepoID,
		Type:           v.Type,
		LatestReviewID: v.LatestReviewID.Ptr(),
		ReviewDecision: v.ReviewDecision,
		SHA:            v.SHA,
	}
	return m
}

func mapInternalPullReqReviewer(v *types.PullReqReviewer) *pullReqReviewer {
	m := &pullReqReviewer{
		PullReqID:      v.PullReqID,
		PrincipalID:    v.PrincipalID,
		CreatedBy:      v.CreatedBy,
		Created:        v.Created,
		Updated:        v.Updated,
		RepoID:         v.RepoID,
		Type:           v.Type,
		LatestReviewID: null.IntFromPtr(v.LatestReviewID),
		ReviewDecision: v.ReviewDecision,
		SHA:            v.SHA,
	}
	return m
}

func (s *ReviewerOrmStore) mapPullReqReviewer(ctx context.Context, v *pullReqReviewer) *types.PullReqReviewer {
	m := &types.PullReqReviewer{
		PullReqID:      v.PullReqID,
		PrincipalID:    v.PrincipalID,
		CreatedBy:      v.CreatedBy,
		Created:        v.Created,
		Updated:        v.Updated,
		RepoID:         v.RepoID,
		Type:           v.Type,
		LatestReviewID: v.LatestReviewID.Ptr(),
		ReviewDecision: v.ReviewDecision,
		SHA:            v.SHA,
	}

	addedBy, err := s.pCache.Get(ctx, v.CreatedBy)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to load PR reviewer addedBy")
	}
	if addedBy != nil {
		m.AddedBy = *addedBy
	}

	reviewer, err := s.pCache.Get(ctx, v.PrincipalID)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to load PR reviewer principal")
	}
	if reviewer != nil {
		m.Reviewer = *reviewer
	}

	return m
}

func (s *ReviewerOrmStore) mapSlicePullReqReviewer(ctx context.Context,
	reviewers []*pullReqReviewer) ([]*types.PullReqReviewer, error) {
	// collect all principal IDs
	ids := make([]int64, 0, 2*len(reviewers))
	for _, v := range reviewers {
		ids = append(ids, v.CreatedBy)
		ids = append(ids, v.PrincipalID)
	}

	// pull principal infos from cache
	infoMap, err := s.pCache.Map(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load PR principal infos: %w", err)
	}

	// attach the principal infos back to the slice items
	m := make([]*types.PullReqReviewer, len(reviewers))
	for i, v := range reviewers {
		m[i] = mapPullReqReviewer(v)
		if addedBy, ok := infoMap[v.CreatedBy]; ok {
			m[i].AddedBy = *addedBy
		}
		if reviewer, ok := infoMap[v.PrincipalID]; ok {
			m[i].Reviewer = *reviewer
		}
	}

	return m, nil
}
