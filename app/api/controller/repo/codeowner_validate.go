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

package repo

import (
	"context"

	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

func (c *Controller) CodeOwnersValidate(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	ref string,
) (*types.CodeOwnersValidation, error) {
	repo, err := c.getRepoCheckAccess(ctx, session, repoRef, enum.PermissionRepoView)
	if err != nil {
		return nil, err
	}

	if ref == "" {
		ref = repo.DefaultBranch
	}

	violations, err := c.codeOwners.Validate(ctx, repo, ref)
	if err != nil {
		return nil, err
	}

	return violations, nil
}
