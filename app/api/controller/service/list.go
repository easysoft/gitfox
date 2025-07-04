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

package service

import (
	"context"
	"fmt"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

// List lists all services of the system.
func (c *Controller) List(ctx context.Context, session *auth.Session) (int64, []*types.Service, error) {
	// Ensure principal has required permissions (service is global, no explicit resource)
	scope := &types.Scope{}
	resource := &types.Resource{
		Type: enum.ResourceTypeService,
	}
	if err := apiauth.Check(ctx, c.authorizer, session, scope, resource, enum.PermissionServiceView); err != nil {
		return 0, nil, err
	}

	count, err := c.principalStore.CountServices(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to count services: %w", err)
	}

	repos, err := c.principalStore.ListServices(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to list services: %w", err)
	}

	return count, repos, nil
}
