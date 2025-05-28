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

package yaml

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
)

func TestCredentials(t *testing.T) {
	tests := []struct {
		yaml string
		want Registry
	}{
		// single connector, string
		{
			yaml: `{ "connector": "account.docker" }`,
			want: Registry{
				Connector: []*RegistryConnector{
					{Name: "account.docker"},
				},
			},
		},
		// single connector, struct
		{
			yaml: `{ "connector": { "name": "account.docker", "match": "docker.io" } }`,
			want: Registry{
				Connector: []*RegistryConnector{
					{Name: "account.docker", Match: "docker.io"},
				},
			},
		},
		// multiple connectors, string slice
		{
			yaml: `{ "connector": [ "account.docker", "account.gcr" ] }`,
			want: Registry{
				Connector: []*RegistryConnector{
					{Name: "account.docker"},
					{Name: "account.gcr"},
				},
			},
		},
		// multiple connectors, mixed string slice and structs
		{
			yaml: `{ "connector": [ "account.docker", { "name": "account.gcr", "match": "us.gcr.io" } ] }`,
			want: Registry{
				Connector: []*RegistryConnector{
					{Name: "account.docker"},
					{Name: "account.gcr", Match: "us.gcr.io"},
				},
			},
		},
	}

	for i, test := range tests {
		got := new(Registry)
		if err := yaml.Unmarshal([]byte(test.yaml), got); err != nil {
			t.Error(err)
			return
		}
		if diff := cmp.Diff(got, &test.want); diff != "" {
			t.Errorf("Unexpected parsing results for test %v", i)
			t.Log(diff)
		}
	}
}
