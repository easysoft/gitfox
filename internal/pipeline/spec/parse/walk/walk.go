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

package walk

import (
	"errors"

	schema "github.com/easysoft/gitfox/internal/pipeline/spec"
)

// Func is the type of the function called for
// stages and steps visited by Walk.
type Func func(interface{}) error

// ErrSkip is used as a return value for Walk to indicate
// child stages or steps should be skipped.
var ErrSkip = errors.New("skip this node")

// ErrSkipAll is used as a return value for Walk to indicate
// all subsequent nodes should be skipped
var ErrSkipAll = errors.New("skip all node")

// Walk walks the configuration file and calls fn for
// stages and steps.
func Walk(in *schema.Config, fn Func) error {
	switch v := in.Spec.(type) {
	case *schema.Pipeline:
		return walkPipeline(v, fn)
	case *schema.PluginStep:
	case *schema.PluginStage:
	case *schema.TemplateStage:
	case *schema.TemplateStep:
	}

	return nil
}

func walkPipeline(pipeline *schema.Pipeline, fn Func) error {
	err := fn(pipeline)
	switch {
	case err == ErrSkip:
		return nil
	case err != nil:
		return err
	}

	for _, vv := range pipeline.Stages {
		err := walkStage(vv, fn)
		switch {
		case err == ErrSkip:
		case err != nil:
			return err
		}
	}

	return nil
}

func walkStage(stage *schema.Stage, fn Func) error {
	err := fn(stage)
	switch {
	case err == ErrSkip:
		return nil
	case err != nil:
		return err
	}
	switch v := stage.Spec.(type) {
	case *schema.StageCI:
		for _, vv := range v.Steps {
			walkStep(vv, fn)
			if err != nil {
				return err
			}
		}
	case *schema.StageGroup:
		for _, vv := range v.Stages {
			walkStage(vv, fn)
			if err != nil {
				return err
			}
		}
	case *schema.StageParallel:
		for _, vv := range v.Stages {
			err = walkStage(vv, fn)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func walkStep(step *schema.Step, fn Func) error {
	err := fn(step)
	switch {
	case err == ErrSkip:
		return nil
	case err != nil:
		return err
	}
	switch v := step.Spec.(type) {
	case *schema.StepGroup:
		for _, vv := range v.Steps {
			err = walkStep(vv, fn)
			if err != nil {
				return err
			}
		}
	case *schema.StepParallel:
		for _, vv := range v.Steps {
			err = walkStep(vv, fn)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
