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

// DefineLabelValue defines a new label value for the specified space and label.
func (c *Controller) DefineLabelValue(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	key string,
	in *types.DefineValueInput,
) (*types.LabelValue, error) {
	space, err := c.getSpaceCheckAuth(ctx, session, spaceRef, enum.PermissionRepoEdit)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire access to space: %w", err)
	}

	if err := in.Sanitize(); err != nil {
		return nil, fmt.Errorf("failed to sanitize input: %w", err)
	}

	label, err := c.labelSvc.Find(ctx, &space.ID, nil, key)
	if err != nil {
		return nil, fmt.Errorf("failed to find repo label: %w", err)
	}

	value, err := c.labelSvc.DefineValue(ctx, session.Principal.ID, label.ID, in)
	if err != nil {
		return nil, fmt.Errorf("failed to create space label value: %w", err)
	}

	return value, nil
}
