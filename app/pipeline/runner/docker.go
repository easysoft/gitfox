// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runner

import (
	goruntime "runtime"

	"github.com/easysoft/gitfox/app/pipeline/resolver"
	"github.com/easysoft/gitfox/internal/runner/docker/engine"
	"github.com/easysoft/gitfox/internal/runner/docker/engine/compiler"
	"github.com/easysoft/gitfox/internal/runner/docker/engine/linter"
	"github.com/easysoft/gitfox/internal/runner/docker/engine/resource"
	compiler2 "github.com/easysoft/gitfox/internal/runner/docker/engine2/compiler"
	engine2 "github.com/easysoft/gitfox/internal/runner/docker/engine2/engine"
	runtime2 "github.com/easysoft/gitfox/internal/runner/docker/engine2/runtime"
	"github.com/easysoft/gitfox/internal/runner/extend"
	"github.com/easysoft/gitfox/types"

	dockerclient "github.com/docker/docker/client"
	"github.com/drone/drone-go/drone"
	runnerclient "github.com/drone/runner-go/client"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/pipeline/reporter/history"
	"github.com/drone/runner-go/pipeline/reporter/remote"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/pipeline/uploader"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"
)

// Privileged provides a list of plugins that execute
// with privileged capabilities in order to run Docker
// in Docker.
var Privileged = []string{
	"plugins/docker",
	"plugins/acr",
	"plugins/ecr",
	"plugins/gcr",
	"plugins/heroku",
}

// dockerOpts returns back the options to be overridden from docker options set
// in the environment. If values are specified in gitness, they get preference.
func dockerOpts(config *types.Config) []dockerclient.Opt {
	var overrides []dockerclient.Opt
	if config.Docker.Host != "" {
		overrides = append(overrides, dockerclient.WithHost(config.Docker.Host))
	}
	if config.Docker.APIVersion != "" {
		overrides = append(overrides, dockerclient.WithVersion(config.Docker.APIVersion))
	}
	return overrides
}

func newDockerRunner(
	config *types.Config,
	client runnerclient.Client,
	resolver *resolver.Manager,
	stageEnvProvider *extend.StageDynamicEnvProvider,
) (*runtime2.Runner, error) {
	// For linux/windows, containers need to have extra hosts set in order to interact with
	// gitness. For docker desktop for mac, this is built in and not needed.
	extraHosts := []string{}
	if goruntime.GOOS != "darwin" {
		extraHosts = []string{"host.docker.internal:host-gateway"}
	}

	compiler := &compiler.Compiler{
		Environ:    provider.Static(map[string]string{}),
		Registry:   registry.Static([]*drone.Registry{}),
		Secret:     secret.Encrypted(),
		ExtraHosts: extraHosts,
		Privileged: Privileged,
		Networks:   config.CI.ContainerNetworks,
	}

	remote := remote.New(client)
	upload := uploader.New(client)
	tracer := history.New(remote)
	engine, err := engine.NewEnv(engine.Opts{})
	if err != nil {
		return nil, err
	}

	exec := runtime.NewExecer(tracer, remote, upload,
		engine, int64(config.CI.ParallelWorkers))

	legacyRunner := &runtime.Runner{
		Machine:  config.InstanceID,
		Client:   client,
		Reporter: tracer,
		Lookup:   resource.Lookup,
		Lint:     linter.New().Lint,
		Compiler: compiler,
		Exec:     exec.Exec,
	}

	engine2, err := engine2.NewEnv(engine2.Opts{})
	if err != nil {
		return nil, err
	}

	exec2 := runtime2.NewExecer(tracer, remote, upload, engine2, int64(config.CI.ParallelWorkers))

	compiler2 := &compiler2.CompilerImpl{
		Environ:    provider.Combine(provider.Static(map[string]string{}), stageEnvProvider),
		Registry:   registry.Static([]*drone.Registry{}),
		Secret:     secret.Encrypted(),
		ExtraHosts: extraHosts,
		Privileged: Privileged,
		Networks:   config.CI.ContainerNetworks,
	}

	runner := &runtime2.Runner{
		Machine:      config.InstanceID,
		Client:       client,
		Resolver:     resolver.GetLookupFn(),
		Reporter:     tracer,
		Compiler:     compiler2,
		Exec:         exec2.Exec,
		LegacyRunner: legacyRunner,
	}

	return runner, nil
}
