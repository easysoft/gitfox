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
	"encoding/json"
	"fmt"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	infraProviderResourceIDColumn      = `ipreso_id`
	infraProviderResourceInsertColumns = `
		ipreso_uid,
		ipreso_display_name,
		ipreso_infra_provider_config_id,
		ipreso_type,
		ipreso_space_id,
		ipreso_created,
		ipreso_updated,
		ipreso_cpu,
		ipreso_memory,
		ipreso_disk,
		ipreso_network,
		ipreso_region,
		ipreso_opentofu_params,
		ipreso_infra_provider_template_id
	`
	infraProviderResourceSelectColumns = "ipreso_id," + infraProviderResourceInsertColumns
	infraProviderResourceTable         = `infra_provider_resources`
)

type infraProviderResource struct {
	ID                    int64                  `gorm:"column:ipreso_id"`
	Identifier            string                 `gorm:"column:ipreso_uid"`
	Name                  string                 `gorm:"column:ipreso_display_name"`
	InfraProviderConfigID int64                  `gorm:"column:ipreso_infra_provider_config_id"`
	InfraProviderType     enum.InfraProviderType `gorm:"column:ipreso_type"`
	SpaceID               int64                  `gorm:"column:ipreso_space_id"`
	CPU                   null.String            `gorm:"column:ipreso_cpu"`
	Memory                null.String            `gorm:"column:ipreso_memory"`
	Disk                  null.String            `gorm:"column:ipreso_disk"`
	Network               null.String            `gorm:"column:ipreso_network"`
	Region                string                 `gorm:"column:ipreso_region"` // need list maybe
	OpenTofuParams        []byte                 `gorm:"column:ipreso_opentofu_params"`
	TemplateID            null.Int               `gorm:"column:ipreso_infra_provider_template_id"`
	Created               int64                  `gorm:"column:ipreso_created"`
	Updated               int64                  `gorm:"column:ipreso_updated"`
}

var _ store.InfraProviderResourceStore = (*infraProviderResourceStore)(nil)

// NewGitspaceConfigStore returns a new GitspaceConfigStore.
func NewInfraProviderResourceStore(db *gorm.DB) store.InfraProviderResourceStore {
	return &infraProviderResourceStore{
		db: db,
	}
}

type infraProviderResourceStore struct {
	db *gorm.DB
}

func (s infraProviderResourceStore) List(ctx context.Context, infraProviderConfigID int64,
	_ types.ListQueryFilter) ([]*types.InfraProviderResource, error) {
	var resources []infraProviderResource
	if err := s.db.Table(infraProviderResourceTable).
		Select(infraProviderResourceSelectColumns).
		Where("ipreso_infra_provider_config_id = ?", infraProviderConfigID).
		Find(&resources).Error; err != nil {
		return nil, err
	}
	return mapToInfraProviderResources(ctx, resources)
}

func (s infraProviderResourceStore) Find(ctx context.Context, id int64) (*types.InfraProviderResource, error) {
	var resource infraProviderResource
	if err := s.db.Table(infraProviderResourceTable).
		Select(infraProviderResourceSelectColumns).
		Where(infraProviderResourceIDColumn+" = ?", id).
		First(&resource).Error; err != nil {
		return nil, errors.Wrapf(err, "Failed to find infra provider resource %d", id)
	}
	return s.mapToInfraProviderResource(ctx, &resource)
}

func (s infraProviderResourceStore) FindByIdentifier(
	ctx context.Context,
	spaceID int64,
	identifier string,
) (*types.InfraProviderResource, error) {
	var resource infraProviderResource
	if err := s.db.Table(infraProviderResourceTable).
		Select(infraProviderResourceSelectColumns).
		Where("ipreso_uid = ?", identifier).
		Where("ipreso_space_id = ?", spaceID).
		First(&resource).Error; err != nil {
		return nil, errors.Wrapf(err, "Failed to find infra provider resource %s", identifier)
	}
	return s.mapToInfraProviderResource(ctx, &resource)
}

func (s infraProviderResourceStore) Update(
	ctx context.Context,
	infraProviderResource *types.InfraProviderResource,
) error {
	dbinfraProviderResource, err := s.mapToInternalInfraProviderResource(ctx, infraProviderResource)
	if err != nil {
		return fmt.Errorf(
			"failed to map to DB Obj for infraprovider resource %s", infraProviderResource.UID)
	}
	result := s.db.Table(infraProviderResourceTable).
		Where("ipreso_id = ?", infraProviderResource.ID).
		Updates(map[string]interface{}{
			"ipreso_display_name":    dbinfraProviderResource.Name,
			"ipreso_updated":         dbinfraProviderResource.Updated,
			"ipreso_memory":          dbinfraProviderResource.Memory,
			"ipreso_disk":            dbinfraProviderResource.Disk,
			"ipreso_network":         dbinfraProviderResource.Network,
			"ipreso_region":          dbinfraProviderResource.Region,
			"ipreso_opentofu_params": dbinfraProviderResource.OpenTofuParams,
		})
	if result.Error != nil {
		return errors.Wrap(result.Error, "Failed to update infraprovider resource")
	}
	return nil
}

func (s infraProviderResourceStore) Create(
	ctx context.Context,
	infraProviderResourcePtr *types.InfraProviderResource,
) error {
	jsonBytes, marshalErr := json.Marshal(infraProviderResourcePtr.Metadata)
	if marshalErr != nil {
		return marshalErr
	}
	resource := infraProviderResource{
		Identifier:            infraProviderResourcePtr.UID,
		Name:                  infraProviderResourcePtr.Name,
		InfraProviderConfigID: infraProviderResourcePtr.InfraProviderConfigID,
		InfraProviderType:     infraProviderResourcePtr.InfraProviderType,
		SpaceID:               infraProviderResourcePtr.SpaceID,
		Created:               infraProviderResourcePtr.Created,
		Updated:               infraProviderResourcePtr.Updated,
		CPU:                   null.StringFromPtr(infraProviderResourcePtr.CPU),
		Memory:                null.StringFromPtr(infraProviderResourcePtr.Memory),
		Disk:                  null.StringFromPtr(infraProviderResourcePtr.Disk),
		Network:               null.StringFromPtr(infraProviderResourcePtr.Network),
		Region:                infraProviderResourcePtr.Region,
		OpenTofuParams:        jsonBytes,
		TemplateID:            null.IntFromPtr(infraProviderResourcePtr.TemplateID),
	}
	if err := s.db.Table(infraProviderResourceTable).Create(&resource).Error; err != nil {
		return err
	}
	infraProviderResourcePtr.ID = resource.ID
	return nil
}

func (s infraProviderResourceStore) DeleteByIdentifier(ctx context.Context, spaceID int64, identifier string) error {
	if err := s.db.Table(infraProviderResourceTable).
		Where("ipreso_uid = ?", identifier).
		Where("ipreso_space_id = ?", spaceID).
		Delete(&infraProviderResource{}).Error; err != nil {
		return errors.Wrapf(err, "Failed to delete infra provider resource %s", identifier)
	}
	return nil
}

func (s infraProviderResourceStore) mapToInfraProviderResource(_ context.Context,
	in *infraProviderResource) (*types.InfraProviderResource, error) {
	openTofuParamsMap := make(map[string]string)
	marshalErr := json.Unmarshal(in.OpenTofuParams, &openTofuParamsMap)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return &types.InfraProviderResource{
		UID:                   in.Identifier,
		InfraProviderConfigID: in.InfraProviderConfigID,
		ID:                    in.ID,
		InfraProviderType:     in.InfraProviderType,
		Name:                  in.Name,
		SpaceID:               in.SpaceID,
		CPU:                   in.CPU.Ptr(),
		Memory:                in.Memory.Ptr(),
		Disk:                  in.Disk.Ptr(),
		Network:               in.Network.Ptr(),
		Region:                in.Region,
		Metadata:              openTofuParamsMap,
		TemplateID:            in.TemplateID.Ptr(),
		Created:               in.Created,
		Updated:               in.Updated,
	}, nil
}

func mapToInfraProviderResource(_ context.Context,
	in *infraProviderResource) (*types.InfraProviderResource, error) {
	openTofuParamsMap := make(map[string]string)
	marshalErr := json.Unmarshal(in.OpenTofuParams, &openTofuParamsMap)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return &types.InfraProviderResource{
		UID:                   in.Identifier,
		InfraProviderConfigID: in.InfraProviderConfigID,
		ID:                    in.ID,
		InfraProviderType:     in.InfraProviderType,
		Name:                  in.Name,
		SpaceID:               in.SpaceID,
		CPU:                   in.CPU.Ptr(),
		Memory:                in.Memory.Ptr(),
		Disk:                  in.Disk.Ptr(),
		Network:               in.Network.Ptr(),
		Region:                in.Region,
		Metadata:              openTofuParamsMap,
		TemplateID:            in.TemplateID.Ptr(),
		Created:               in.Created,
		Updated:               in.Updated,
	}, nil
}

func (s infraProviderResourceStore) mapToInternalInfraProviderResource(_ context.Context,
	in *types.InfraProviderResource) (*infraProviderResource, error) {
	jsonBytes, marshalErr := json.Marshal(in.Metadata)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return &infraProviderResource{
		Identifier:            in.UID,
		InfraProviderConfigID: in.InfraProviderConfigID,
		InfraProviderType:     in.InfraProviderType,
		Name:                  in.Name,
		SpaceID:               in.SpaceID,
		CPU:                   null.StringFromPtr(in.CPU),
		Memory:                null.StringFromPtr(in.Memory),
		Disk:                  null.StringFromPtr(in.Disk),
		Network:               null.StringFromPtr(in.Network),
		Region:                in.Region,
		OpenTofuParams:        jsonBytes,
		TemplateID:            null.IntFromPtr(in.TemplateID),
		Created:               in.Created,
		Updated:               in.Updated,
	}, nil
}

func mapToInfraProviderResources(ctx context.Context,
	resources []infraProviderResource) ([]*types.InfraProviderResource, error) {
	var err error
	res := make([]*types.InfraProviderResource, len(resources))
	for i := range resources {
		res[i], err = mapToInfraProviderResource(ctx, &resources[i])
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

var _ store.InfraProviderResourceView = (*InfraProviderResourceView)(nil)

// NewInfraProviderResourceView returns a new InfraProviderResourceView.
// It's used by the infraprovider resource cache.
func NewInfraProviderResourceView(db *gorm.DB) *InfraProviderResourceView {
	return &InfraProviderResourceView{
		db: db,
	}
}

type InfraProviderResourceView struct {
	db *gorm.DB
}

func (i InfraProviderResourceView) Find(ctx context.Context, id int64) (*types.InfraProviderResource, error) {
	dst := new(infraProviderResource)
	if err := i.db.Table(infraProviderResourceTable).Where(infraProviderResourceIDColumn+" = ?", id).First(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find infraprovider resource %d", id)
	}
	return mapToInfraProviderResource(ctx, dst)
}

func (i InfraProviderResourceView) FindMany(ctx context.Context, ids []int64) ([]*types.InfraProviderResource, error) {
	var resources []*types.InfraProviderResource
	if err := i.db.Table(infraProviderResourceTable).
		Select(infraProviderResourceSelectColumns).
		Where(infraProviderResourceIDColumn+" IN ?", ids).
		Find(&resources).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find infraprovider resources")
	}
	return resources, nil
}
