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
	"fmt"

	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types/enum"
)

// SecurityFind returns the security settings of a repo.
func (c *Controller) SecurityFind(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
) (*SecuritySettings, error) {
	repo, err := c.getRepoCheckAccess(ctx, session, repoRef, enum.PermissionRepoView)
	if err != nil {
		return nil, err
	}

	out := GetDefaultSecuritySettings()
	mappings := GetSecuritySettingsMappings(out)
	err = c.settings.RepoMap(ctx, repo.ID, mappings...)
	if err != nil {
		return nil, fmt.Errorf("failed to map settings: %w", err)
	}

	return out, nil
}
