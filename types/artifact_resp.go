// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package types

import (
	"fmt"

	"github.com/easysoft/gitfox/types/enum"
)

type ArtifactListItem struct {
	Format      string `gorm:"column:package_format" json:"format"`
	Name        string `gorm:"column:package_name" json:"name"`
	Namespace   string `gorm:"column:package_namespace" json:"namespace"`
	DisplayName string `gorm:"-" json:"display_name"`
	Version     string `gorm:"column:version" json:"version"`
	UpdateTime  int64  `gorm:"column:version_updated" json:"update_time"`
}

// ArtifactFilter define artifacts query parameters.
type ArtifactFilter struct {
	Page   int            `json:"page"`
	Size   int            `json:"size"`
	Query  string         `json:"query"`
	Format string         `json:"format"`
	Sort   enum.SpaceAttr `json:"sort"`
	Order  enum.Order     `json:"order"`
}

// ArtifactVersionFilter define artifact versions query parameters.
type ArtifactVersionFilter struct {
	Page    int    `json:"page"`
	Size    int    `json:"size"`
	Query   string `json:"query"`
	Package string `json:"package"`
	Group   string `json:"group"`
}

type ArtifactTreeFilter struct {
	Path   string         `json:"path"`
	Format ArtifactFormat `json:"format"`
	Level  string         `json:"level"`
}

type ArtifactRepositoryRes struct {
	ID         int64  `json:"id"`
	Path       string `json:"path"`
	Identifier string `json:"identifier"`
	IsPublic   bool   `json:"is_public"`
	CreatedBy  int64  `json:"created_by"`
	Created    int64  `json:"created"`
	Updated    int64  `json:"updated"`
}

type ArtifactAssetsRes struct {
	Id             int64  `gorm:"column:asset_id"`
	Path           string `gorm:"column:asset_path" json:"path"`
	ContentType    string `gorm:"column:asset_content_type" json:"content_type"`
	Size           int64  `gorm:"column:blob_size" json:"size"`
	CreatorName    string `gorm:"column:principal_uid_unique" json:"creator_name"`
	CheckSumString string `gorm:"column:asset_checksum" json:"-"`
	Created        int64  `gorm:"column:asset_created" json:"created"`
	Updated        int64  `gorm:"column:blob_created" json:"updated"`

	CheckSum *CheckSumRes `gorm:"-" json:"checksum"`
	Link     string       `gorm:"-" json:"link"`
}

type CheckSumRes struct {
	Md5    string `json:"md5"`
	Sha1   string `json:"sha1"`
	SHA256 string `json:"sha256"`
	SHA512 string `json:"sha512"`
}

type ArtifactVersionsRes struct {
	Version     string `json:"version"`
	CreatorName string `json:"creator_name"`
	Updated     int64  `json:"updated"`
}

type ArtifactTreeNodeType string

const (
	ArtifactTreeNodeTypeFormat    ArtifactTreeNodeType = "format"
	ArtifactTreeNodeTypeDirectory ArtifactTreeNodeType = "directory"
	ArtifactTreeNodeTypeVersion   ArtifactTreeNodeType = "version"
	ArtifactTreeNodeTypeAsset     ArtifactTreeNodeType = "asset"
)

type ArtifactTreeRes struct {
	Name     string                `json:"name"`
	Path     string                `json:"path"`
	Leaf     bool                  `json:"leaf"`
	Format   ArtifactFormat        `json:"format"`
	Metadata *ArtifactTreeNodeMeta `json:"metadata,omitempty"`
}

type ArtifactTreeNodeMeta struct {
	Type    ArtifactTreeNodeType `json:"type"`
	Name    string               `json:"name"`
	Group   string               `json:"group"`
	Version string               `json:"version"`
	NodeId  string               `json:"node_id,omitempty"`
}

type ArtifactTreeNodeId struct {
	Type ArtifactTreeNodeType `json:"type"`
	Pk   int64                `json:"pk"`
}

func (nid *ArtifactTreeNodeId) String() string {
	return fmt.Sprintf("%s.%d", nid.Type, nid.Pk)
}

type ArtifactStatus string

const (
	ArtifactStatusOK        ArtifactStatus = "ok"
	ArtifactStatusInvalidId ArtifactStatus = "invalid_id"
	ArtifactStatusNotFound  ArtifactStatus = "not found"
	ArtifactStatusUnknown   ArtifactStatus = "unknown"
)

type ArtifactNodeInfo struct {
	Status   string                  `json:"status"`
	Format   ArtifactFormat          `json:"format"`
	Metadata *ArtifactPkgVerMetadata `json:"metadata"`

	Path        string       `json:"path"`
	ContentType string       `json:"content_type"`
	Size        int64        `json:"size"`
	CreatorName string       `json:"creator_name"`
	Created     int64        `json:"created"`
	Updated     int64        `json:"updated"`
	Link        string       `json:"link"`
	CheckSum    *CheckSumRes `json:"checksum"`
}

type ArtifactPkgVerMetadata struct {
	Space   string `json:"space"`
	Name    string `json:"name"`
	Group   string `json:"group"`
	Version string `json:"version"`
}

type ArtifactNodeRemoveRes struct {
	NodeId   string         `json:"node_id"`
	Status   ArtifactStatus `json:"status"`
	Packages int            `json:"packages,omitempty"`
	Versions int            `json:"versions,omitempty"`
	Assets   int            `json:"assets,omitempty"`
}

type ArtifactNodeRemoveReport struct {
	Total   int                      `json:"total"`
	Success int                      `json:"success"`
	Failed  int                      `json:"failed"`
	Data    []*ArtifactNodeRemoveRes `json:"data"`
}

type ArtifactGarbageReportRes struct {
	Count int64   `json:"count"`
	Size  int64   `json:"size"`
	Ids   []int64 `json:"ids"`
}

type ArtifactStatisticResponse struct {
	GarbageCollect *ArtifactGarbageReportRes     `json:"garbage_collect"`
	Capacity       []*ArtifactPackageCapacityRes `json:"capacity"`
}

type ArtifactPackageCapacityRes struct {
	Space         string                       `json:"space"`
	Name          string                       `json:"name"`
	Namespace     string                       `json:"namespace"`
	Format        ArtifactFormat               `json:"format"`
	Size          int64                        `json:"size"`
	ExclusiveSize int64                        `json:"exclusive_size"`
	Versions      []ArtifactVersionCapacityRes `json:"versions"`
}

type ArtifactVersionCapacityRes struct {
	Version       string `json:"version"`
	Size          int64  `json:"size"`
	ExclusiveSize int64  `json:"exclusive_size"`
}
