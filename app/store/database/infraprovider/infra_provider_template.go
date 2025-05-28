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

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	infraProviderTemplateIDColumn = `iptemp_id`
	infraProviderTemplateColumns  = `
		iptemp_uid,
    	iptemp_infra_provider_config_id,
    	iptemp_description,
    	iptemp_space_id,
    	iptemp_data,
    	iptemp_created,
    	iptemp_updated,
    	iptemp_version
	`
	infraProviderTemplateSelectColumns = infraProviderTemplateIDColumn + `,
    	` + infraProviderTemplateColumns
	infraProviderTemplateTable = `infra_provider_templates`
)

var _ store.InfraProviderTemplateStore = (*infraProviderTemplateStore)(nil)

type infraProviderTemplateStore struct {
	db *gorm.DB
}

type infraProviderTemplate struct {
	ID                    int64  `gorm:"column:iptemp_id"`
	Identifier            string `gorm:"column:iptemp_uid"`
	InfraProviderConfigID int64  `gorm:"column:iptemp_infra_provider_config_id"`
	Description           string `gorm:"column:iptemp_description"`
	SpaceID               int64  `gorm:"column:iptemp_space_id"`
	Data                  string `gorm:"column:iptemp_data"`
	Created               int64  `gorm:"column:iptemp_created"`
	Updated               int64  `gorm:"column:iptemp_updated"`
	Version               int64  `gorm:"column:iptemp_version"`
}

func NewInfraProviderTemplateStore(db *gorm.DB) store.InfraProviderTemplateStore {
	return &infraProviderTemplateStore{
		db: db,
	}
}

func (i infraProviderTemplateStore) FindByIdentifier(
	ctx context.Context,
	spaceID int64,
	identifier string,
) (*types.InfraProviderTemplate, error) {
	infraProviderTemplateEntity := new(infraProviderTemplate)
	if err := i.db.Table(infraProviderTemplateTable).
		Select(infraProviderTemplateSelectColumns).
		Where("iptemp_uid = ?", identifier).
		Where("iptemp_space_id = ?", spaceID).
		First(infraProviderTemplateEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "Failed to find infraprovider template %s", identifier)
	}
	return infraProviderTemplateEntity.mapToDTO(), nil
}

func (i infraProviderTemplateStore) Find(
	ctx context.Context,
	id int64,
) (*types.InfraProviderTemplate, error) {
	infraProviderTemplateEntity := new(infraProviderTemplate)
	if err := i.db.Table(infraProviderTemplateTable).
		Select(infraProviderTemplateSelectColumns).
		Where(infraProviderTemplateIDColumn+" = ?", id).
		First(infraProviderTemplateEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "Failed to find infraprovider template %d", id)
	}
	return infraProviderTemplateEntity.mapToDTO(), nil
}

func (i infraProviderTemplateStore) Create(
	ctx context.Context,
	infraProviderTemplatePtr *types.InfraProviderTemplate,
) error {
	infraProviderTemplateEntity := &infraProviderTemplate{
		Identifier:            infraProviderTemplatePtr.Identifier,
		InfraProviderConfigID: infraProviderTemplatePtr.InfraProviderConfigID,
		Description:           infraProviderTemplatePtr.Description,
		SpaceID:               infraProviderTemplatePtr.SpaceID,
		Data:                  infraProviderTemplatePtr.Data,
		Created:               infraProviderTemplatePtr.Created,
		Updated:               infraProviderTemplatePtr.Updated,
		Version:               infraProviderTemplatePtr.Version,
	}
	if err := i.db.Table(infraProviderTemplateTable).
		Create(infraProviderTemplateEntity).Error; err != nil {
		return errors.Wrapf(err, "infraprovider template create failed %s", infraProviderTemplatePtr.Identifier)
	}
	infraProviderTemplatePtr.ID = infraProviderTemplateEntity.ID
	return nil
}

func (i infraProviderTemplateStore) Update(
	ctx context.Context,
	infraProviderTemplatePtr *types.InfraProviderTemplate,
) error {
	dbinfraProviderTemplate := i.mapToInternalInfraProviderTemplate(infraProviderTemplatePtr)
	if err := i.db.Table(infraProviderTemplateTable).
		Where("iptemp_id = ?", infraProviderTemplatePtr.ID).
		Updates(infraProviderTemplate{
			Description: dbinfraProviderTemplate.Description,
			Updated:     dbinfraProviderTemplate.Updated,
			Data:        dbinfraProviderTemplate.Data,
			Version:     dbinfraProviderTemplate.Version + 1,
		}).Error; err != nil {
		return errors.Wrapf(err, "Failed to update infraprovider template %s", infraProviderTemplatePtr.Identifier)
	}
	return nil
}

func (i infraProviderTemplateStore) Delete(ctx context.Context, id int64) error {
	if err := i.db.Table(infraProviderTemplateTable).
		Where(infraProviderTemplateIDColumn+" = ?", id).
		Delete(nil).Error; err != nil {
		return errors.Wrap(err, "Failed to delete infraprovider template")
	}
	return nil
}

func (i infraProviderTemplateStore) mapToInternalInfraProviderTemplate(
	template *types.InfraProviderTemplate) infraProviderTemplate {
	return infraProviderTemplate{
		Identifier:            template.Identifier,
		InfraProviderConfigID: template.InfraProviderConfigID,
		Description:           template.Description,
		Data:                  template.Data,
		Version:               template.Version,
		SpaceID:               template.SpaceID,
		Created:               template.Created,
		Updated:               template.Updated,
	}
}

func (entity infraProviderTemplate) mapToDTO() *types.InfraProviderTemplate {
	return &types.InfraProviderTemplate{
		ID:                    entity.ID,
		Identifier:            entity.Identifier,
		InfraProviderConfigID: entity.InfraProviderConfigID,
		Description:           entity.Description,
		Data:                  entity.Data,
		Version:               entity.Version,
		SpaceID:               entity.SpaceID,
		Created:               entity.Created,
		Updated:               entity.Updated,
	}
}
