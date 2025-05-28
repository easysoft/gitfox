// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"context"
	"errors"
	"time"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

var _ store.ArtifactBlobInterface = (*blobs)(nil)

type blobs struct {
	db *gorm.DB
}

func (c *blobs) Create(ctx context.Context, newObj *types.ArtifactBlob) error {
	if err := validatorEmpty(newObj.StorageID, newObj.Ref); err != nil {
		return err
	}

	if err := dbtx.GetOrmAccessor(ctx, c.db).Create(newObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "exec artifact blob create failed")
	}
	return nil
}

func (c *blobs) GetById(ctx context.Context, blobId int64) (*types.ArtifactBlob, error) {
	var blob types.ArtifactBlob
	if err := dbtx.GetOrmAccessor(ctx, c.db).First(&blob, blobId).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "artifact blob find by id failed")
	}

	return &blob, nil
}

func (c *blobs) GetByRef(ctx context.Context, ref string, storageId int64) (*types.ArtifactBlob, error) {
	var blob types.ArtifactBlob
	if err := validatorEmpty(ref, storageId); err != nil {
		return nil, types.ErrArgsValueEmpty
	}

	q := types.ArtifactBlob{
		Ref: ref, StorageID: storageId,
	}

	if err := dbtx.GetOrmAccessor(ctx, c.db).Where(q).First(&blob).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "blob find by ref failed")
	}

	return &blob, nil
}

func (c *blobs) Update(ctx context.Context, upObj *types.ArtifactBlob) error {
	if err := dbtx.GetOrmAccessor(ctx, c.db).Save(upObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "update blob %s failed", upObj.Ref)
	}
	return nil
}

// UpdateOptLock updates the blob using the optimistic locking mechanism.
func (c *blobs) UpdateOptLock(
	ctx context.Context,
	upObj *types.ArtifactBlob,
	mutateFn func(blob *types.ArtifactBlob) error,
) (*types.ArtifactBlob, error) {
	for {
		dup := *upObj

		err := mutateFn(&dup)
		if err != nil {
			return nil, err
		}

		err = c.Update(ctx, &dup)
		if err == nil {
			return &dup, nil
		}
		if !errors.Is(err, gitfox_store.ErrVersionConflict) {
			return nil, err
		}

		upObj, err = c.GetById(ctx, upObj.ID)
		if err != nil {
			return nil, err
		}
	}
}

// SoftDeleteById soft remove a blob record by primary key
func (c *blobs) SoftDeleteById(ctx context.Context, blobId int64) error {
	err := dbtx.GetOrmAccessor(ctx, c.db).Model(&types.ArtifactBlob{ID: blobId}).
		UpdateColumn("blob_deleted", time.Now().UnixMilli()).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "exec blob soft delete failed")
	}
	return nil
}

func (c *blobs) DeleteById(ctx context.Context, blobId int64) error {
	result := dbtx.GetOrmAccessor(ctx, c.db).Delete(&types.ArtifactBlob{}, blobId)
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "exec blob delete failed")
	}

	if result.RowsAffected == 0 {
		return types.ErrBlobNoItemDeleted
	}
	return nil
}
