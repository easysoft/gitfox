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

package runner

import (
	"github.com/easysoft/gitfox/app/pipeline/resolver"
	"github.com/easysoft/gitfox/internal/runner/extend"
	"github.com/easysoft/gitfox/types"

	runnerclient "github.com/drone/runner-go/client"
	"github.com/drone/runner-go/poller"
	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideExecutionRunner,
	ProvideExecutionPoller,
)

// ProvideExecutionRunner provides an execution runner.
func ProvideExecutionRunner(
	config *types.Config,
	client runnerclient.Client,
	resolver *resolver.Manager,
	stageEnvProvider *extend.StageDynamicEnvProvider,
) (Runner, error) {
	return NewExecutionRunner(config, client, resolver, stageEnvProvider)
}

// ProvideExecutionPoller provides a poller which can poll the manager
// for new builds and execute them.
func ProvideExecutionPoller(
	runner Runner,
	client runnerclient.Client,
) *poller.Poller {
	return NewExecutionPoller(runner, client)
}
