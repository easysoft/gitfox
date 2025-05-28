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

package reposettings

import (
	"context"

	"github.com/easysoft/gitfox/app/api/controller/repo"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/auth/authz"
	eventsrepo "github.com/easysoft/gitfox/app/events/repo"
	"github.com/easysoft/gitfox/app/services/settings"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/audit"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

type Controller struct {
	authorizer   authz.Authorizer
	repoStore    store.RepoStore
	aiStore      store.AIStore
	settings     *settings.Service
	auditService audit.Service
	repoReporter *eventsrepo.Reporter
}

func NewController(
	authorizer authz.Authorizer,
	repoStore store.RepoStore,
	aiStore store.AIStore,
	settings *settings.Service,
	auditService audit.Service,
	repoReporter *eventsrepo.Reporter,
) *Controller {
	return &Controller{
		authorizer:   authorizer,
		repoStore:    repoStore,
		aiStore:      aiStore,
		settings:     settings,
		auditService: auditService,
		repoReporter: repoReporter,
	}
}

// getRepoCheckAccess fetches an active repo (not one that is currently being imported)
// and checks if the current user has permission to access it.
func (c *Controller) getRepoCheckAccess(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	reqPermission enum.Permission,
) (*types.Repository, error) {
	// migrating repositories need to adjust the repo settings during the import, hence expanding the allowedstates.
	return repo.GetRepoCheckAccess(
		ctx,
		c.repoStore,
		c.authorizer,
		session,
		repoRef,
		reqPermission,
		[]enum.RepoState{enum.RepoStateActive, enum.RepoStateMigrateGitPush},
	)
}
