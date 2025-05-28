// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo

import (
	"context"
	"fmt"
	"time"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
)

type HardDeleteResponse struct {
	DeletedAt int64 `json:"deleted_at"`
}

// HardDelete exactly deletes a repo
func (c *Controller) HardDelete(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
) (*HardDeleteResponse, error) {
	// note: can't use c.getRepoCheckAccess because import job for repositories being imported must be cancelled.
	repo, err := c.repoStore.FindByRef(ctx, repoRef)
	if err != nil {
		return nil, fmt.Errorf("failed to find the repo for soft delete: %w", err)
	}

	if err = apiauth.CheckRepo(ctx, c.authorizer, session, repo, enum.PermissionRepoDelete); err != nil {
		return nil, fmt.Errorf("access check failed: %w", err)
	}

	log.Ctx(ctx).Info().
		Int64("repo.id", repo.ID).
		Str("repo.path", repo.Path).
		Msg("exactly deleting repository")

	if repo.State == enum.RepoStateGitImport {
		log.Ctx(ctx).Info().Msg("repository is importing. cancelling the import job and purge the repo.")
		err = c.importer.Cancel(ctx, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to cancel repository import")
		}
		return nil, c.PurgeNoAuth(ctx, session, repo)
	}

	now := time.Now().UnixMilli()
	if repo.Deleted == nil {
		err = c.SoftDeleteNoAuth(ctx, session, repo, now)
		if err != nil {
			return nil, usererror.BadRequest("repository soft deletion is failed")
		}
		repo.Deleted = &now
	}

	err = c.PurgeNoAuth(ctx, session, repo)
	if err != nil {
		return nil, usererror.BadRequest("repository purge is failed")
	}

	return &HardDeleteResponse{DeletedAt: now}, nil
}
