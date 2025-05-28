// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pullreq

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.PullReqReviewStore = (*ReviewOrmStore)(nil)

// NewPullReqReviewOrmStore returns a new PullReqReviewOrmStore.
func NewPullReqReviewOrmStore(db *gorm.DB) *ReviewOrmStore {
	return &ReviewOrmStore{
		db: db,
	}
}

// ReviewOrmStore implements store.PullReqReviewStore backed by a relational database.
type ReviewOrmStore struct {
	db *gorm.DB
}

// pullReqReview is used to fetch pull request review data from the database.
type pullReqReview struct {
	ID int64 `gorm:"column:pullreq_review_id;primaryKey"`

	CreatedBy int64 `gorm:"column:pullreq_review_created_by"`
	Created   int64 `gorm:"column:pullreq_review_created"`
	Updated   int64 `gorm:"column:pullreq_review_updated"`

	PullReqID int64 `gorm:"column:pullreq_review_pullreq_id"`

	Decision enum.PullReqReviewDecision `gorm:"column:pullreq_review_decision"`
	SHA      string                     `gorm:"column:pullreq_review_sha"`
}

const (
	tablePullReview = "pullreq_reviews"
)

// Find finds the pull request activity by id.
func (s *ReviewOrmStore) Find(ctx context.Context, id int64) (*types.PullReqReview, error) {
	dst := &pullReqReview{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReview).First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pull request activity")
	}

	return mapPullReqReview(dst), nil
}

// Create creates a new pull request.
func (s *ReviewOrmStore) Create(ctx context.Context, v *types.PullReqReview) error {
	dbObj := mapInternalPullReqReview(v)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReview).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to insert pull request review")
	}

	v.ID = dbObj.ID
	return nil
}

func mapPullReqReview(v *pullReqReview) *types.PullReqReview {
	return (*types.PullReqReview)(v) // the two types are identical, except for the tags
}

func mapInternalPullReqReview(v *types.PullReqReview) *pullReqReview {
	return (*pullReqReview)(v) // the two types are identical, except for the tags
}
