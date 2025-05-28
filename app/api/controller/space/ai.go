// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.
package space

import (
	"context"
	"fmt"
	"time"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/pkg/util/common"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
)

type AiConfigCreateInput struct {
	IsDefault bool          `json:"is_default,omitempty"`
	Provider  enum.Provider `json:"provider"`
	Model     string        `json:"model"`
	Endpoint  string        `json:"endpoint"`
	Token     string        `json:"token"`
}

type AiConfigUpdateInput struct {
	IsDefault bool          `json:"is_default,omitempty"`
	Provider  enum.Provider `json:"provider,omitempty"`
	Model     string        `json:"model,omitempty"`
	Endpoint  string        `json:"endpoint,omitempty"`
	Token     string        `json:"token,omitempty"`
}

func (in *AiConfigCreateInput) Sanitize() error {
	if !in.Provider.IsValid() {
		return fmt.Errorf("invalid provider: %s", in.Provider)
	}
	if len(in.Model) == 0 {
		return fmt.Errorf("model is required")
	}
	if len(in.Endpoint) == 0 {
		return fmt.Errorf("endpoint is required")
	}
	if len(in.Token) == 0 {
		return fmt.Errorf("token is required")
	}
	return nil
}

// ListAIConfigs lists the ai configs in a space.
func (c *Controller) ListAIConfigs(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
) ([]types.AIConfig, int64, error) {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find parent space: %w", err)
	}

	err = apiauth.CheckSecret(
		ctx,
		c.authorizer,
		session,
		space.Path,
		"",
		enum.PermissionAIView,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("could not authorize: %w", err)
	}

	var count int64
	var aiCfgs []types.AIConfig

	err = c.tx.WithTx(ctx, func(ctx context.Context) (err error) {
		count, err = c.aiStore.Count(ctx, space.ID)
		if err != nil {
			return fmt.Errorf("failed to count space %v ai configs: %w", spaceRef, err)
		}

		aiCfgs, err = c.aiStore.List(ctx, space.ID)
		if err != nil {
			return fmt.Errorf("failed to list space %v ai configs: %w", spaceRef, err)
		}
		return
	}, dbtx.TxDefaultReadOnly)
	if err != nil {
		return aiCfgs, count, fmt.Errorf("failed to list space %v ai configs: %w", spaceRef, err)
	}

	return aiCfgs, count, nil
}

// CreateAIConfig creates a new ai config in a space.
func (c *Controller) CreateAIConfig(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	in *AiConfigCreateInput,
) (*types.AIConfig, error) {
	if err := in.Sanitize(); err != nil {
		return nil, fmt.Errorf("invalid ai config: %w", err)
	}
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return nil, fmt.Errorf("failed to find parent space: %w", err)
	}
	err = apiauth.CheckSecret(
		ctx,
		c.authorizer,
		session,
		space.Path,
		"",
		enum.PermissionAIEdit,
	)
	if err != nil {
		return nil, fmt.Errorf("could not authorize: %w", err)
	}
	now := time.Now().UnixMilli()
	cfg := &types.AIConfig{
		SpaceID:   space.ID,
		Provider:  in.Provider,
		Model:     in.Model,
		Endpoint:  in.Endpoint,
		Token:     in.Token,
		Created:   now,
		Updated:   now,
		CreatedBy: session.Principal.ID,
		UpdatedBy: session.Principal.ID,
	}
	cfgs, _ := c.aiStore.List(ctx, space.ID)
	if len(cfgs) == 0 {
		cfg.IsDefault = true
	}
	if err := c.aiStore.Create(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to create ai config: %w", err)
	}
	return cfg, nil
}

func (c *Controller) TestAIConfig(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	id int64,
) error {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return fmt.Errorf("failed to find parent space: %w", err)
	}
	err = apiauth.CheckSecret(
		ctx,
		c.authorizer,
		session,
		space.Path,
		"",
		enum.PermissionAIEdit,
	)
	if err != nil {
		return fmt.Errorf("could not authorize: %w", err)
	}
	cfg, err := c.aiStore.Find(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find ai config: %w", err)
	}
	now := time.Now().UnixMilli()
	aiReq := &types.AIRequest{
		Created:  now,
		Updated:  now,
		ConfigID: cfg.ID,
	}
	defer c.saveAITestRequest(ctx, space, aiReq)
	log.Ctx(ctx).Trace().Msgf("test [%d] %s ai config", id, cfg.Provider.String())
	client, err := common.GetClient(cfg)
	if err != nil {
		return c.handleAIError(aiReq, fmt.Errorf("failed to get client: %w", err))
	}
	resp, err := client.Completion(ctx, "test model")
	if err != nil {
		return c.handleAIError(aiReq, fmt.Errorf("failed to test ai config: %w", err))
	}
	aiReq.Duration = time.Now().UnixMilli() - now
	aiReq.Token = int64(resp.Usage.TotalTokens)
	aiReq.Status = enum.AIRequestStatusSuccess
	log.Ctx(ctx).Trace().Msgf("test [%d] %s ai config done, usage: %v tokens", id, cfg.Provider.String(), resp.Usage.TotalTokens)
	return nil
}

func (c *Controller) saveAITestRequest(ctx context.Context, space *types.Space, aiReq *types.AIRequest) {
	repos, _ := c.repoStore.List(ctx, space.ID, &types.RepoFilter{})
	if len(repos) > 0 {
		aiReq.RepoID = repos[0].ID
	}
	if err := c.aiStore.Record(ctx, aiReq); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to save ai request record")
	}
}

func (c *Controller) handleAIError(aiReq *types.AIRequest, err error) error {
	aiReq.Status = enum.AIRequestStatusFailed
	aiReq.Error = err.Error()
	return err
}

func (c *Controller) UpdateAIConfig(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	id int64,
	in *AiConfigUpdateInput,
) (*types.AIConfig, error) {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return nil, fmt.Errorf("failed to find parent space: %w", err)
	}
	err = apiauth.CheckSecret(
		ctx,
		c.authorizer,
		session,
		space.Path,
		"",
		enum.PermissionAIEdit,
	)
	if err != nil {
		return nil, fmt.Errorf("could not authorize: %w", err)
	}
	cfg, err := c.aiStore.Find(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find ai config: %w", err)
	}
	log.Ctx(ctx).Trace().Msgf("found [%d] %s ai config", id, cfg.Provider.String())
	if in.Endpoint != "" {
		cfg.Endpoint = in.Endpoint
	}
	if in.Token != "" {
		cfg.Token = in.Token
	}
	if in.Model != "" {
		cfg.Model = in.Model
	}
	if err := c.aiStore.Update(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to update ai config: %w", err)
	}
	return cfg, nil
}

func (c *Controller) DeleteAIConfig(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	id int64,
) error {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return fmt.Errorf("failed to find parent space: %w", err)
	}
	err = apiauth.CheckSecret(
		ctx,
		c.authorizer,
		session,
		space.Path,
		"",
		enum.PermissionAIEdit,
	)
	if err != nil {
		return fmt.Errorf("could not authorize: %w", err)
	}
	if err := c.aiStore.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete ai config: %w", err)
	}
	// TODO: 如果提供商都删除了，则需要禁用所有仓库AI评审
	return nil
}

func (c *Controller) SetDefaultAIConfig(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	id int64,
) error {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return fmt.Errorf("failed to find parent space: %w", err)
	}
	err = apiauth.CheckSecret(
		ctx,
		c.authorizer,
		session,
		space.Path,
		"",
		enum.PermissionAIEdit,
	)
	if err != nil {
		return fmt.Errorf("could not authorize: %w", err)
	}
	cfg, err := c.aiStore.Find(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find ai config: %w", err)
	}
	cfgOld, err := c.aiStore.Default(ctx, space.ID)
	if err != nil {
		return fmt.Errorf("failed to find default ai config: %w", err)
	}
	if cfgOld != nil && cfgOld.ID != cfg.ID {
		cfgOld.IsDefault = false
		if err := c.aiStore.Update(ctx, cfgOld); err != nil {
			return fmt.Errorf("failed to update default ai config: %w", err)
		}
	}
	cfg.IsDefault = true
	if err := c.aiStore.Update(ctx, cfg); err != nil {
		return fmt.Errorf("failed to update ai config: %w", err)
	}
	return nil
}
