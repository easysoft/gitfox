// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package execution

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
)

const (
	customEnvKey = ".custom_env"
)

type stageMetadata struct {
	repo      *types.Repository
	pipeline  *types.Pipeline
	execution *types.Execution
	stage     *types.Stage
}

func (c *Controller) StashFile(
	ctx context.Context,
	session *auth.Session,
	executionNum int64,
	stageNum int64,
	req *http.Request,
) (interface{}, error) {
	meta, err := c.validateStage(ctx, session, executionNum, stageNum)
	if err != nil {
		return nil, err
	}

	key := req.FormValue("key")
	if key == "" {
		return nil, errors.New("require field 'key'")
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		return nil, err
	}

	switch key {
	case customEnvKey:
		content, _ := io.ReadAll(file)
		return nil, c.store.PutStageFile(ctx, meta.execution.ID, meta.stage.ID, key, content)
	default:
		return nil, errors.New("unsupported key")
	}
}

func (c *Controller) validateStage(
	ctx context.Context,
	session *auth.Session,
	executionId int64,
	stageId int64,
) (*stageMetadata, error) {
	stage, err := c.stageStore.Find(ctx, stageId)
	if err != nil {
		return nil, fmt.Errorf("failed to find stage: %w", err)
	}

	if stage.ExecutionID != executionId {
		return nil, fmt.Errorf("stage execution id does not match pipeline execution id")
	}

	execution, err := c.executionStore.Find(ctx, executionId)
	if err != nil {
		return nil, fmt.Errorf("failed to find execution: %w", err)
	}

	pipeline, err := c.pipelineStore.Find(ctx, execution.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pipeline: %w", err)
	}

	repo, err := c.repoStore.Find(ctx, execution.RepoID)
	if err != nil {
		return nil, fmt.Errorf("failed to find repo by id: %w", err)
	}
	err = apiauth.CheckPipeline(ctx, c.authorizer, session, repo.Path, pipeline.Identifier, enum.PermissionPipelineView)
	if err != nil {
		log.Ctx(ctx).Info().Msg("todo: add auth")
		//return nil, fmt.Errorf("failed to authorize pipeline: %w", err)
	}

	return &stageMetadata{
		repo:      repo,
		pipeline:  pipeline,
		execution: execution,
		stage:     stage,
	}, nil
}
