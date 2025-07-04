// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"testing"

	"github.com/easysoft/gitfox/internal/runner/docker/engine"
	"github.com/easysoft/gitfox/internal/runner/docker/engine/resource"
	"github.com/easysoft/gitfox/pkg/util/common"

	"github.com/dchest/uniuri"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestClone(t *testing.T) {
	random = notRandom
	defer func() {
		random = uniuri.New
	}()

	c := &Compiler{
		Registry: registry.Static(nil),
		Secret:   secret.Static(nil),
		Environ:  provider.Static(nil),
	}
	args := runtime.CompilerArgs{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{},
		Manifest: &manifest.Manifest{},
		Pipeline: &resource.Pipeline{},
	}
	want := []*engine.Step{
		{
			ID:         "random",
			Image:      common.GitImage(),
			Name:       "clone",
			Pull:       engine.PullIfNotExists,
			RunPolicy:  runtime.RunAlways,
			WorkingDir: "/drone/src",
			Volumes: []*engine.VolumeMount{
				{
					Name: "_workspace",
					Path: "/drone/src",
				},
			},
		},
	}
	got := c.Compile(nocontext, args).(*engine.Spec)
	ignore := cmpopts.IgnoreFields(engine.Step{}, "Envs", "Labels")
	if diff := cmp.Diff(got.Steps, want, ignore); len(diff) != 0 {
		t.Errorf(diff)
	}
}

func TestCloneDisable(t *testing.T) {
	c := &Compiler{
		Environ:  provider.Static(nil),
		Registry: registry.Static(nil),
		Secret:   secret.Static(nil),
	}
	args := runtime.CompilerArgs{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{},
		Manifest: &manifest.Manifest{},
		Pipeline: &resource.Pipeline{Clone: manifest.Clone{Disable: true}},
	}
	got := c.Compile(nocontext, args).(*engine.Spec)
	if len(got.Steps) != 0 {
		t.Errorf("Expect no clone step added when disabled")
	}
}

func TestCloneCreate(t *testing.T) {
	want := &engine.Step{
		Name:      "clone",
		Image:     common.GitImage(),
		RunPolicy: runtime.RunAlways,
		Envs:      map[string]string{"PLUGIN_DEPTH": "50"},
	}
	src := &resource.Pipeline{Clone: manifest.Clone{Depth: 50}}
	got := createClone(src)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

func TestCloneImage(t *testing.T) {
	tests := []struct {
		in  manifest.Platform
		out string
	}{
		{
			in:  manifest.Platform{},
			out: common.GitImage(),
		},
		{
			in:  manifest.Platform{OS: "linux"},
			out: common.GitImage(),
		},
		{
			in:  manifest.Platform{OS: "windows"},
			out: common.GitImage(),
		},
	}
	for _, test := range tests {
		got, want := cloneImage(test.in), test.out
		if got != want {
			t.Errorf("Want clone image %q, got %q", want, got)
		}
	}
}

func TestCloneParams(t *testing.T) {
	params := cloneParams(manifest.Clone{})
	if len(params) != 0 {
		t.Errorf("Expect empty clone parameters")
	}
	params = cloneParams(manifest.Clone{Depth: 0})
	if len(params) != 0 {
		t.Errorf("Expect zero depth ignored")
	}
	params = cloneParams(manifest.Clone{Retries: 0})
	if len(params) != 0 {
		t.Errorf("Expect zero retries ignored")
	}
	params = cloneParams(manifest.Clone{Depth: 50, SkipVerify: true, Retries: 4})
	if params["PLUGIN_DEPTH"] != "50" {
		t.Errorf("Expect clone depth 50")
	}
	if params["PLUGIN_RETRIES"] != "4" {
		t.Errorf("Expect clone retries 4")
	}
	if params["GIT_SSL_NO_VERIFY"] != "true" {
		t.Errorf("Expect GIT_SSL_NO_VERIFY is true")
	}
	if params["PLUGIN_SKIP_VERIFY"] != "true" {
		t.Errorf("Expect PLUGIN_SKIP_VERIFY is true")
	}
}
