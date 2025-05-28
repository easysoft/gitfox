// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

var _ store.ArtifactVersionInterface = (*versions)(nil)

type versions struct {
	db *gorm.DB
}

func (c *versions) Create(ctx context.Context, newObj *types.ArtifactVersion) error {
	if err := dbtx.GetOrmAccessor(ctx, c.db).Model(new(types.ArtifactVersion)).Create(newObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "failed to create artifact version")
	}
	return nil
}

func (c *versions) GetByID(ctx context.Context, versionId int64) (*types.ArtifactVersion, error) {
	var ver types.ArtifactVersion
	if err := dbtx.GetOrmAccessor(ctx, c.db).First(&ver, versionId).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec package version query id failed")
	}

	return &ver, nil
}

func (c *versions) GetByVersion(ctx context.Context, packageId, viewId int64, version string) (*types.ArtifactVersion, error) {
	var ver types.ArtifactVersion
	var err error

	q := types.ArtifactVersion{
		PackageID: packageId,
		ViewID:    viewId,
		Version:   version,
	}

	if err = dbtx.GetOrmAccessor(ctx, c.db).Where(q).First(&ver).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec package version query path failed")
	}

	return &ver, nil
}

func (c *versions) Update(ctx context.Context, ver *types.ArtifactVersion) error {
	if err := dbtx.GetOrmAccessor(ctx, c.db).Save(ver).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "save updated version failed")
	}

	return nil
}

func (c *versions) DeleteById(ctx context.Context, versionId int64) error {
	delNum, err := c.deleteByIds(ctx, versionId)
	if err != nil {
		return err
	}

	if delNum == 0 {
		return types.ErrPkgVersionNoItemDeleted
	}
	return nil
}

func (c *versions) DeleteByIds(ctx context.Context, versionIds ...int64) (int64, error) {
	return c.deleteByIds(ctx, versionIds...)
}

func (c *versions) deleteByIds(ctx context.Context, versionIds ...int64) (int64, error) {
	result := dbtx.GetOrmAccessor(ctx, c.db).Delete(&types.ArtifactVersion{}, versionIds)

	if result.Error != nil {
		return result.RowsAffected, database.ProcessGormSQLErrorf(ctx, result.Error, "exec package version delete failed")
	}

	return result.RowsAffected, nil
}

func (c *versions) Find(ctx context.Context, opt types.SearchVersionOption) ([]*types.ArtifactVersion, error) {
	stmt := opt.Apply(dbtx.GetOrmAccessor(ctx, c.db))

	if opt.Query != "" {
		stmt = stmt.Where("version LIKE ?", fmt.Sprintf("%%%s%%", opt.Query))
	}

	if !opt.IncludeDeleted {
		stmt = stmt.Where("version_deleted = 0")
	}

	stmt = stmt.Limit(int(database.Limit(opt.Size)))
	stmt = stmt.Offset(int(database.Offset(opt.Page, opt.Size)))

	var data []*types.ArtifactVersion
	result := stmt.Order("version_updated desc").Find(&data)

	if result.Error != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, result.Error, "exec package version search failed")
	}

	return data, nil
}

func (c *versions) SoftDelete(ctx context.Context, ver *types.ArtifactVersion) error {
	ver.Deleted = time.Now().UnixMilli()
	return c.Update(ctx, ver)
}

func (c *versions) UnDelete(ctx context.Context, ver *types.ArtifactVersion) error {
	ver.Deleted = 0
	return c.Update(ctx, ver)
}

func (c *versions) Search(ctx context.Context, options ...store.SearchOption) ([]*types.ArtifactVersionInfo, error) {
	db := dbtx.GetOrmAccessor(ctx, c.db)
	for _, opt := range options {
		db = opt.Apply(db)
	}

	selectFields := []string{
		"artifact_versions.version_id", "artifact_versions.version", "artifact_versions.version_view_id", "artifact_versions.version_deleted",
		"artifact_packages.package_id", "artifact_packages.package_name",
		"artifact_packages.package_namespace", "artifact_packages.package_format",
		"spaces.space_uid",
	}

	var data []*types.ArtifactVersionInfo
	result := db.Model(&types.ArtifactVersion{}).
		Select(strings.Join(selectFields, ",")).
		Joins("left join artifact_packages on artifact_packages.package_id = artifact_versions.version_package_id").
		Joins("left join artifact_views on artifact_views.view_id = artifact_versions.version_view_id").
		Joins("left join spaces on spaces.space_id = artifact_views.view_space_id").
		Scan(&data)

	if result.Error != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, result.Error, "exec version info search failed")
	}

	return data, nil
}
