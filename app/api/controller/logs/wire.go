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

package logs

import (
	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/livelog"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideController,
)

func ProvideController(
	authorizer authz.Authorizer,
	executionStore store.ExecutionStore,
	repoStore store.RepoStore,
	pipelineStore store.PipelineStore,
	stageStore store.StageStore,
	stepStore store.StepStore,
	logStore store.LogStore,
	logStream livelog.LogStream,
) *Controller {
	return NewController(authorizer, executionStore, repoStore,
		pipelineStore, stageStore, stepStore, logStore, logStream)
}
