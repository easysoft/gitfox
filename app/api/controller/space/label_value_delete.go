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
	"github.com/easysoft/gitfox/types/enum"
)

// DeleteLabelValue deletes a label value for the specified space.
func (c *Controller) DeleteLabelValue(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	key string,
	value string,
) error {
	space, err := c.getSpaceCheckAuth(ctx, session, spaceRef, enum.PermissionRepoEdit)
	if err != nil {
		return fmt.Errorf("failed to acquire access to space: %w", err)
	}

	if err := c.labelSvc.DeleteValue(ctx, &space.ID, nil, key, value); err != nil {
		return fmt.Errorf("failed to delete space label value: %w", err)
	}

	return nil
}
