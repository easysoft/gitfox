// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package extend

import (
	"bytes"
	"context"
	"errors"

	"github.com/easysoft/gitfox/app/pipeline/storage"
	"github.com/easysoft/gitfox/app/store"
	schema "github.com/easysoft/gitfox/internal/pipeline/spec"

	"github.com/drone/runner-go/environ/provider"
	"github.com/rs/zerolog/log"
)

type StageDynamicEnvProvider struct {
	store      storage.PipelineStorage
	stageStore store.StageStore
}

func NewStageDynamicEnvProvider(store storage.PipelineStorage, stageStore store.StageStore) *StageDynamicEnvProvider {
	return &StageDynamicEnvProvider{
		store:      store,
		stageStore: stageStore,
	}
}

func (p *StageDynamicEnvProvider) List(ctx context.Context, in *provider.Request) ([]*provider.Variable, error) {
	if p.store == nil {
		return nil, nil
	}

	stageCtx := FromContext(ctx)
	if stageCtx == nil {
		return nil, nil
	}

	stageSpec, ok := stageCtx.stage.Spec.(*schema.StageCI)
	if !ok {
		return nil, errors.New("ci stage resource not found")
	}

	if len(stageSpec.Vars) == 0 {
		return nil, nil
	}

	var dynamicEnvs = make([]*provider.Variable, 0)
	for _, varObj := range stageSpec.Vars {
		if varObj.Type != schema.VarTypeStage {
			continue
		}

		srcStage, stageNum := loadSourceStage(varObj.Name, stageCtx.pipeline)
		if srcStage == nil {
			continue
		}

		dbStage, _ := p.stageStore.FindByNumber(ctx, in.Build.ID, stageNum)
		envBytes, e := p.store.GetStageFile(ctx, dbStage.ExecutionID, dbStage.ID, ".custom_env")
		if e != nil {
			return nil, e
		}

		for _, line := range bytes.Split(envBytes, []byte("\n")) {
			i := bytes.Index(line, []byte("="))
			if i == -1 {
				continue
			}
			key := string(line[:i])
			value := string(line[i+1:])
			dynamicEnvs = append(dynamicEnvs, &provider.Variable{
				Name: key, Data: value,
			})
			log.Ctx(ctx).Info().Msgf("add variable: %s", key)
		}
	}

	return dynamicEnvs, nil
}

func loadSourceStage(key string, in *schema.Pipeline) (*schema.Stage, int) {
	var s *schema.Stage
	var number int
	for idx, stage := range in.Stages {
		if stage.Name == key {
			s = stage
			number = idx + 1
			break
		}
	}
	return s, number
}
