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

package resolver

import (
	"errors"

	schema "github.com/easysoft/gitfox/internal/pipeline/spec"
)

// Lookup returns a resource by name, kind and type.
type LookupFunc func(name, kind, typ, version string) (*schema.Config, error)

// Resolve injects named resources in the pipeline. This
// function is recursive and inject stage and step resources.
func Resolve(in *schema.Config, fn LookupFunc) error {
	switch v := in.Spec.(type) {
	case *schema.Pipeline:
		for _, vv := range v.Stages {
			if err := resolveStage(vv, fn); err != nil {
				return err
			}
		}
	case *schema.PluginStep:
	case *schema.PluginStage:
	case *schema.TemplateStage:
	case *schema.TemplateStep:
	}

	return nil
}

// ResolveStage injects named resource in the stage. This
// is non-recursive and does not inject in steps.
func ResolveStage(stage *schema.Stage, fn LookupFunc) error {
	switch v := stage.Spec.(type) {
	case *schema.StageGroup:
		for _, vv := range v.Stages {
			if err := ResolveStage(vv, fn); err != nil {
				return err
			}
		}
	case *schema.StageParallel:
		for _, vv := range v.Stages {
			if err := ResolveStage(vv, fn); err != nil {
				return err
			}
		}
	case *schema.StageTemplate:
		config, err := fn(v.Name, "template", "stage", "")
		if err != nil {
			return err
		}
		switch vv := config.Spec.(type) {
		case *schema.TemplateStage:
			if vv.Stage == nil {
				return errors.New("invalid template stage")
			}
			stage.Spec = vv.Stage.Spec
			stage.Type = vv.Stage.Type
		default:
			return errors.New("invalid resource type")
		}
	}

	return nil
}

func resolveStage(stage *schema.Stage, fn LookupFunc) error {
	switch v := stage.Spec.(type) {
	case *schema.StageCI:
		for _, vv := range v.Steps {
			if err := resolveStep(vv, fn); err != nil {
				return err
			}
		}
	case *schema.StageGroup:
		for _, vv := range v.Stages {
			if err := resolveStage(vv, fn); err != nil {
				return err
			}
		}
	case *schema.StageParallel:
		for _, vv := range v.Stages {
			if err := resolveStage(vv, fn); err != nil {
				return err
			}
		}
	case *schema.StageTemplate:
		config, err := fn(v.Name, "template", "stage", "")
		if err != nil {
			return err
		}
		switch vv := config.Spec.(type) {
		case *schema.TemplateStage:
			if vv.Stage == nil {
				return errors.New("invalid template stage")
			}
			stage.Spec = vv.Stage.Spec
			stage.Type = vv.Stage.Type
			stage.Inputs = map[string]interface{}{}
			for k, v := range vv.Inputs {
				if v != nil {
					stage.Inputs[k] = v.Default
				}
			}
			for k, v := range v.Inputs {
				stage.Inputs[k] = v
			}
		default:
			return errors.New("invalid resource type")
		}
	}

	return nil
}

func resolveStep(step *schema.Step, fn LookupFunc) error {
	switch v := step.Spec.(type) {
	case *schema.StepGroup:
		for _, vv := range v.Steps {
			if err := resolveStep(vv, fn); err != nil {
				return err
			}
		}
	case *schema.StepParallel:
		for _, vv := range v.Steps {
			if err := resolveStep(vv, fn); err != nil {
				return err
			}
		}
	case *schema.StepTemplate:
		config, err := fn(v.Name, "template", "step", "")
		if err != nil {
			return err
		}
		switch vv := config.Spec.(type) {
		case *schema.TemplateStep:
			if vv.Step == nil {
				return errors.New("invalid template step")
			}
			step.Spec = vv.Step.Spec
			step.Type = vv.Step.Type
			step.Inputs = map[string]interface{}{}
			for k, v := range vv.Inputs {
				if v != nil {
					step.Inputs[k] = v.Default
				}
			}
			for k, v := range v.Inputs {
				step.Inputs[k] = v
			}
		default:
			return errors.New("invalid resource type")
		}
	case *schema.StepPlugin:
		name := v.Name
		if s := v.Uses; s != "" {
			name = s
		}
		config, err := fn(name, "plugin", "step", "")
		if err != nil {
			return err
		}
		switch vv := config.Spec.(type) {
		case *schema.PluginStep:
			if vv.Step == nil {
				return errors.New("invalid template step")
			}
			step.Spec = vv.Step.Spec
			step.Type = vv.Step.Type
			step.Inputs = map[string]interface{}{}
			for k, v := range vv.Inputs {
				if v != nil {
					step.Inputs[k] = v.Default
				}
			}
			for k, v := range v.Inputs {
				step.Inputs[k] = v
			}

		default:
			return errors.New("invalid resource type")
		}
	}

	return nil
}
