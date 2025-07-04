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

package execution

import (
	"context"
	"fmt"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

func (c *Controller) List(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	pipelineIdentifier string,
	pagination types.Pagination,
	createdAfter int64,
) ([]*types.Execution, int64, error) {
	repo, err := c.repoStore.FindByRef(ctx, repoRef)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find repo by ref: %w", err)
	}

	err = apiauth.CheckPipeline(ctx, c.authorizer, session, repo.Path, pipelineIdentifier, enum.PermissionPipelineView)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to authorize: %w", err)
	}

	pipeline, err := c.pipelineStore.FindByIdentifier(ctx, repo.ID, pipelineIdentifier)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find pipeline: %w", err)
	}

	var count int64
	var executions []*types.Execution

	err = c.tx.WithTx(ctx, func(ctx context.Context) (err error) {
		count, err = c.executionStore.Count(ctx, pipeline.ID, createdAfter)
		if err != nil {
			return fmt.Errorf("failed to count child executions: %w", err)
		}

		executions, err = c.executionStore.List(ctx, pipeline.ID, pagination, createdAfter)
		if err != nil {
			return fmt.Errorf("failed to list child executions: %w", err)
		}

		return
	}, dbtx.TxDefaultReadOnly)
	if err != nil {
		return executions, count, fmt.Errorf("failed to fetch list: %w", err)
	}

	return executions, count, nil
}
