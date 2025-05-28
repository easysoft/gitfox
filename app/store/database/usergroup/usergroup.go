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

	gitfoxAppStore "github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ gitfoxAppStore.UserGroupStore = (*UserGroupStore)(nil)

func NewUserGroupStore(db *gorm.DB) *UserGroupStore {
	return &UserGroupStore{
		db: db,
	}
}

type UserGroupStore struct {
	db *gorm.DB
}

type UserGroup struct {
	SpaceID     int64  `gorm:"column:usergroup_space_id"`
	ID          int64  `gorm:"column:usergroup_id"`
	Identifier  string `gorm:"column:usergroup_identifier"`
	Name        string `gorm:"column:usergroup_name"`
	Description string `gorm:"column:usergroup_description"`
	Created     int64  `gorm:"column:usergroup_created"`
	Updated     int64  `gorm:"column:usergroup_updated"`
}

const (
	userGroupColumns = `
	usergroup_id
	,usergroup_identifier
	,usergroup_name
	,usergroup_description
	,usergroup_space_id
	,usergroup_created
	,usergroup_updated`

	userGroupSelectBase = `SELECT ` + userGroupColumns + ` FROM usergroups`
)

func mapUserGroup(ug *UserGroup) *types.UserGroup {
	return &types.UserGroup{
		ID:          ug.ID,
		Identifier:  ug.Identifier,
		Name:        ug.Name,
		Description: ug.Description,
		SpaceID:     ug.SpaceID,
		Created:     ug.Created,
		Updated:     ug.Updated,
	}
}

// FindByIdentifier returns a usergroup by its identifier.
func (s *UserGroupStore) FindByIdentifier(
	ctx context.Context,
	spaceID int64,
	identifier string,
) (*types.UserGroup, error) {
	dst := &UserGroup{}
	if err := s.db.Where("usergroup_identifier = ? AND usergroup_space_id = ?", identifier, spaceID).First(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find usergroup by identifier %s", identifier)
	}
	return mapUserGroup(dst), nil
}

// Find returns a usergroup by its id.
func (s *UserGroupStore) Find(ctx context.Context, id int64) (*types.UserGroup, error) {
	dst := &UserGroup{}
	if err := s.db.Where("usergroup_id = ?", id).First(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find usergroup by id %d", id)
	}
	return mapUserGroup(dst), nil
}

func (s *UserGroupStore) Map(ctx context.Context, ids []int64) (map[int64]*types.UserGroup, error) {
	result, err := s.FindManyByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, store.ErrResourceNotFound
	}
	mapResult := make(map[int64]*types.UserGroup, len(result))
	for _, r := range result {
		mapResult[r.ID] = r
	}
	return mapResult, nil
}

func (s *UserGroupStore) FindManyByIDs(ctx context.Context, ids []int64) ([]*types.UserGroup, error) {
	var dst []*UserGroup
	if err := s.db.Where("usergroup_id IN ?", ids).Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to find many usergroups by ids")
	}

	result := make([]*types.UserGroup, len(dst))
	for i, u := range dst {
		result[i] = u.toUserGroupType()
	}

	return result, nil
}

func (s *UserGroupStore) FindManyByIdentifiersAndSpaceID(
	ctx context.Context,
	identifiers []string,
	spaceID int64,
) ([]*types.UserGroup, error) {
	var dst []*UserGroup
	if err := s.db.Where("usergroup_identifier IN ?", identifiers).Where("usergroup_space_id = ?", spaceID).Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to find many usergroups by identifiers")
	}
	result := make([]*types.UserGroup, len(dst))
	for i, u := range dst {
		result[i] = mapUserGroup(u)
	}
	return result, nil
}

// Create Creates a usergroup in the database.
func (s *UserGroupStore) Create(
	ctx context.Context,
	spaceID int64,
	userGroup *types.UserGroup,
) error {
	dst := mapInternalUserGroup(userGroup, spaceID)
	if err := s.db.Create(dst).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "failed to create usergroup")
	}
	return nil
}

func (s *UserGroupStore) CreateOrUpdate(
	ctx context.Context,
	spaceID int64,
	userGroup *types.UserGroup,
) error {
	dst := mapInternalUserGroup(userGroup, spaceID)
	err := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "usergroup_identifier"}, {Name: "usergroup_space_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"usergroup_name",
			"usergroup_description",
			"usergroup_updated",
		}),
	}).Create(dst).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to create or update usergroup")
	}
	return nil
}

func mapInternalUserGroup(u *types.UserGroup, spaceID int64) *UserGroup {
	return &UserGroup{
		ID:          u.ID,
		SpaceID:     spaceID,
		Identifier:  u.Identifier,
		Name:        u.Name,
		Description: u.Description,
		Created:     u.Created,
		Updated:     u.Updated,
	}
}

func (u *UserGroup) toUserGroupType() *types.UserGroup {
	return &types.UserGroup{
		ID:          u.ID,
		Identifier:  u.Identifier,
		Name:        u.Name,
		Description: u.Description,
		Created:     u.Created,
		Updated:     u.Updated,
	}
}
