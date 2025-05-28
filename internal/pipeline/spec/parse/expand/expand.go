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

package expand

import (
	"encoding/json"

	schema "github.com/easysoft/gitfox/internal/pipeline/spec"
	"github.com/easysoft/gitfox/internal/pipeline/spec/parse/expand/matrix"
)

// Expand expands the matrix strategies.
func Expand(in *schema.Config) error {
	switch v := in.Spec.(type) {
	case *schema.Pipeline:
		for _, vv := range v.Stages {
			expandStageMatrix(vv)
		}
	case *schema.PluginStep:
	case *schema.PluginStage:
	case *schema.TemplateStage:
	case *schema.TemplateStep:
	}

	return nil
}

// expand the stage matrix
func expandStageMatrix(v *schema.Stage) {
	if v.Strategy != nil {
		vv, ok := v.Strategy.Spec.(*schema.Matrix)
		if ok {
			// calculate the matrix permutations.
			perms := matrix.Calc(vv.Axis)

			var stages []*schema.Stage
			// create a new stage for each item in the
			// matrix and update the strategy to include
			// ony the relevant matrix includes for this
			// specific permutation.
			for _, perm := range perms {
				stage := &schema.Stage{
					Id:       v.Id,
					Name:     v.Name,
					Delegate: v.Delegate,
					Status:   v.Status,
					Type:     v.Type,
					When:     v.When,
					Failure:  v.Failure,
					Spec:     v.Spec,
					Strategy: &schema.Strategy{
						Type: "matrix",
						Spec: &schema.Matrix{
							Include: []map[string]string{perm},
						},
					},
				}

				// we need to make a deep copy to prevent
				// multiple stages from sharing the same
				// child objects
				stage = deepCopyStage(stage)

				expandStage(stage)
				stages = append(stages, stage)
			}

			// change the stage to a parallel stage,
			// that executes matrix permutations in
			// parallel.
			v.Strategy = nil
			v.Type = "parallel"
			v.Spec = &schema.StageParallel{
				Stages: stages,
			}
			return
		}
	}

	// if the stage cannot be expanded, we
	// traverse and expand the child nodes.
	expandStage(v)
}

// expand the stage
func expandStage(v *schema.Stage) {
	switch v := v.Spec.(type) {
	case *schema.StageCI:
		for _, vv := range v.Steps {
			expandStepMatrix(vv)
		}
	case *schema.StageGroup:
		for _, vv := range v.Stages {
			expandStageMatrix(vv)
		}
	case *schema.StageParallel:
		for _, vv := range v.Stages {
			expandStageMatrix(vv)
		}
	}
}

// expand the step matrix
func expandStepMatrix(v *schema.Step) {
	if v.Strategy != nil {
		vv, ok := v.Strategy.Spec.(*schema.Matrix)
		if ok {
			// calculate the matrix permutations.
			perms := matrix.Calc(vv.Axis)

			var steps []*schema.Step
			// create a new stage for each item in the
			// matrix and update the strategy to include
			// ony the relevant matrix includes for this
			// specific permutation.
			for _, perm := range perms {
				step := &schema.Step{
					Id:      v.Id,
					Name:    v.Name,
					Type:    v.Type,
					Timeout: v.Timeout,
					When:    v.When,
					Failure: v.Failure,
					Spec:    v.Spec,
					Strategy: &schema.Strategy{
						Type: "matrix",
						Spec: &schema.Matrix{
							Include: []map[string]string{perm},
						},
					},
				}

				// we need to make a deep copy to prevent
				// multiple steps from sharing the same
				// child objects
				step = deepCopyStep(step)

				expandStep(step)
				steps = append(steps, step)
			}

			// change the stage to a parallel stage,
			// that executes matrix permutations in
			// parallel.
			v.Strategy = nil
			v.Type = "parallel"
			v.Spec = &schema.StepParallel{
				Steps: steps,
			}
			return
		}
	}

	// if the stage cannot be expanded, we
	// traverse and expand the child nodes.
	expandStep(v)
}

// expand the step
func expandStep(v *schema.Step) {
	switch v := v.Spec.(type) {
	case *schema.StepGroup:
		for _, vv := range v.Steps {
			expandStepMatrix(vv)
		}
	case *schema.StepParallel:
		for _, vv := range v.Steps {
			expandStepMatrix(vv)
		}
	}
}

// helper function creates a deep copy of a stage
func deepCopyStage(in *schema.Stage) *schema.Stage {
	out := new(schema.Stage)
	raw, _ := json.Marshal(in)   // assumes no errors
	_ = json.Unmarshal(raw, out) // assumes no errors
	return out
}

// helper function creates a deep copy of a stage
func deepCopyStep(in *schema.Step) *schema.Step {
	out := new(schema.Step)
	raw, _ := json.Marshal(in)   // assumes no errors
	_ = json.Unmarshal(raw, out) // assumes no errors
	return out
}
