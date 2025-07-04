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

package infrastructure

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/easysoft/gitfox/infraprovider"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
)

func (i infraProvisioner) TriggerProvision(
	ctx context.Context,
	gitspaceConfig types.GitspaceConfig,
	requiredGitspacePorts []types.GitspacePort,
) error {
	infraProviderEntity, err := i.getConfigFromResource(ctx, gitspaceConfig.InfraProviderResource)
	if err != nil {
		return err
	}

	infraProvider, err := i.getInfraProvider(infraProviderEntity.Type)
	if err != nil {
		return err
	}

	if infraProvider.ProvisioningType() == enum.InfraProvisioningTypeNew {
		return i.triggerProvisionForNewProvisioning(
			ctx, infraProvider, infraProviderEntity.Type, gitspaceConfig, requiredGitspacePorts)
	}
	return i.triggerProvisionForExistingProvisioning(
		ctx, infraProvider, gitspaceConfig, requiredGitspacePorts)
}

func (i infraProvisioner) triggerProvisionForNewProvisioning(
	ctx context.Context,
	infraProvider infraprovider.InfraProvider,
	infraProviderType enum.InfraProviderType,
	gitspaceConfig types.GitspaceConfig,
	requiredGitspacePorts []types.GitspacePort,
) error {
	infraProvisionedLatest, _ := i.infraProvisionedStore.FindLatestByGitspaceInstanceID(
		ctx, gitspaceConfig.SpaceID, gitspaceConfig.GitspaceInstance.ID)

	if infraProvisionedLatest != nil &&
		infraProvisionedLatest.InfraStatus == enum.InfraStatusPending &&
		time.Since(time.UnixMilli(infraProvisionedLatest.Updated)).Milliseconds() < (10*60*1000) {
		return fmt.Errorf("there is already infra provisioning in pending state %d", infraProvisionedLatest.ID)
	} else if infraProvisionedLatest != nil {
		infraProvisionedLatest.InfraStatus = enum.InfraStatusUnknown
		err := i.infraProvisionedStore.Update(ctx, infraProvisionedLatest)
		if err != nil {
			return fmt.Errorf("could not update Infra Provisioned entity: %w", err)
		}
	}

	infraProviderResource := gitspaceConfig.InfraProviderResource
	allParams, err := i.getAllParamsFromDB(ctx, infraProviderResource, infraProvider)
	if err != nil {
		return fmt.Errorf("could not get all params from DB while provisioning: %w", err)
	}

	err = infraProvider.ValidateParams(allParams)
	if err != nil {
		return fmt.Errorf("invalid provisioning params %v: %w", infraProviderResource.Metadata, err)
	}

	now := time.Now()
	paramsBytes, err := serializeInfraProviderParams(allParams)
	if err != nil {
		return err
	}
	infraProvisioned := &types.InfraProvisioned{
		GitspaceInstanceID:      gitspaceConfig.GitspaceInstance.ID,
		InfraProviderType:       infraProviderType,
		InfraProviderResourceID: infraProviderResource.ID,
		Created:                 now.UnixMilli(),
		Updated:                 now.UnixMilli(),
		InputParams:             paramsBytes,
		InfraStatus:             enum.InfraStatusPending,
		SpaceID:                 gitspaceConfig.SpaceID,
	}

	err = i.infraProvisionedStore.Create(ctx, infraProvisioned)
	if err != nil {
		return fmt.Errorf("unable to create infraProvisioned entry for %d", gitspaceConfig.GitspaceInstance.ID)
	}

	agentPort := i.config.AgentPort

	err = infraProvider.Provision(
		ctx,
		gitspaceConfig.SpaceID,
		gitspaceConfig.SpacePath,
		gitspaceConfig.Identifier,
		gitspaceConfig.GitspaceInstance.Identifier,
		agentPort,
		requiredGitspacePorts,
		allParams,
	)
	if err != nil {
		infraProvisioned.InfraStatus = enum.InfraStatusUnknown
		infraProvisioned.Updated = time.Now().UnixMilli()
		err2 := i.infraProvisionedStore.Update(ctx, infraProvisioned)
		if err2 != nil {
			log.Err(err2).Msgf("unable to update infraProvisioned Entry for %d", infraProvisioned.ID)
		}

		return fmt.Errorf(
			"unable to trigger provision infrastructure for gitspaceConfigIdentifier %v: %w",
			gitspaceConfig.Identifier,
			err,
		)
	}

	return nil
}

func (i infraProvisioner) triggerProvisionForExistingProvisioning(
	ctx context.Context,
	infraProvider infraprovider.InfraProvider,
	gitspaceConfig types.GitspaceConfig,
	requiredGitspacePorts []types.GitspacePort,
) error {
	allParams, err := i.getAllParamsFromDB(ctx, gitspaceConfig.InfraProviderResource, infraProvider)
	if err != nil {
		return fmt.Errorf("could not get all params from DB while provisioning: %w", err)
	}

	err = infraProvider.ValidateParams(allParams)
	if err != nil {
		return fmt.Errorf("invalid provisioning params %v: %w", gitspaceConfig.InfraProviderResource.Metadata, err)
	}

	err = infraProvider.Provision(
		ctx,
		gitspaceConfig.SpaceID,
		gitspaceConfig.SpacePath,
		gitspaceConfig.Identifier,
		gitspaceConfig.GitspaceInstance.Identifier,
		0, // NOTE: Agent port is not required for provisioning type Existing.
		requiredGitspacePorts,
		allParams,
	)
	if err != nil {
		return fmt.Errorf(
			"unable to trigger provision infrastructure for gitspaceConfigIdentifier %v: %w",
			gitspaceConfig.Identifier,
			err,
		)
	}

	return nil
}

func (i infraProvisioner) ResumeProvision(
	ctx context.Context,
	gitspaceConfig types.GitspaceConfig,
	provisionedInfra types.Infrastructure,
) error {
	infraProvider, err := i.getInfraProvider(provisionedInfra.ProviderType)
	if err != nil {
		return err
	}

	if infraProvider.ProvisioningType() == enum.InfraProvisioningTypeNew {
		return i.resumeProvisionForNewProvisioning(ctx, gitspaceConfig, provisionedInfra)
	}

	return nil
}

func (i infraProvisioner) resumeProvisionForNewProvisioning(
	ctx context.Context,
	gitspaceConfig types.GitspaceConfig,
	provisionedInfra types.Infrastructure,
) error {
	infraProvisionedLatest, err := i.infraProvisionedStore.FindLatestByGitspaceInstanceID(
		ctx, gitspaceConfig.SpaceID, gitspaceConfig.GitspaceInstance.ID)
	if err != nil {
		return fmt.Errorf(
			"could not find latest infra provisioned entity for instance %d: %w",
			gitspaceConfig.GitspaceInstance.ID, err)
	}

	responseMetadata, err := i.responseMetadata(provisionedInfra)
	if err != nil {
		return err
	}

	infraProvisionedLatest.InfraStatus = provisionedInfra.Status
	infraProvisionedLatest.ServerHostIP = provisionedInfra.AgentHost
	infraProvisionedLatest.ServerHostPort = strconv.Itoa(provisionedInfra.AgentPort)
	infraProvisionedLatest.ProxyHost = provisionedInfra.ProxyAgentHost
	infraProvisionedLatest.ProxyPort = int32(provisionedInfra.ProxyAgentPort)
	infraProvisionedLatest.ResponseMetadata = &responseMetadata
	infraProvisionedLatest.Updated = time.Now().UnixMilli()

	err = i.infraProvisionedStore.Update(ctx, infraProvisionedLatest)
	if err != nil {
		return fmt.Errorf("unable to update infraProvisioned Entry %d", infraProvisionedLatest.ID)
	}

	return nil
}
