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
	"fmt"
)

// Config defines the a resource configuration.
type Config struct {
	Version StringorInt `json:"version,omitempty"`
	Kind    string      `json:"kind,omitempty"`
	Type    string      `json:"type,omitempty"`
	Name    string      `json:"name,omitempty"`
	Spec    interface{} `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Config) UnmarshalJSON(data []byte) error {
	type S Config
	type T struct {
		*S
		Spec json.RawMessage `json:"spec"`
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	switch v.Kind {
	case "pipeline":
		v.Spec = new(Pipeline)
	case "template":
		switch v.Type {
		case "stage":
			v.Spec = new(TemplateStage)
		case "step":
			v.Spec = new(TemplateStep)
		default:
			return fmt.Errorf("unknown template type %s", v.Type)
		}
	case "plugin":
		switch v.Type {
		case "stage":
			v.Spec = new(PluginStage)
		case "step":
			v.Spec = new(PluginStep)
		default:
			return fmt.Errorf("unknown plugin type %s", v.Type)
		}
	default:
		return fmt.Errorf("unknown kind %s", v.Kind)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}
