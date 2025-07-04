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

package codeowners

import (
	"github.com/easysoft/gitfox/app/services/usergroup"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/git"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideCodeOwners,
)

func ProvideCodeOwners(
	git git.Interface,
	repoStore store.RepoStore,
	config Config,
	principalStore store.PrincipalStore,
	userGroupResolver usergroup.Resolver,
) *Service {
	return New(repoStore, git, config, principalStore, userGroupResolver)
}
