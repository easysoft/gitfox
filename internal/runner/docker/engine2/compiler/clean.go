// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package compiler

import (
	"github.com/easysoft/gitfox/internal/runner/docker/engine2/engine"
)

const cleanStepName = "clean"

func createStageClean() *engine.Step {
	return &engine.Step{
		Name:      cleanStepName,
		Image:     "hub.zentao.net/ci/pipeline-helper",
		Envs:      make(map[string]string),
		RunPolicy: engine.RunAlways,
	}
}
