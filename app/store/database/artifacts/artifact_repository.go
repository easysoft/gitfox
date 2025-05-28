// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ store.ArtifactRepositoryInterface = (*repositories)(nil)

type repositories struct {
	db *gorm.DB
}

func (s *repositories) Create(ctx context.Context, newObj *types.ArtifactRepository) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Create(newObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "artifact repository create failed")
	}
	return nil
}

func (s *repositories) GetById(ctx context.Context, repoId int64) (*types.ArtifactRepository, error) {
	var repository types.ArtifactRepository
	if err := dbtx.GetOrmAccessor(ctx, s.db).First(&repository, repoId).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "query artifact repository by id failed")
	}

	return &repository, nil
}

func (s *repositories) GetByIdentifier(ctx context.Context, Identifier string, kind enum.ArtifactRepoKind) (*types.ArtifactRepository, error) {
	var repository types.ArtifactRepository
	q := types.ArtifactRepository{Identifier: Identifier, Kind: kind}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Where(&q).First(&repository).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "query artifact repository by name failed")
	}

	return &repository, nil
}

func (s *repositories) GetByRefID(ctx context.Context, refId int64, kind enum.ArtifactRepoKind) (*types.ArtifactRepository, error) {
	var repository types.ArtifactRepository
	q := types.ArtifactRepository{RefID: null.IntFrom(refId), Kind: kind}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Where(&q).First(&repository).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "query artifact repository by name failed")
	}

	return &repository, nil
}

func (s *repositories) Update(ctx context.Context, upObj *types.ArtifactRepository) error {
	err := dbtx.GetOrmAccessor(ctx, s.db).Save(&upObj).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "exec artifact repository update failed")
	}

	return nil
}

func (s *repositories) DeleteById(ctx context.Context, repoId int64) error {
	result := dbtx.GetOrmAccessor(ctx, s.db).Delete(&types.ArtifactRepository{}, repoId)
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "exec artifact repository delete failed")
	}

	if result.RowsAffected == 0 {
		return types.ErrAssetNoItemDeleted
	}
	return nil
}
