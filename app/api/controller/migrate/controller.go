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

package migrate

import (
	"context"
	"fmt"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/api/controller/limiter"
	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/services/migrate"
	"github.com/easysoft/gitfox/app/services/publicaccess"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/audit"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/check"
	"github.com/easysoft/gitfox/types/enum"
)

type Controller struct {
	authorizer      authz.Authorizer
	publicAccess    publicaccess.Service
	git             git.Interface
	urlProvider     url.Provider
	pullreqImporter *migrate.PullReq
	ruleImporter    *migrate.Rule
	webhookImporter *migrate.Webhook
	labelImporter   *migrate.Label
	resourceLimiter limiter.ResourceLimiter
	auditService    audit.Service
	identifierCheck check.RepoIdentifier
	tx              dbtx.Transactor
	spaceStore      store.SpaceStore
	repoStore       store.RepoStore
}

func NewController(
	authorizer authz.Authorizer,
	publicAccess publicaccess.Service,
	git git.Interface,
	urlProvider url.Provider,
	pullreqImporter *migrate.PullReq,
	ruleImporter *migrate.Rule,
	webhookImporter *migrate.Webhook,
	labelImporter *migrate.Label,
	resourceLimiter limiter.ResourceLimiter,
	auditService audit.Service,
	identifierCheck check.RepoIdentifier,
	tx dbtx.Transactor,
	spaceStore store.SpaceStore,
	repoStore store.RepoStore,
) *Controller {
	return &Controller{
		authorizer:      authorizer,
		publicAccess:    publicAccess,
		git:             git,
		urlProvider:     urlProvider,
		pullreqImporter: pullreqImporter,
		ruleImporter:    ruleImporter,
		webhookImporter: webhookImporter,
		labelImporter:   labelImporter,
		resourceLimiter: resourceLimiter,
		auditService:    auditService,
		identifierCheck: identifierCheck,
		tx:              tx,
		spaceStore:      spaceStore,
		repoStore:       repoStore,
	}
}

func (c *Controller) getRepoCheckAccess(ctx context.Context,
	session *auth.Session, repoRef string, reqPermission enum.Permission) (*types.Repository, error) {
	if repoRef == "" {
		return nil, usererror.BadRequest("A valid repository reference must be provided.")
	}

	repo, err := c.repoStore.FindByRef(ctx, repoRef)
	if err != nil {
		return nil, fmt.Errorf("failed to find repo: %w", err)
	}

	if err = apiauth.CheckRepo(ctx, c.authorizer, session, repo, reqPermission); err != nil {
		return nil, fmt.Errorf("failed to verify authorization: %w", err)
	}

	return repo, nil
}

func (c *Controller) getSpaceCheckAccess(
	ctx context.Context,
	session *auth.Session,
	parentRef string,
	reqPermission enum.Permission,
) (*types.Space, error) {
	space, err := c.spaceStore.FindByRef(ctx, parentRef)
	if err != nil {
		return nil, fmt.Errorf("parent space not found: %w", err)
	}

	err = apiauth.CheckSpaceScope(
		ctx,
		c.authorizer,
		session,
		space,
		enum.ResourceTypeSpace,
		reqPermission,
	)
	if err != nil {
		return nil, fmt.Errorf("auth check failed: %w", err)
	}

	return space, nil
}
