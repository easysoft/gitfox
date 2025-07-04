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

package usergroup

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

var _ store.UserGroupReviewersStore = (*UsergroupReviewerStore)(nil)

func NewUsergroupReviewerStore(
	db *gorm.DB,
	pCache store.PrincipalInfoCache,
	userGroupStore store.UserGroupStore,
) *UsergroupReviewerStore {
	return &UsergroupReviewerStore{
		db:             db,
		pInfoCache:     pCache,
		userGroupStore: userGroupStore,
	}
}

// UsergroupReviewerStore implements store.UsergroupReviewerStore backed by a relational database.
type UsergroupReviewerStore struct {
	db             *gorm.DB
	pInfoCache     store.PrincipalInfoCache
	userGroupStore store.UserGroupStore
}

type usergroupReviewer struct {
	PullReqID   int64 `gorm:"column:usergroup_reviewer_pullreq_id"`
	UserGroupID int64 `gorm:"column:usergroup_reviewer_usergroup_id"`
	CreatedBy   int64 `gorm:"column:usergroup_reviewer_created_by"`
	Created     int64 `gorm:"column:usergroup_reviewer_created"`
	Updated     int64 `gorm:"column:usergroup_reviewer_updated"`
	RepoID      int64 `gorm:"column:usergroup_reviewer_repo_id"`
}

const (
	pullreqUserGroupReviewerColumns = `
		 usergroup_reviewer_pullreq_id
		,usergroup_reviewer_usergroup_id
		,usergroup_reviewer_created_by
		,usergroup_reviewer_created
		,usergroup_reviewer_updated
		,usergroup_reviewer_repo_id`

	pullreqUserGroupReviewerSelectBase = `
	SELECT` + pullreqUserGroupReviewerColumns + `
	FROM usergroup_reviewers`
)

// Create creates a new pull request usergroup reviewer.
func (s *UsergroupReviewerStore) Create(ctx context.Context, v *types.UserGroupReviewer) error {
	reviewer := mapInternalPullReqUserGroupReviewer(v)
	if err := s.db.Create(&reviewer).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to create pull request usergroup reviewer")
	}
	return nil
}

// Delete deletes a pull request usergroup reviewer.
func (s *UsergroupReviewerStore) Delete(ctx context.Context, prID, userGroupReviewerID int64) error {
	reviewer := &usergroupReviewer{
		PullReqID:   prID,
		UserGroupID: userGroupReviewerID,
	}
	if err := s.db.Delete(reviewer).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to delete pull request usergroup reviewer")
	}
	return nil
}

// List returns a list of pull request usergroup reviewers.
func (s *UsergroupReviewerStore) List(ctx context.Context, prID int64) ([]*types.UserGroupReviewer, error) {
	const sqlQuery = pullreqUserGroupReviewerSelectBase + `
	WHERE usergroup_reviewer_pullreq_id = ?`

	var dst []*usergroupReviewer
	if err := s.db.Raw(sqlQuery, prID).Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list pull request usergroup reviewers")
	}

	result, err := s.mapSlicePullReqUserGroupReviewer(ctx, dst)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Find returns a pull request usergroup reviewer by userGroupReviewerID.
func (s *UsergroupReviewerStore) Find(
	ctx context.Context,
	prID,
	userGroupReviewerID int64,
) (*types.UserGroupReviewer, error) {
	const sqlQuery = pullreqUserGroupReviewerSelectBase + `
	WHERE usergroup_reviewer_pullreq_id = ? AND usergroup_reviewer_usergroup_id = ?`

	dst := &usergroupReviewer{}
	if err := s.db.Raw(sqlQuery, prID, userGroupReviewerID).Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pull request usergroup reviewer")
	}

	return mapPullReqUserGroupReviewer(dst), nil
}

func mapInternalPullReqUserGroupReviewer(v *types.UserGroupReviewer) *usergroupReviewer {
	m := &usergroupReviewer{
		PullReqID:   v.PullReqID,
		UserGroupID: v.UserGroupID,
		CreatedBy:   v.CreatedBy,
		Created:     v.Created,
		Updated:     v.Updated,
		RepoID:      v.RepoID,
	}
	return m
}

func mapPullReqUserGroupReviewer(v *usergroupReviewer) *types.UserGroupReviewer {
	m := &types.UserGroupReviewer{
		PullReqID:   v.PullReqID,
		UserGroupID: v.UserGroupID,
		CreatedBy:   v.CreatedBy,
		Created:     v.Created,
		Updated:     v.Updated,
		RepoID:      v.RepoID,
	}
	return m
}

func (s *UsergroupReviewerStore) mapSlicePullReqUserGroupReviewer(
	ctx context.Context,
	userGroupReviewers []*usergroupReviewer,
) ([]*types.UserGroupReviewer, error) {
	result := make([]*types.UserGroupReviewer, 0, len(userGroupReviewers))
	var addedByIDs []int64
	var userGroupIDs []int64
	for _, v := range userGroupReviewers {
		addedByIDs = append(addedByIDs, v.CreatedBy)
		userGroupIDs = append(userGroupIDs, v.UserGroupID)
	}

	// pull all the usergroups info
	userGroupsMap, err := s.userGroupStore.Map(ctx, userGroupIDs)
	if err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to load PR usergroups")
	}

	// pull principal infos from cache
	infoMap, err := s.pInfoCache.Map(ctx, addedByIDs)
	if err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to load PR principal infos")
	}

	for _, v := range userGroupReviewers {
		pullReqUsergroupReviewer := mapPullReqUserGroupReviewer(v)
		pullReqUsergroupReviewer.UserGroup = *userGroupsMap[v.UserGroupID].ToUserGroupInfo()
		pullReqUsergroupReviewer.AddedBy = *infoMap[v.CreatedBy]

		result = append(result, pullReqUsergroupReviewer)
	}
	return result, nil
}
