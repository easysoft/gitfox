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

package check

import (
	"context"
	"fmt"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

type Controller struct {
	tx         dbtx.Transactor
	authorizer authz.Authorizer
	repoStore  store.RepoStore
	checkStore store.CheckStore
	git        git.Interface
	sanitizers map[enum.CheckPayloadKind]func(in *ReportInput, s *auth.Session) error
}

func NewController(
	tx dbtx.Transactor,
	authorizer authz.Authorizer,
	repoStore store.RepoStore,
	checkStore store.CheckStore,
	git git.Interface,
	sanitizers map[enum.CheckPayloadKind]func(in *ReportInput, s *auth.Session) error,
) *Controller {
	return &Controller{
		tx:         tx,
		authorizer: authorizer,
		repoStore:  repoStore,
		checkStore: checkStore,
		git:        git,
		sanitizers: sanitizers,
	}
}

func (c *Controller) getRepoCheckAccess(ctx context.Context,
	session *auth.Session, repoRef string, reqPermission enum.Permission,
) (*types.Repository, error) {
	if repoRef == "" {
		return nil, usererror.BadRequest("A valid repository reference must be provided.")
	}

	repo, err := c.repoStore.FindByRef(ctx, repoRef)
	if err != nil {
		return nil, fmt.Errorf("failed to find repository: %w", err)
	}

	if err = apiauth.CheckRepo(ctx, c.authorizer, session, repo, reqPermission); err != nil {
		return nil, fmt.Errorf("access check failed: %w", err)
	}

	return repo, nil
}
