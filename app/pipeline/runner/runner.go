// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package runner

import (
	"context"

	"github.com/easysoft/gitfox/app/pipeline/resolver"
	"github.com/easysoft/gitfox/internal/runner/extend"
	"github.com/easysoft/gitfox/types"

	"github.com/drone/drone-go/drone"
	runnerclient "github.com/drone/runner-go/client"
)

type Runner interface {
	Run(ctx context.Context, stage *drone.Stage) error
}

func NewExecutionRunner(
	config *types.Config,
	client runnerclient.Client,
	resolver *resolver.Manager,
	stageEnvProvider *extend.StageDynamicEnvProvider,
) (Runner, error) {
	return newDockerRunner(config, client, resolver, stageEnvProvider)
}
