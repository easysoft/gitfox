// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo

import (
	"context"
	"fmt"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

// AIConfig get ai repository config
func (c *Controller) AIConfig(ctx context.Context, session *auth.Session, repoRef string) (*types.RepositoryMirror, error) {
	repo, err := c.repoStore.FindByRef(ctx, repoRef)
	if err != nil || !repo.Mirror {
		return nil, fmt.Errorf("repo %s failed to find or not a mirror", repoRef)
	}

	if err = apiauth.CheckRepo(ctx, c.authorizer, session, repo, enum.PermissionRepoView); err != nil {
		return nil, err
	}

	return c.repoStore.GetMirror(ctx, repo.ID)
}

// UpdateAIConfig update ai repository config
func (c *Controller) UpdateAIConfig(ctx context.Context, session *auth.Session, repoRef string, opts *types.RepositoryMirror) error {
	repo, err := c.repoStore.FindByRef(ctx, repoRef)
	if err != nil || !repo.Mirror {
		return fmt.Errorf("repo %s failed to find or not a mirror", repoRef)
	}

	if err = apiauth.CheckRepo(ctx, c.authorizer, session, repo, enum.PermissionRepoView); err != nil {
		return err
	}
	opts.RepoID = repo.ID
	return c.repoStore.UpdateMirror(ctx, opts)
}
