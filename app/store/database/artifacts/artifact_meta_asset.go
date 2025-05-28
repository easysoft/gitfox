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

	"gorm.io/gorm"
)

// AssetMeta
var _ store.ArtifactMetaAssetInterface = (*metaAssets)(nil)

type metaAssets struct {
	db *gorm.DB
}

func (s *metaAssets) Create(ctx context.Context, newObj *types.ArtifactMetaAsset) error {
	if err := validatorEmpty(newObj.OwnerID, newObj.Path, newObj.ViewID, newObj.Format); err != nil {
		return err
	}

	if err := types.ValidateAssetKind(newObj.Kind); err != nil {
		return err
	}

	if err := dbtx.GetOrmAccessor(ctx, s.db).Model(new(types.ArtifactMetaAsset)).Create(newObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "insert artifact asset meta failed")
	}
	return nil
}

func (s *metaAssets) GetById(ctx context.Context, metaAssetId int64) (*types.ArtifactMetaAsset, error) {
	var metaAsset types.ArtifactMetaAsset
	if err := dbtx.GetOrmAccessor(ctx, s.db).First(&metaAsset, metaAssetId).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "query artifact meta asset by id failed")
	}

	return &metaAsset, nil
}

func (s *metaAssets) GetByPath(ctx context.Context, path string, viewId int64, format types.ArtifactFormat) (*types.ArtifactMetaAsset, error) {
	var metaAsset types.ArtifactMetaAsset
	var err error

	if err = validatorEmpty(path, viewId, format); err != nil {
		return nil, err
	}

	q := types.ArtifactMetaAsset{
		Path:   path,
		ViewID: viewId,
		Format: format,
	}

	if err = dbtx.GetOrmAccessor(ctx, s.db).Where(q).First(&metaAsset).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "query artifact meta asset by path failed")
	}

	return &metaAsset, nil
}

func (s *metaAssets) UpdateBlobId(ctx context.Context, metaAssetId, blobId int64) error {
	if err := validatorEmpty(blobId); err != nil {
		return err
	}

	result := dbtx.GetOrmAccessor(ctx, s.db).Model(&types.ArtifactMetaAsset{ID: metaAssetId}).
		Updates(&types.ArtifactMetaAsset{BlobID: blobId})
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "exec meta asset update blob failed")
	}

	if result.RowsAffected == 0 {
		return types.ErrMetaAssetNotFound
	}

	return nil
}

func (s *metaAssets) Update(ctx context.Context, metaAsset *types.ArtifactMetaAsset) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Save(metaAsset).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "save updated meta asset failed")
	}

	return nil
}

func (s *metaAssets) DeleteById(ctx context.Context, metaAssetId int64) error {
	result := dbtx.GetOrmAccessor(ctx, s.db).Delete(&types.ArtifactMetaAsset{}, metaAssetId)
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "exec meta asset delete failed")
	}

	if result.RowsAffected == 0 {
		return types.ErrMetaAssetNoItemDeleted
	}
	return nil
}
