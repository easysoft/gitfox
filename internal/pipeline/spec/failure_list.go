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

type FailureList struct {
	Items []*Failure `json:"failure,omitempty"`
}

// UnmarshalJSON implements the unmarshal interface.
func (v *FailureList) UnmarshalJSON(data []byte) error {
	// parse the failure clause array.
	if err := json.Unmarshal(data, &v.Items); err == nil {
		return nil
	}

	// parse the simple failure clause.
	vv := new(Failure)
	if err := json.Unmarshal(data, &vv); err == nil {
		v.Items = append(v.Items, vv)
		return nil
	}

	return errors.New("failed to unmarshal failure clause")
}

// MarshalJSON implements the marshal interface.
func (v *FailureList) MarshalJSON() ([]byte, error) {
	if len(v.Items) == 1 {
		return json.Marshal(v.Items[0])
	}
	if len(v.Items) > 1 {
		return json.Marshal(v.Items)
	}
	return []byte("null"), nil
}
