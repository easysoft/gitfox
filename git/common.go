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

package git

import (
	"github.com/easysoft/gitfox/errors"
)

// ReadParams contains the base parameters for read operations.
type ReadParams struct {
	RepoUID             string
	AlternateObjectDirs []string
}

func (p ReadParams) Validate() error {
	if p.RepoUID == "" {
		return errors.InvalidArgument("repository id cannot be empty")
	}
	return nil
}

// WriteParams contains the base parameters for write operations.
type WriteParams struct {
	RepoUID string
	Actor   Identity
	EnvVars map[string]string
	Mirror  bool
}

func (p *WriteParams) Validate() error {
	if p.RepoUID == "" {
		return errors.InvalidArgument("RepoUID is mandatory field")
	}

	return p.Actor.Validate()
}
