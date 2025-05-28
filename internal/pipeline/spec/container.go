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
	"encoding/json"
)

type Container struct {
	Image       string         `json:"image,omitempty"`
	Connector   string         `json:"connector,omitempty"`
	Credentials *Credentials   `json:"credentials,omitempty"`
	Pull        string         `json:"pull,omitempty"`
	Entrypoint  string         `json:"entrypoint,omitempty"`
	Args        []string       `json:"args,omitempty"`
	Dns         Stringorslice  `json:"dns,omitempty"`
	DsnSearch   Stringorslice  `json:"dsn_search,omitempty"`
	ExtraHosts  Stringorslice  `json:"extra_hosts,omitempty"`
	Network     string         `json:"network,omitempty"`
	NetworkMode string         `json:"network_mode,omitempty"`
	Privileged  bool           `json:"privileged,omitempty"`
	Workdir     string         `json:"workdir,omitempty"`
	Ports       []string       `json:"ports,omitempty"`
	Volumes     []string       `json:"volumes,omitempty"`
	User        string         `json:"user,omitempty"`
	Group       string         `json:"group,omitempty"`
	Cpu         StringorInt    `json:"cpu,omitempty"`
	Memory      MemStringorInt `json:"memory,omitempty"`
	ShmSize     MemStringorInt `json:"shm_size,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Container) UnmarshalJSON(data []byte) error {
	type S Container
	type T struct {
		*S
	}

	var out string
	if err := json.Unmarshal(data, &out); err == nil {
		v.Image = out
		return nil
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	return nil
}
