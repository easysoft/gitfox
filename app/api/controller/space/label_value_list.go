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
	"fmt"

	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

// ListLabelValues lists all label values defined in the specified space.
func (c *Controller) ListLabelValues(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	key string,
	filter *types.ListQueryFilter,
) ([]*types.LabelValue, error) {
	space, err := c.getSpaceCheckAuth(ctx, session, spaceRef, enum.PermissionRepoView)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire access to space: %w", err)
	}

	values, err := c.labelSvc.ListValues(ctx, &space.ID, nil, key, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list space label values: %w", err)
	}

	return values, nil
}
