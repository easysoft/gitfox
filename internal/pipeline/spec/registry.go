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
	"errors"
)

// Registry provides container registry details.
type Registry struct {
	Connector []*RegistryConnector `json:"connector,omitempty"`
	Mirror    Stringorslice        `json:"mirror,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Registry) UnmarshalJSON(data []byte) error {
	var out1 = struct {
		Connector json.RawMessage `json:"connector,omitempty"`
	}{}
	// unmarshal into a temporary struct
	if err := json.Unmarshal(data, &out1); err != nil {
		return err
	}

	var out2 string
	var out3 *RegistryConnector
	var out4 []*RegistryConnector

	if err := json.Unmarshal(out1.Connector, &out2); err == nil {
		v.Connector = append(v.Connector, &RegistryConnector{Name: out2})
		return nil
	}

	if err := json.Unmarshal(out1.Connector, &out3); err == nil {
		v.Connector = append(v.Connector, out3)
		return nil
	}

	if err := json.Unmarshal(out1.Connector, &out4); err == nil {
		v.Connector = append(v.Connector, out4...)
		return nil
	}

	return errors.New("yaml: cannot unmarshal registry connector")
}

// RegistryConnector provides registry credentials.
type RegistryConnector struct {
	Name  string `json:"name,omitempty"`
	Match string `json:"match,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *RegistryConnector) UnmarshalJSON(data []byte) error {
	var out1 string
	var out2 = struct {
		Name  string `json:"name,omitempty"`
		Match string `json:"match,omitempty"`
	}{}

	if err := json.Unmarshal(data, &out1); err == nil {
		v.Name = out1
		return nil
	}

	if err := json.Unmarshal(data, &out2); err == nil {
		v.Name = out2.Name
		v.Match = out2.Match
		return nil
	}

	return errors.New("yaml: cannot unmarshal registry connector")
}
