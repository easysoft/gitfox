// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package types

import (
	"github.com/guregu/null"
	"gorm.io/gorm"
)

const (
	emptyNamespace = "@"
)

type SearchVersionOption struct {
	PackageId      int64
	ViewId         int64
	Page           int    `json:"page"`
	Size           int    `json:"size"`
	Query          string `json:"query"`
	IncludeDeleted bool   `json:"include_deleted"`
}

func (opt SearchVersionOption) Apply(db *gorm.DB) *gorm.DB {
	q := ArtifactVersion{}
	if opt.PackageId != 0 {
		q.PackageID = opt.PackageId
	}

	if opt.ViewId != 0 {
		q.ViewID = opt.ViewId
	}

	return db.Where(&q)
}

type SearchAssetOption struct {
	VersionId int64
	Kind      AssetKind
	Path      string
}

func (opt SearchAssetOption) Apply(db *gorm.DB) *gorm.DB {
	q := ArtifactAsset{
		VersionID: null.IntFromPtr(&opt.VersionId),
	}

	if opt.Kind != ("") {
		q.Kind = opt.Kind
	}

	return db.Where(&q)
}
