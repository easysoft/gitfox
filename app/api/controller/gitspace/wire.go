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

package gitspace

import (
	"github.com/easysoft/gitfox/app/auth/authz"
	gitspaceevents "github.com/easysoft/gitfox/app/events/gitspace"
	"github.com/easysoft/gitfox/app/gitspace/logutil"
	"github.com/easysoft/gitfox/app/gitspace/orchestrator"
	"github.com/easysoft/gitfox/app/gitspace/scm"
	"github.com/easysoft/gitfox/app/services/gitspace"
	"github.com/easysoft/gitfox/app/services/infraprovider"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database/dbtx"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideController,
)

func ProvideController(
	tx dbtx.Transactor,
	authorizer authz.Authorizer,
	infraProviderSvc *infraprovider.Service,
	configStore store.GitspaceConfigStore,
	instanceStore store.GitspaceInstanceStore,
	spaceStore store.SpaceStore,
	reporter *gitspaceevents.Reporter,
	orchestrator orchestrator.Orchestrator,
	eventStore store.GitspaceEventStore,
	statefulLogger *logutil.StatefulLogger,
	scm scm.SCM,
	repoStore store.RepoStore,
	gitspaceSvc *gitspace.Service,
) *Controller {
	return NewController(
		tx,
		authorizer,
		infraProviderSvc,
		configStore,
		instanceStore,
		spaceStore,
		reporter,
		orchestrator,
		eventStore,
		statefulLogger,
		scm,
		repoStore,
		gitspaceSvc,
	)
}
