// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package gc

import (
	"time"

	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

type TagAssetOption struct {
}

func (opt TagAssetOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("asset_version_id IS NOT NULL").Where("asset_format = ?", types.ArtifactContainerFormat)
}

type ContainerFormatOption struct {
}

func (opt ContainerFormatOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("asset_format = ?", types.ArtifactContainerFormat)
}

// TagUndeletedAndRecentDeletedOption return
// - asset_deleted = 0
// - asset_deleted > now - x
type TagUndeletedAndRecentDeletedOption struct {
	After time.Time
}

func (opt TagUndeletedAndRecentDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("asset_deleted > ?", opt.After.UnixMilli()).Or("asset_deleted = 0")
}
