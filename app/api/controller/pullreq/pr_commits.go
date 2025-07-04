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

package pullreq

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox/app/api/controller"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

// Commits lists all commits from pr head branch.
func (c *Controller) Commits(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	pullreqNum int64,
	filter *types.PaginationFilter,
) ([]types.Commit, error) {
	repo, err := c.getRepoCheckAccess(ctx, session, repoRef, enum.PermissionRepoView)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire access to repo: %w", err)
	}

	pr, err := c.pullreqStore.FindByNumber(ctx, repo.ID, pullreqNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request by number: %w", err)
	}

	gitRef := pr.SourceSHA
	afterRef := pr.MergeBaseSHA

	output, err := c.git.ListCommits(ctx, &git.ListCommitsParams{
		ReadParams: git.CreateReadParams(repo),
		GitREF:     gitRef,
		After:      afterRef,
		Page:       int32(filter.Page),
		Limit:      int32(filter.Limit),
	})
	if err != nil {
		return nil, err
	}

	commits := make([]types.Commit, len(output.Commits))
	for i := range output.Commits {
		var commit *types.Commit
		commit, err = controller.MapCommit(&output.Commits[i])
		if err != nil {
			return nil, fmt.Errorf("failed to map commit: %w", err)
		}
		commits[i] = *commit
	}

	return commits, nil
}
