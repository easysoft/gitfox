// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package linter

import (
	"errors"

	"github.com/easysoft/gitfox/pkg/util/yaml"
)

// ErrDuplicatePipelineName is returned when two Pipeline
// resources have the same name.
var ErrDuplicatePipelineName = errors.New("linter: duplicate pipeline names")

// ErrMissingPipelineDependency is returned when a Pipeline
// defines dependencies that are invalid or unknown.
var ErrMissingPipelineDependency = errors.New("linter: invalid or unknown pipeline dependency")

// ErrCyclicalPipelineDependency is returned when a Pipeline
// defines a cyclical dependency, which would result in an
// infinite execution loop.
var ErrCyclicalPipelineDependency = errors.New("linter: cyclical pipeline dependency detected")

// ErrPipelineSelfDependency is returned when a Pipeline
// defines a dependency on itself.
var ErrPipelineSelfDependency = errors.New("linter: pipeline cannot have a dependency on itself")

// Manifest performs lint operations for a manifest.
func Manifest(manifest *yaml.Manifest, trusted bool) error {
	return checkPipelines(manifest, trusted)
}

func checkPipelines(manifest *yaml.Manifest, trusted bool) error {
	names := map[string]struct{}{}
	for _, resource := range manifest.Resources {
		switch v := resource.(type) {
		case *yaml.Pipeline:
			_, ok := names[v.Name]
			if ok {
				return ErrDuplicatePipelineName
			}
			names[v.Name] = struct{}{}
			err := checkPipelineDeps(v, names)
			if err != nil {
				return err
			}
			if (v.Kind == "pipeline" || v.Kind == "") && (v.Type == "" || v.Type == "docker") {
				err = checkPlatform(v.Platform)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkPipelineDeps(pipeline *yaml.Pipeline, deps map[string]struct{}) error {
	for _, dep := range pipeline.DependsOn {
		_, ok := deps[dep]
		if !ok {
			return ErrMissingPipelineDependency
		}
		if pipeline.Name == dep {
			return ErrPipelineSelfDependency
		}
	}
	return nil
}
