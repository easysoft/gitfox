// Copyright 2022 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package normalize

import (
	schema "github.com/easysoft/gitfox/internal/pipeline/spec"
)

// Normalize normalizes the yaml to ensure all stages and
// steps have unique identifiers.
func Normalize(in *schema.Config) error {
	gen := newGenerator()

	switch v := in.Spec.(type) {
	case *schema.Pipeline:
		for _, vv := range v.Stages {
			normalizeStage(vv, gen)
		}
	case *schema.PluginStep:
	case *schema.PluginStage:
	case *schema.TemplateStage:
	case *schema.TemplateStep:
	}

	return nil
}

// normalize the stage
func normalizeStage(stage *schema.Stage, gen *generator) {
	stage.Id = gen.generate(stage.Id, stage.Name, stage.Type)

	switch v := stage.Spec.(type) {
	case *schema.StageCI:
		for _, vv := range v.Steps {
			normalizeStep(vv, gen)
		}
	case *schema.StageGroup:
		for _, vv := range v.Stages {
			normalizeStage(vv, gen)
		}
	case *schema.StageParallel:
		for _, vv := range v.Stages {
			normalizeStage(vv, gen)
		}
	}
}

// normalize the step
func normalizeStep(step *schema.Step, gen *generator) {
	step.Id = gen.generate(step.Id, step.Name, step.Type)

	switch v := step.Spec.(type) {
	case *schema.StepGroup:
		for _, vv := range v.Steps {
			normalizeStep(vv, gen)
		}
	case *schema.StepParallel:
		for _, vv := range v.Steps {
			normalizeStep(vv, gen)
		}
	}
}
