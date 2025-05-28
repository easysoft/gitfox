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
	"path/filepath"
	"testing"

	schema "github.com/easysoft/gitfox/internal/pipeline/spec"
	"github.com/easysoft/gitfox/internal/pipeline/spec/parse"

	"github.com/google/go-cmp/cmp"
)

func TestResolver(t *testing.T) {
	lookup := func(name, kind, typ, version string) (*schema.Config, error) {
		return parse.ParseFile(
			filepath.Join("testdata/resources", name+".yaml"),
		)
	}

	// parse the pipeline yaml
	got, err := parse.ParseFile("testdata/step_template.yaml")
	if err != nil {
		t.Log("cannot parse pipeline")
		t.Error(err)
		return
	}

	if err := Resolve(got, lookup); err != nil {
		t.Error(err)
		return
	}

	want, err := parse.ParseFile("testdata/step_template.yaml.golden")
	if err != nil {
		t.Log("cannot parse golden file")
		t.Error(err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Log(diff)
		t.Fail()
	}
}
