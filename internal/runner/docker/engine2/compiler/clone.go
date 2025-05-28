// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"fmt"

	harness "github.com/easysoft/gitfox/internal/pipeline/spec"
	"github.com/easysoft/gitfox/internal/runner/docker/engine2/engine"
	"github.com/easysoft/gitfox/pkg/util/common"
)

// default name of the clone step.
const cloneStepName = "clone"

// helper function creates a default container configuration
// for the clone stage. The clone stage is automatically
// added to each pipeline.
func createClone(platform *harness.Platform, clone *harness.Clone) *engine.Step {
	return &engine.Step{
		Name:      cloneStepName,
		Image:     cloneImage(platform),
		Envs:      cloneParams(clone),
		RunPolicy: engine.RunAlways,
	}
}

// helper function returns the clone image based on the
// target operating system.
func cloneImage(platform *harness.Platform) string {
	switch platform.Os {
	case "windows":
		return common.GitImage()
	default:
		return common.GitImage()
	}
}

// helper function configures the clone depth parameter,
// specific to the clone plugin.
func cloneParams(src *harness.Clone) map[string]string {
	dst := map[string]string{}
	if v := src.Depth; v > 0 {
		dst["PLUGIN_DEPTH"] = fmt.Sprint(v)
	}
	if v := src.Insecure; v {
		dst["GIT_SSL_NO_VERIFY"] = "true"
		dst["PLUGIN_SKIP_VERIFY"] = "true"
	}
	return dst
}
