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

package gitspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type gitspaceConfig struct {
	ID                      int64                     `gorm:"column:gconf_id"`
	Identifier              string                    `gorm:"column:gconf_uid"`
	Name                    string                    `gorm:"column:gconf_display_name"`
	IDE                     enum.IDEType              `gorm:"column:gconf_ide"`
	InfraProviderResourceID int64                     `gorm:"column:gconf_infra_provider_resource_id"`
	CodeAuthType            string                    `gorm:"column:gconf_code_auth_type"`
	CodeRepoRef             null.String               `gorm:"column:gconf_code_repo_ref"`
	CodeAuthID              string                    `gorm:"column:gconf_code_auth_id"`
	CodeRepoType            enum.GitspaceCodeRepoType `gorm:"column:gconf_code_repo_type"`
	CodeRepoIsPrivate       bool                      `gorm:"column:gconf_code_repo_is_private"`
	CodeRepoURL             string                    `gorm:"column:gconf_code_repo_url"`
	DevcontainerPath        null.String               `gorm:"column:gconf_devcontainer_path"`
	Branch                  string                    `gorm:"column:gconf_branch"`
	// TODO: migrate to principal int64 id to use principal cache and consistent with gitfox code.
	UserUID            string   `gorm:"column:gconf_user_uid"`
	SpaceID            int64    `gorm:"column:gconf_space_id"`
	Created            int64    `gorm:"column:gconf_created"`
	Updated            int64    `gorm:"column:gconf_updated"`
	IsDeleted          bool     `gorm:"column:gconf_is_deleted"`
	SSHTokenIdentifier string   `gorm:"column:gconf_ssh_token_identifier"`
	CreatedBy          null.Int `gorm:"column:gconf_created_by"`
}

const (
	gitspaceConfigsTable = `gitspace_configs`
)

var _ store.GitspaceConfigStore = (*gitspaceConfigStore)(nil)

// NewGitspaceConfigStore returns a new GitspaceConfigStore.
func NewGitspaceConfigStore(db *gorm.DB, pCache store.PrincipalInfoCache, rCache store.InfraProviderResourceCache) store.GitspaceConfigStore {
	return &gitspaceConfigStore{
		db:     db,
		pCache: pCache,
		rCache: rCache,
	}
}

type gitspaceConfigStore struct {
	db     *gorm.DB
	pCache store.PrincipalInfoCache
	rCache store.InfraProviderResourceCache
}

func (s gitspaceConfigStore) Count(ctx context.Context, filter *types.GitspaceFilter) (int64, error) {
	var count int64
	countStmt := s.db.Table(gitspaceConfigsTable)
	if !filter.IncludeDeleted {
		countStmt = countStmt.Where("gconf_is_deleted = ?", false)
	}
	if filter.UserID != "" {
		countStmt = countStmt.Where("gconf_user_uid = ?", filter.UserID)
	}
	if filter.SpaceIDs != nil {
		countStmt = countStmt.Where("gconf_space_id IN ?", filter.SpaceIDs)
	}

	err := countStmt.Count(&count).Error
	if err != nil {
		return 0, errors.Wrap(err, "Failed executing custom count query")
	}
	return count, nil
}

func (s gitspaceConfigStore) Find(ctx context.Context, id int64) (*types.GitspaceConfig, error) {
	dst := new(gitspaceConfig)
	err := s.db.
		Table(gitspaceConfigsTable).
		Where("gconf_id = ?", id).
		Where("gconf_is_deleted = ?", false).
		Take(dst).Error
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find gitspace config")
	}
	return s.mapToGitspaceConfig(ctx, dst)
}

func (s gitspaceConfigStore) FindByIdentifier(
	ctx context.Context,
	spaceID int64,
	identifier string,
) (*types.GitspaceConfig, error) {
	dst := new(gitspaceConfig)
	err := s.db.
		Table(gitspaceConfigsTable).
		Where("LOWER(gconf_uid) = ?", strings.ToLower(identifier)).
		Where("gconf_space_id = ?", spaceID).
		Where("gconf_is_deleted = ?", false).
		Take(dst).Error
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find gitspace config")
	}
	return s.mapToGitspaceConfig(ctx, dst)
}

func (s gitspaceConfigStore) Create(ctx context.Context, gitspaceConfig *types.GitspaceConfig) error {
	dbGitspaceConfig := mapToInternalGitspaceConfig(gitspaceConfig)
	err := s.db.
		Table(gitspaceConfigsTable).
		Create(dbGitspaceConfig).Error
	if err != nil {
		return errors.Wrap(err, "Failed to create gitspace config")
	}
	return nil
}

func (s gitspaceConfigStore) Update(ctx context.Context, gitspaceConfig *types.GitspaceConfig) error {
	dbGitspaceConfig := mapToInternalGitspaceConfig(gitspaceConfig)
	err := s.db.
		Table(gitspaceConfigsTable).
		Where("gconf_id = ?", gitspaceConfig.ID).
		Updates(dbGitspaceConfig).Error
	if err != nil {
		return errors.Wrap(err, "Failed to update gitspace config")
	}
	return nil
}

func mapToInternalGitspaceConfig(config *types.GitspaceConfig) *gitspaceConfig {
	return &gitspaceConfig{
		ID:                      config.ID,
		Identifier:              config.Identifier,
		Name:                    config.Name,
		IDE:                     config.IDE,
		InfraProviderResourceID: config.InfraProviderResource.ID,
		CodeAuthType:            config.CodeRepo.AuthType,
		CodeAuthID:              config.CodeRepo.AuthID,
		CodeRepoIsPrivate:       config.CodeRepo.IsPrivate,
		CodeRepoType:            config.CodeRepo.Type,
		CodeRepoRef:             null.StringFromPtr(config.CodeRepo.Ref),
		CodeRepoURL:             config.CodeRepo.URL,
		DevcontainerPath:        null.StringFromPtr(config.DevcontainerPath),
		Branch:                  config.Branch,
		UserUID:                 config.GitspaceUser.Identifier,
		SpaceID:                 config.SpaceID,
		IsDeleted:               config.IsDeleted,
		Created:                 config.Created,
		Updated:                 config.Updated,
		SSHTokenIdentifier:      config.SSHTokenIdentifier,
		CreatedBy:               null.IntFromPtr(config.GitspaceUser.ID),
	}
}

func (s gitspaceConfigStore) List(ctx context.Context, filter *types.GitspaceFilter) ([]*types.GitspaceConfig, error) {
	var dst []*gitspaceConfig
	err := s.db.
		Table(gitspaceConfigsTable).
		Where("gconf_is_deleted = ?", false).
		Where("gconf_user_uid = ?", filter.UserID).
		Where("gconf_space_id IN ?", filter.SpaceIDs).
		Limit(int(filter.QueryFilter.Size)).
		Offset(int(filter.QueryFilter.Page * filter.QueryFilter.Size)).
		Find(&dst).Error
	if err != nil {
		return nil, errors.Wrap(err, "Failed executing custom list query")
	}
	return s.mapToGitspaceConfigs(ctx, dst)
}

func (s *gitspaceConfigStore) mapToGitspaceConfig(
	ctx context.Context,
	in *gitspaceConfig,
) (*types.GitspaceConfig, error) {
	codeRepo := types.CodeRepo{
		URL:              in.CodeRepoURL,
		Ref:              in.CodeRepoRef.Ptr(),
		Type:             in.CodeRepoType,
		Branch:           in.Branch,
		DevcontainerPath: in.DevcontainerPath.Ptr(),
		IsPrivate:        in.CodeRepoIsPrivate,
		AuthType:         in.CodeAuthType,
		AuthID:           in.CodeAuthID,
	}
	var res = &types.GitspaceConfig{
		ID:                 in.ID,
		Identifier:         in.Identifier,
		Name:               in.Name,
		IDE:                in.IDE,
		SpaceID:            in.SpaceID,
		Created:            in.Created,
		Updated:            in.Updated,
		SSHTokenIdentifier: in.SSHTokenIdentifier,
		CodeRepo:           codeRepo,
		GitspaceUser: types.GitspaceUser{
			ID:         in.CreatedBy.Ptr(),
			Identifier: in.UserUID},
	}
	if res.GitspaceUser.ID != nil {
		author, _ := s.pCache.Get(ctx, *res.GitspaceUser.ID)
		if author != nil {
			res.GitspaceUser.DisplayName = author.DisplayName
			res.GitspaceUser.Email = author.Email
		}
	}
	if resource, err := s.rCache.Get(ctx, in.InfraProviderResourceID); err == nil {
		res.InfraProviderResource = *resource
	} else {
		return nil, fmt.Errorf("couldn't set resource to the config in DB: %s", in.Identifier)
	}
	return res, nil
}

func (s *gitspaceConfigStore) mapToGitspaceConfigs(ctx context.Context,
	configs []*gitspaceConfig) ([]*types.GitspaceConfig, error) {
	var err error
	res := make([]*types.GitspaceConfig, len(configs))
	for i := range configs {
		res[i], err = s.mapToGitspaceConfig(ctx, configs[i])
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (s gitspaceConfigStore) ListAll(
	ctx context.Context,
	userUID string,
) ([]*types.GitspaceConfig, error) {
	var dst []*gitspaceConfig
	err := s.db.
		Table(gitspaceConfigsTable).
		Where("gconf_is_deleted = ?", false).
		Where("gconf_user_uid = ?", userUID).
		Find(&dst).Error
	if err != nil {
		return nil, errors.Wrap(err, "Failed executing custom list query")
	}
	return s.mapToGitspaceConfigs(ctx, dst)
}
