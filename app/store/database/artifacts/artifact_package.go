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
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

var _ store.ArtifactPackageInterface = (*packages)(nil)

type packages struct {
	db *gorm.DB
}

func (c *packages) Create(ctx context.Context, newObj *types.ArtifactPackage) error {
	if err := dbtx.GetOrmAccessor(ctx, c.db).Model(new(types.ArtifactPackage)).Create(newObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "failed to create artifact package")
	}
	return nil
}

func (c *packages) GetByID(ctx context.Context, packageId int64) (*types.ArtifactPackage, error) {
	var pkg types.ArtifactPackage
	if err := dbtx.GetOrmAccessor(ctx, c.db).First(&pkg, packageId).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec package query failed")
	}

	return &pkg, nil
}

func (c *packages) GetByName(ctx context.Context, name, namespace string, ownerId int64, format types.ArtifactFormat) (*types.ArtifactPackage, error) {
	var pkg types.ArtifactPackage
	var err error

	q := types.ArtifactPackage{
		Name:      name,
		Namespace: namespace,
		OwnerID:   ownerId,
		Format:    format,
	}

	if err = dbtx.GetOrmAccessor(ctx, c.db).Where(q).Where("package_namespace = ?", namespace).First(&pkg).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec package query failed")
	}

	return &pkg, nil
}

func (c *packages) List(ctx context.Context, ownerId int64, includeDeleted bool) ([]*types.ArtifactPackage, error) {
	q := types.ArtifactPackage{OwnerID: ownerId}
	stmt := dbtx.GetOrmAccessor(ctx, c.db).Where(q)
	if !includeDeleted {
		stmt = stmt.Where("package_deleted = 0")
	}

	var dst []*types.ArtifactPackage
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed executing artifact package list query")
	}
	return dst, nil
}

func (c *packages) ListByNamespace(ctx context.Context, ownerId int64, namespace string, format types.ArtifactFormat, recurse, includeDeleted bool) ([]*types.ArtifactPackage, error) {
	q := types.ArtifactPackage{OwnerID: ownerId}
	stmt := dbtx.GetOrmAccessor(ctx, c.db).Where(q)
	if recurse {
		stmt = stmt.Where("package_namespace LIKE ?", namespace+"%")
	} else {
		stmt = stmt.Where("package_namespace = ?", namespace)
	}

	if format != types.ArtifactAllFormat {
		stmt = stmt.Where("package_format = ?", format)
	}

	if !includeDeleted {
		stmt = stmt.Where("package_deleted = 0")
	}

	var dst []*types.ArtifactPackage
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed executing artifact package list query")
	}
	return dst, nil
}

func (c *packages) Update(ctx context.Context, upObj *types.ArtifactPackage) error {
	err := dbtx.GetOrmAccessor(ctx, c.db).Save(&upObj).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "exec package update failed")
	}

	return nil
}

func (c *packages) SoftDelete(ctx context.Context, pkg *types.ArtifactPackage) error {
	pkg.Deleted = time.Now().UnixMilli()
	return c.Update(ctx, pkg)
}

func (c *packages) UnDelete(ctx context.Context, pkg *types.ArtifactPackage) error {
	pkg.Deleted = 0
	return c.Update(ctx, pkg)
}

func (c *packages) DeleteById(ctx context.Context, packageId int64) error {
	delNum, err := c.deleteByIds(ctx, packageId)
	if err != nil {
		return err
	}

	if delNum == 0 {
		return types.ErrPkgNoItemDeleted
	}
	return nil
}

func (c *packages) DeleteByIds(ctx context.Context, packageIds ...int64) (int64, error) {
	return c.deleteByIds(ctx, packageIds...)
}

func (c *packages) deleteByIds(ctx context.Context, packageIds ...int64) (int64, error) {
	result := dbtx.GetOrmAccessor(ctx, c.db).Delete(&types.ArtifactPackage{}, packageIds)

	if result.Error != nil {
		return result.RowsAffected, database.ProcessGormSQLErrorf(ctx, result.Error, "exec package delete failed")
	}

	return result.RowsAffected, nil
}

func (c *packages) Search(ctx context.Context, options ...store.SearchOption) ([]*types.ArtifactPackage, error) {
	db := dbtx.GetOrmAccessor(ctx, c.db)
	for _, opt := range options {
		db = opt.Apply(db)
	}

	var data []*types.ArtifactPackage
	result := db.Order("package_id asc").Find(&data)

	if result.Error != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, result.Error, "exec package search failed")
	}

	return data, nil
}
