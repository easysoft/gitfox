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

package infraprovider

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	infraProviderConfigIDColumn      = `ipconf_id`
	infraProviderConfigInsertColumns = `
		ipconf_uid,
		ipconf_display_name,
		ipconf_type,
		ipconf_space_id,
		ipconf_created,
		ipconf_updated
	`
	infraProviderConfigSelectColumns = "ipconf_id," + infraProviderConfigInsertColumns
	infraProviderConfigTable         = `infra_provider_configs`
)

type infraProviderConfig struct {
	ID         int64                  `gorm:"column:ipconf_id"`
	Identifier string                 `gorm:"column:ipconf_uid"`
	Name       string                 `gorm:"column:ipconf_display_name"`
	Type       enum.InfraProviderType `gorm:"column:ipconf_type"`
	SpaceID    int64                  `gorm:"column:ipconf_space_id"`
	Created    int64                  `gorm:"column:ipconf_created"`
	Updated    int64                  `gorm:"column:ipconf_updated"`
}

var _ store.InfraProviderConfigStore = (*infraProviderConfigStore)(nil)

// NewGitspaceConfigStore returns a new GitspaceConfigStore.
func NewInfraProviderConfigStore(db *gorm.DB) store.InfraProviderConfigStore {
	return &infraProviderConfigStore{
		db: db,
	}
}

type infraProviderConfigStore struct {
	db *gorm.DB
}

func (i infraProviderConfigStore) Update(ctx context.Context, infraProviderConfig *types.InfraProviderConfig) error {
	dbinfraProviderConfig := i.mapToInternalInfraProviderConfig(ctx, infraProviderConfig)
	if err := i.db.Table(infraProviderConfigTable).
		Where("ipconf_id = ?", infraProviderConfig.ID).
		Updates(dbinfraProviderConfig).Error; err != nil {
		return errors.Wrapf(err, "Failed to update infra provider config %s", infraProviderConfig.Identifier)
	}
	return nil
}

func (i infraProviderConfigStore) Find(ctx context.Context, id int64) (*types.InfraProviderConfig, error) {
	dst := new(infraProviderConfig)
	if err := i.db.Table(infraProviderConfigTable).Select(infraProviderConfigSelectColumns).Where(infraProviderConfigIDColumn+" = ?", id).Scan(dst).Error; err != nil {
		return nil, errors.Wrapf(err, "Failed to find infraprovider config %d", id)
	}
	return i.mapToInfraProviderConfig(ctx, dst), nil
}

func (i infraProviderConfigStore) FindByIdentifier(
	ctx context.Context,
	spaceID int64,
	identifier string,
) (*types.InfraProviderConfig, error) {
	dst := new(infraProviderConfig)
	if err := i.db.Table(infraProviderConfigTable).
		Select(infraProviderConfigSelectColumns).
		Where("ipconf_uid = ?", identifier).
		Where("ipconf_space_id = ?", spaceID).
		Scan(dst).Error; err != nil {
		return nil, errors.Wrapf(err, "Failed to find infraprovider config %s", identifier)
	}
	return i.mapToInfraProviderConfig(ctx, dst), nil
}

func (i infraProviderConfigStore) Create(ctx context.Context, infraProviderConfig *types.InfraProviderConfig) error {
	if err := i.db.Table(infraProviderConfigTable).
		Create(&infraProviderConfig).
		Error; err != nil {
		return errors.Wrapf(err, "infraprovider config create query failed for %s", infraProviderConfig.Identifier)
	}
	return nil
}

func (i infraProviderConfigStore) mapToInfraProviderConfig(
	_ context.Context,
	in *infraProviderConfig) *types.InfraProviderConfig {
	infraProviderConfigEntity := &types.InfraProviderConfig{
		ID:         in.ID,
		Identifier: in.Identifier,
		Name:       in.Name,
		Type:       in.Type,
		SpaceID:    in.SpaceID,
		Created:    in.Created,
		Updated:    in.Updated,
	}
	return infraProviderConfigEntity
}

func (i infraProviderConfigStore) mapToInternalInfraProviderConfig(
	_ context.Context,
	in *types.InfraProviderConfig) *infraProviderConfig {
	infraProviderConfigEntity := &infraProviderConfig{
		Identifier: in.Identifier,
		Name:       in.Name,
		Type:       in.Type,
		SpaceID:    in.SpaceID,
		Created:    in.Created,
		Updated:    in.Updated,
	}
	return infraProviderConfigEntity
}
