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
	"fmt"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	infraProvisionedIDColumn = `iprov_id`
	infraProvisionedColumns  = `
		iprov_gitspace_id,
		iprov_type,
		iprov_infra_provider_resource_id,
		iprov_space_id,
		iprov_created,
		iprov_updated,
		iprov_response_metadata,
		iprov_opentofu_params,
		iprov_infra_status,
		iprov_server_host_ip,
		iprov_server_host_port,
		iprov_proxy_host,
		iprov_proxy_port
	`
	infraProvisionedSelectColumns = infraProvisionedIDColumn + `,
		` + infraProvisionedColumns
	infraProvisionedTable = `infra_provisioned`
	gitspaceInstanceTable = `gitspaces`
)

var _ store.InfraProvisionedStore = (*infraProvisionedStore)(nil)

type infraProvisionedStore struct {
	db *gorm.DB
}

type infraProvisioned struct {
	ID                      int64                  `gorm:"column:iprov_id"`
	GitspaceInstanceID      int64                  `gorm:"column:iprov_gitspace_id"`
	InfraProviderType       enum.InfraProviderType `gorm:"column:iprov_type"`
	InfraProviderResourceID int64                  `gorm:"column:iprov_infra_provider_resource_id"`
	SpaceID                 int64                  `gorm:"column:iprov_space_id"`
	Created                 int64                  `gorm:"column:iprov_created"`
	Updated                 int64                  `gorm:"column:iprov_updated"`
	ResponseMetadata        *string                `gorm:"column:iprov_response_metadata"`
	InputParams             string                 `gorm:"column:iprov_opentofu_params"`
	InfraStatus             enum.InfraStatus       `gorm:"column:iprov_infra_status"`
	ServerHostIP            string                 `gorm:"column:iprov_server_host_ip"`
	ServerHostPort          string                 `gorm:"column:iprov_server_host_port"`
	ProxyHost               string                 `gorm:"column:iprov_proxy_host"`
	ProxyPort               int32                  `gorm:"column:iprov_proxy_port"`
}

type infraProvisionedGatewayView struct {
	GitspaceInstanceIdentifier string  `gorm:"column:iprov_gitspace_uid"`
	SpaceID                    int64   `gorm:"column:iprov_space_id"`
	ServerHostIP               string  `gorm:"column:iprov_server_host_ip"`
	ServerHostPort             string  `gorm:"column:iprov_server_host_port"`
	Infrastructure             *string `gorm:"column:iprov_response_metadata"`
}

func NewInfraProvisionedStore(db *gorm.DB) store.InfraProvisionedStore {
	return &infraProvisionedStore{
		db: db,
	}
}

func (i infraProvisionedStore) Find(ctx context.Context, id int64) (*types.InfraProvisioned, error) {
	var entity infraProvisioned
	err := i.db.Table(infraProvisionedTable).
		Select(infraProvisionedSelectColumns).
		Where(infraProvisionedIDColumn+" = ?", id).
		First(&entity).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to find infraprovisioned for %d", id)
	}
	return entity.toDTO(), nil
}

func (i infraProvisionedStore) FindAllLatestByGateway(
	ctx context.Context,
	gatewayHost string,
) ([]*types.InfraProvisionedGatewayView, error) {
	var entities []*infraProvisionedGatewayView
	err := i.db.Table(infraProvisionedTable).
		Select("gits_uid as iprov_gitspace_uid, iprov_space_id, iprov_server_host_ip, iprov_server_host_port, iprov_response_metadata").
		Joins(fmt.Sprintf("JOIN %s ON iprov_gitspace_id = gits_id", gitspaceInstanceTable)).
		Where("iprov_proxy_host = ?", gatewayHost).
		Where("iprov_infra_status = ?", enum.InfraStatusProvisioned).
		Order("iprov_created DESC").
		Scan(&entities).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to find infraprovisioned for host %s", gatewayHost)
	}

	var result = make([]*types.InfraProvisionedGatewayView, len(entities))
	for index, entity := range entities {
		result[index] = &types.InfraProvisionedGatewayView{
			GitspaceInstanceIdentifier: entity.GitspaceInstanceIdentifier,
			SpaceID:                    entity.SpaceID,
			ServerHostIP:               entity.ServerHostIP,
			ServerHostPort:             entity.ServerHostPort,
			Infrastructure:             entity.Infrastructure,
		}
	}

	return result, nil
}

func (i infraProvisionedStore) FindLatestByGitspaceInstanceID(
	ctx context.Context,
	spaceID int64,
	gitspaceInstanceID int64,
) (*types.InfraProvisioned, error) {
	var entity infraProvisioned
	err := i.db.Table(infraProvisionedTable).
		Select(infraProvisionedSelectColumns).
		Where("iprov_gitspace_id = ?", gitspaceInstanceID).
		Where("iprov_space_id = ?", spaceID).
		Order("iprov_created DESC").
		First(&entity).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to find latest infraprovisioned for instance %d", gitspaceInstanceID)
	}
	return entity.toDTO(), nil
}

func (i infraProvisionedStore) FindLatestByGitspaceInstanceIdentifier(
	ctx context.Context,
	spaceID int64,
	gitspaceInstanceIdentifier string,
) (*types.InfraProvisioned, error) {
	var entity infraProvisioned
	err := i.db.Table(infraProvisionedTable).
		Select(infraProvisionedSelectColumns).
		Joins(fmt.Sprintf("JOIN %s ON iprov_gitspace_id = gits_id", gitspaceInstanceTable)).
		Where("gits_uid = ?", gitspaceInstanceIdentifier).
		Where("iprov_space_id = ?", spaceID).
		Order("iprov_created DESC").
		First(&entity).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to find infraprovisioned for instance %s", gitspaceInstanceIdentifier)
	}
	return entity.toDTO(), nil
}

func (i infraProvisionedStore) Create(ctx context.Context, infraProvisionedPtr *types.InfraProvisioned) error {
	entity := &infraProvisioned{
		GitspaceInstanceID:      infraProvisionedPtr.GitspaceInstanceID,
		InfraProviderType:       infraProvisionedPtr.InfraProviderType,
		InfraProviderResourceID: infraProvisionedPtr.InfraProviderResourceID,
		SpaceID:                 infraProvisionedPtr.SpaceID,
		Created:                 infraProvisionedPtr.Created,
		Updated:                 infraProvisionedPtr.Updated,
		ResponseMetadata:        infraProvisionedPtr.ResponseMetadata,
		InputParams:             infraProvisionedPtr.InputParams,
		InfraStatus:             infraProvisionedPtr.InfraStatus,
		ServerHostIP:            infraProvisionedPtr.ServerHostIP,
		ServerHostPort:          infraProvisionedPtr.ServerHostPort,
		ProxyHost:               infraProvisionedPtr.ProxyHost,
		ProxyPort:               infraProvisionedPtr.ProxyPort,
	}

	err := i.db.Table(infraProvisionedTable).Create(entity).Error
	if err != nil {
		return errors.Wrapf(err, "infraprovisioned create query failed for instance: %d", infraProvisionedPtr.GitspaceInstanceID)
	}

	infraProvisionedPtr.ID = entity.ID
	return nil
}

func (i infraProvisionedStore) Delete(ctx context.Context, id int64) error {
	err := i.db.Table(infraProvisionedTable).
		Where(infraProvisionedIDColumn+" = ?", id).
		Delete(nil).Error
	if err != nil {
		return errors.Wrapf(err, "Failed to delete infraprovisioned for %d", id)
	}
	return nil
}

func (i infraProvisionedStore) Update(ctx context.Context, infraProvisioned *types.InfraProvisioned) error {
	err := i.db.Table(infraProvisionedTable).
		Where(infraProvisionedIDColumn+" = ?", infraProvisioned.ID).
		Updates(map[string]interface{}{
			"iprov_response_metadata": infraProvisioned.ResponseMetadata,
			"iprov_infra_status":      infraProvisioned.InfraStatus,
			"iprov_server_host_ip":    infraProvisioned.ServerHostIP,
			"iprov_server_host_port":  infraProvisioned.ServerHostPort,
			"iprov_opentofu_params":   infraProvisioned.InputParams,
			"iprov_updated":           infraProvisioned.Updated,
			"iprov_proxy_host":        infraProvisioned.ProxyHost,
			"iprov_proxy_port":        infraProvisioned.ProxyPort,
		}).Error
	if err != nil {
		return errors.Wrapf(err, "Failed to update infra provisioned for instance %d", infraProvisioned.GitspaceInstanceID)
	}
	return nil
}

func (entity infraProvisioned) toDTO() *types.InfraProvisioned {
	return &types.InfraProvisioned{
		ID:                      entity.ID,
		GitspaceInstanceID:      entity.GitspaceInstanceID,
		InfraProviderType:       entity.InfraProviderType,
		InfraProviderResourceID: entity.InfraProviderResourceID,
		SpaceID:                 entity.SpaceID,
		Created:                 entity.Created,
		Updated:                 entity.Updated,
		ResponseMetadata:        entity.ResponseMetadata,
		InputParams:             entity.InputParams,
		InfraStatus:             entity.InfraStatus,
		ServerHostIP:            entity.ServerHostIP,
		ServerHostPort:          entity.ServerHostPort,
		ProxyHost:               entity.ProxyHost,
		ProxyPort:               entity.ProxyPort,
	}
}
