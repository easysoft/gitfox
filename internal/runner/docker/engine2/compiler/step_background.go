// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	harness "github.com/easysoft/gitfox/internal/pipeline/spec"
	"github.com/easysoft/gitfox/internal/runner/docker/engine2/engine"
	"github.com/easysoft/gitfox/internal/runner/docker/internal/docker/image"

	"github.com/drone/runner-go/environ"
)

func createStepBackground(src *harness.Step, spec *harness.StepBackground) *engine.Step {
	dst := &engine.Step{
		ID:         random(),
		Name:       src.Id,
		Image:      image.Expand(spec.Image),
		Command:    spec.Args,
		Entrypoint: nil,
		Detach:     true,
		// TODO re-enable
		// DependsOn:    src.DependsOn,
		// DNS:          spec.DNS,
		// TODO re-enable
		// DNSSearch:    spec.DNSSearch,
		Envs: spec.Envs,
		// TODO re-enable
		// ExtraHosts:   spec.ExtraHosts,
		IgnoreStderr: false,
		IgnoreStdout: false,
		Network:      spec.Network,
		Privileged:   spec.Privileged,
		Pull:         convertPullPolicy(spec.Pull),
		User:         spec.User,
		// TODO re-enable
		// Secrets:      convertSecretEnv(src.Environment),
		// TODO re-enable
		// ShmSize:    int64(spec.ShmSize),
		// TODO re-enable
		WorkingDir: spec.Workdir,

		//
		//
		//

		Networks: nil, // set in compiler.go
		Volumes:  nil, // set below
		Devices:  nil, // see below
		// Resources:    toResources(src), // TODO

		Display: src.Name,
	}

	if spec.Entrypoint != "" {
		dst.Entrypoint = []string{spec.Entrypoint}
	}

	if dst.Envs == nil {
		dst.Envs = map[string]string{}
	}

	if container := spec.Container; container != nil {
		dst.Image = image.Expand(container.Image)
		dst.Command = container.Args
		dst.Network = container.Network
		dst.Privileged = container.Privileged
		dst.Pull = convertPullPolicy(container.Pull)
		dst.User = container.User
		// dst.Group = container.Group

		if container.Entrypoint != "" {
			dst.Entrypoint = []string{container.Entrypoint}
		}
	}

	// append all matrix parameters as environment
	// variables into the step
	if src.Strategy != nil && src.Strategy.Spec != nil {
		v, ok := src.Strategy.Spec.(*harness.Matrix)
		if ok {
			for _, axis := range v.Include {
				dst.Envs = environ.Combine(dst.Envs, axis)
			}
		}
	}

	// TODO re-enable
	// set container limits
	// if v := int64(src.MemLimit); v > 0 {
	// 	dst.MemLimit = v
	// }
	// if v := int64(src.MemSwapLimit); v > 0 {
	// 	dst.MemSwapLimit = v
	// }

	// appends the volumes to the container def.
	for _, vol := range spec.Mount {
		dst.Volumes = append(dst.Volumes, &engine.VolumeMount{
			Name: vol.Name,
			Path: vol.Path,
		})
	}

	// TODO re-enable
	// // set the pipeline step run policy. steps run on
	// // success by default, but may be optionally configured
	// // to run on failure.
	// if isRunAlways(src) {
	// 	dst.RunPolicy = RunAlways
	// } else if isRunOnFailure(src) {
	// 	dst.RunPolicy = RunOnFailure
	// }

	// TODO re-enable
	// // set the pipeline failure policy. steps can choose
	// // to ignore the failure, or fail fast.
	// switch src.Failure {
	// case "ignore":
	// 	dst.ErrPolicy = ErrIgnore
	// case "fast", "fast-fail", "fail-fast":
	// 	dst.ErrPolicy = ErrFailFast
	// }

	return dst
}
