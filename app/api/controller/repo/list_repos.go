// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"
)

// ListRepositories lists the repositories of a space.
func (c *Controller) ListRepositories(
	ctx context.Context,
	session *auth.Session,
	filter *types.RepoFilter,
) ([]*types.Repository, int64, error) {
	var err error
	var count int64
	var repos []*types.Repository

	if session.User.Admin {
		repos, count, err = c.listReposForAdmin(ctx, filter)
	} else {
		repos, count, err = c.listReposForUser(ctx, session.User, filter)
	}

	if err != nil {
		return nil, 0, err
	}

	// backfill URLs
	for _, repo := range repos {
		repo.GitURL = c.urlProvider.GenerateGITCloneURL(ctx, repo.Path)
	}

	return repos, count, nil
}

func (c *Controller) listReposForAdmin(ctx context.Context, filter *types.RepoFilter) ([]*types.Repository, int64, error) {
	count, err := c.repoStore.CountAllWithoutParent(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count child repos: %w", err)
	}

	repos, err := c.repoStore.ListAllWithoutParent(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all repos: %w", err)
	}
	return repos, count, nil
}

func (c *Controller) listReposForUser(ctx context.Context, user *types.Principal, filter *types.RepoFilter) ([]*types.Repository, int64, error) {
	spaces, err := c.memberShipStore.ListSpaces(ctx, user.ID, types.MembershipSpaceFilter{})
	if err != nil {
		return nil, 0, err
	}

	if len(spaces) == 0 {
		return []*types.Repository{}, 0, nil
	}

	parentIds := make([]int64, 0)
	for _, space := range spaces {
		parentIds = append(parentIds, space.SpaceID)
	}

	count, err := c.repoStore.CountMulti(ctx, parentIds, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count child repos: %w", err)
	}

	repos, err := c.repoStore.ListMulti(ctx, parentIds, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all repos: %w", err)
	}
	return repos, count, nil
}
