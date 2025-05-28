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
	"github.com/easysoft/gitfox/app/services/settings"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/gotidy/ptr"
)

// AIFind returns the ai settings of a repo.
func (c *Controller) AIFind(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
) (*AISettings, error) {
	repo, err := c.getRepoCheckAccess(ctx, session, repoRef, enum.PermissionRepoView)
	if err != nil {
		return nil, err
	}

	out := GetDefaultAISettings()
	mappings := GetAISettingsMappings(out)

	count, _ := c.aiStore.Count(ctx, repo.ParentID)
	out.SpaceAIProvider = count
	if count == 0 {
		out.AIReviewEnabled = ptr.Bool(settings.DefaultAIReviewEnabled)
		return out, nil
	}
	err = c.settings.RepoMap(ctx, repo.ID, mappings...)
	if err != nil {
		return nil, fmt.Errorf("failed to map settings: %w", err)
	}

	return out, nil
}
