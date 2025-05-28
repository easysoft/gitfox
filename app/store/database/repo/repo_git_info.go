// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

var _ store.RepoGitInfoView = (*GitInfoView)(nil)

// NewRepoGitOrmInfoView returns a new GitInfoView.
// It's used by the repository git UID cache.
func NewRepoGitOrmInfoView(db *gorm.DB) *GitInfoView {
	return &GitInfoView{
		db: db,
	}
}

type GitInfoView struct {
	db *gorm.DB
}

type gitInfo struct {
	ID       int64  `gorm:"column:repo_id;primaryKey"`
	ParentID int64  `gorm:"column:repo_parent_id"`
	GitUID   string `gorm:"column:repo_git_uid"`
}

func (s *GitInfoView) Find(ctx context.Context, id int64) (*types.RepositoryGitInfo, error) {
	var info gitInfo

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table("repositories").First(&info, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to scan git uid")
	}

	return &types.RepositoryGitInfo{ID: id, ParentID: info.ParentID, GitUID: info.GitUID}, nil
}
