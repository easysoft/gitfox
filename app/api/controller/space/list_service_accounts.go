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

package space

import (
	"context"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

/*
* ListServiceAccounts lists the service accounts of a space.
 */
func (c *Controller) ListServiceAccounts(ctx context.Context, session *auth.Session,
	spaceRef string) ([]*types.ServiceAccount, error) {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return nil, err
	}

	if err = apiauth.CheckServiceAccount(
		ctx,
		c.authorizer,
		session,
		c.spaceStore,
		c.repoStore,
		enum.ParentResourceTypeSpace,
		space.ID,
		"",
		enum.PermissionServiceAccountView,
	); err != nil {
		return nil, err
	}

	return c.principalStore.ListServiceAccounts(ctx, enum.ParentResourceTypeSpace, space.ID)
}
