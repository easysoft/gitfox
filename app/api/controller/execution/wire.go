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
	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/pipeline/canceler"
	"github.com/easysoft/gitfox/app/pipeline/commit"
	"github.com/easysoft/gitfox/app/pipeline/storage"
	"github.com/easysoft/gitfox/app/pipeline/triggerer"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database/dbtx"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideController,
)

func ProvideController(
	tx dbtx.Transactor,
	authorizer authz.Authorizer,
	executionStore store.ExecutionStore,
	checkStore store.CheckStore,
	canceler canceler.Canceler,
	commitService commit.Service,
	triggerer triggerer.Triggerer,
	repoStore store.RepoStore,
	stageStore store.StageStore,
	pipelineStore store.PipelineStore,
	pipelineStorage storage.PipelineStorage,
) *Controller {
	return NewController(tx, authorizer, executionStore, checkStore,
		canceler, commitService, triggerer, repoStore, stageStore, pipelineStore, pipelineStorage)
}
