// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	harness "github.com/easysoft/gitfox/internal/pipeline/spec"
	"github.com/easysoft/gitfox/internal/runner/docker/engine2/engine"
)

type state struct {
	options   *harness.Default
	platform  *harness.Platform
	resources *harness.Resources
	labels    map[string]string
	envs      map[string]string
}

func convertStep(stage *harness.Stage, step *harness.Step) []*engine.Step {
	switch v := step.Spec.(type) {
	case *harness.StepExec:
		dst := createStep(step, v)
		dst.WorkingDir = "/gitfox"
		setupScript(dst, v.Run, "linux")
		return []*engine.Step{dst}
	case *harness.StepRun:
		dst := createRunStep(step, v)
		dst.WorkingDir = "/gitfox"
		setupScripts(dst, v.Script, "linux")
		return []*engine.Step{dst}
	case *harness.StepBackground:
		dst := createStepBackground(step, v)
		setupScript(dst, v.Run, "linux")
		return []*engine.Step{dst}
	case *harness.StepParallel:
		var steps []*engine.Step
		for _, vv := range v.Steps {
			steps = append(steps, convertStep(stage, vv)...)
		}
		return steps
	case *harness.StepGroup:
		var steps []*engine.Step
		for _, vv := range v.Steps {
			steps = append(steps, convertStep(stage, vv)...)
		}
		return steps
	case *harness.StepPlugin:
		dst := createStepPlugin(step, v)
		return []*engine.Step{dst}
	case *harness.StepTemplate:
		dst := createStepTemplate(step, v)
		return []*engine.Step{dst}
	}
	return nil
}
