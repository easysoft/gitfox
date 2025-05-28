// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox/pkg/storage"
)

const (
	pipelinePathFormatter = "executions/%d/stages/%d/%s"
)

type PipelineStorage interface {
	PutStageFile(ctx context.Context, executionId, stageId int64, fileName string, data []byte) error
	GetStageFile(ctx context.Context, executionId, stageId int64, fileName string) ([]byte, error)
}

type pipelineStorage struct {
	base storage.ContentStorage
}

func (s *pipelineStorage) PutStageFile(ctx context.Context, executionId, stageId int64, fileName string, data []byte) error {
	absPath := fmt.Sprintf(pipelinePathFormatter, executionId, stageId, fileName)
	return s.base.Put(ctx, absPath, data)
}

func (s *pipelineStorage) GetStageFile(ctx context.Context, executionId, stageId int64, fileName string) ([]byte, error) {
	absPath := fmt.Sprintf(pipelinePathFormatter, executionId, stageId, fileName)
	return s.base.Get(ctx, absPath)
}
