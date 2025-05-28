// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"time"

	"gorm.io/gorm"
)

type AssetWithDeletedOption struct {
}

func (opt AssetWithDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("asset_deleted > 0")
}

type AssetExcludeDeletedOption struct {
}

func (opt AssetExcludeDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("asset_deleted = 0")
}

type AssetWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt AssetWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("asset_deleted > 0").Where("asset_deleted < ?", opt.Before.UnixMilli())
}

type PackageWithDeletedOption struct {
}

func (opt PackageWithDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("package_deleted > 0")
}

type PackageWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt PackageWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("package_deleted > 0").Where("package_deleted < ?", opt.Before.UnixMilli())
}

type VersionWithDeletedOption struct {
}

func (opt VersionWithDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("version_deleted > 0")
}

type VersionWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt VersionWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("version_deleted > 0").Where("version_deleted < ?", opt.Before.UnixMilli())
}
