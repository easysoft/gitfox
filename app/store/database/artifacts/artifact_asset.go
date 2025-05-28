// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"context"
	"time"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ store.ArtifactAssetInterface = (*assets)(nil)

type assets struct {
	db *gorm.DB
}

func (c *assets) Create(ctx context.Context, newObj *types.ArtifactAsset) error {
	// check duplicate for nullable field
	if newObj.VersionID.Ptr() == nil {
		var err error
		if newObj.ViewID.Ptr() == nil {
			// like container blobs shared for all owners
			_, err = c.GetPath(ctx, newObj.Path, newObj.Format)
		} else {
			// like format index files
			_, err = c.GetMetaAsset(ctx, newObj.Path, newObj.ViewID.Int64, newObj.Format)
		}
		// raise duplicate if object found
		if err == nil {
			return gitfox_store.ErrDuplicate
		}
	}

	if err := types.ValidateAssetKind(newObj.Kind); err != nil {
		return err
	}

	if err := dbtx.GetOrmAccessor(ctx, c.db).Create(newObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "exec artifact asset create failed")
	}
	return nil
}

func (c *assets) GetById(ctx context.Context, assetId int64) (*types.ArtifactAsset, error) {
	var asset types.ArtifactAsset
	if err := dbtx.GetOrmAccessor(ctx, c.db).First(&asset, assetId).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec artifact asset query id failed")
	}

	return &asset, nil
}

func (c *assets) find(ctx context.Context, query *types.ArtifactAsset) (*types.ArtifactAsset, error) {
	var asset types.ArtifactAsset
	var err error

	if err = dbtx.GetOrmAccessor(ctx, c.db).Where(query).Where("asset_deleted = 0").First(&asset).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec artifact asset query failed")
	}

	return &asset, nil
}
func (c *assets) GetPath(ctx context.Context, path string, format types.ArtifactFormat) (*types.ArtifactAsset, error) {
	q := types.ArtifactAsset{
		Path: path, Format: format,
	}
	return c.find(ctx, &q)
}

func (c *assets) GetMetaAsset(ctx context.Context, path string, viewId int64, format types.ArtifactFormat) (*types.ArtifactAsset, error) {
	q := types.ArtifactAsset{
		Path: path, Format: format, ViewID: null.IntFrom(viewId),
	}
	return c.find(ctx, &q)
}

func (c *assets) GetVersionAsset(ctx context.Context, path string, versionId int64) (*types.ArtifactAsset, error) {
	q := types.ArtifactAsset{
		Path: path, VersionID: null.IntFrom(versionId),
	}
	return c.find(ctx, &q)
}

func (c *assets) FindMain(ctx context.Context, versionId int64) ([]*types.ArtifactAsset, error) {
	var assetList []*types.ArtifactAsset
	var err error

	if err = validatorEmpty(versionId); err != nil {
		return nil, err
	}

	q := types.ArtifactAsset{
		VersionID: null.IntFrom(versionId),
	}

	if err = dbtx.GetOrmAccessor(ctx, c.db).Where(q).Where("asset_deleted = 0").Find(&assetList).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "find main assets")
	}

	return assetList, nil
}

func (c *assets) UpdateBlobId(ctx context.Context, assetId int64, blobId int64) error {
	if err := validatorEmpty(assetId); err != nil {
		return err
	}

	result := dbtx.GetOrmAccessor(ctx, c.db).Model(&types.ArtifactAsset{ID: assetId}).
		Updates(&types.ArtifactAsset{BlobID: blobId})
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "exec asset update blob failed")
	}

	if result.RowsAffected == 0 {
		return gitfox_store.ErrResourceNotFound
	}

	return nil
}

func (c *assets) Update(ctx context.Context, upObj *types.ArtifactAsset) error {
	err := dbtx.GetOrmAccessor(ctx, c.db).Save(&upObj).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "exec asset update failed")
	}

	return nil
}

// SoftDeleteById soft remove an asset record by primary key
func (c *assets) SoftDeleteById(ctx context.Context, assetId int64) error {
	err := dbtx.GetOrmAccessor(ctx, c.db).Model(&types.ArtifactAsset{ID: assetId}).
		UpdateColumn("asset_deleted", time.Now().UnixMilli()).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "exec delete artifact asset %d failed", assetId)
	}
	return nil
}

// SoftDeleteExcludeById remove all asset records of a version except special one
func (c *assets) SoftDeleteExcludeById(ctx context.Context, versionId int64, assetId int64) error {
	err := dbtx.GetOrmAccessor(ctx, c.db).Model(&types.ArtifactAsset{}).
		Where("asset_version_id = ?", versionId).
		Where("asset_id <> ?", assetId).
		Where("asset_deleted = 0").
		UpdateColumn("asset_deleted", time.Now().UnixMilli()).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "exec artifact delete asset exclude failed")
	}
	return nil
}

func (c *assets) DeleteById(ctx context.Context, assetId int64) error {
	result := dbtx.GetOrmAccessor(ctx, c.db).Delete(&types.ArtifactAsset{}, assetId)
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "exec asset delete failed")
	}

	if result.RowsAffected == 0 {
		return types.ErrAssetNoItemDeleted
	}
	return nil
}

func (c *assets) Search(ctx context.Context, options ...store.SearchOption) ([]*types.ArtifactAsset, error) {
	db := dbtx.GetOrmAccessor(ctx, c.db)
	for _, opt := range options {
		db = opt.Apply(db)
	}

	var data []*types.ArtifactAsset
	result := db.Order("asset_id asc").Find(&data)

	if result.Error != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, result.Error, "exec asset search failed")
	}

	return data, nil
}

func (c *assets) SearchExtendBlob(ctx context.Context, options ...store.SearchOption) ([]*types.ArtifactAssetExtendBlob, error) {
	db := dbtx.GetOrmAccessor(ctx, c.db)
	for _, opt := range options {
		db = opt.Apply(db)
	}

	var data []*types.ArtifactAssetExtendBlob
	result := db.Model(&types.ArtifactAsset{}).
		Select("artifact_assets.*, artifact_blobs.blob_id, artifact_blobs.storage_id, artifact_blobs.blob_ref, artifact_blobs.blob_size").
		Joins("left join artifact_blobs on artifact_blobs.blob_id = artifact_assets.asset_blob_id").
		Order("asset_id asc").Scan(&data)

	if result.Error != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, result.Error, "exec asset search failed")
	}

	return data, nil
}
