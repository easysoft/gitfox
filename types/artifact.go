// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package types

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
)

type ArtifactFormat string

const (
	ArtifactAllFormat       ArtifactFormat = "all"
	ArtifactRawFormat       ArtifactFormat = "raw"
	ArtifactMavenFormat     ArtifactFormat = "maven"
	ArtifactContainerFormat ArtifactFormat = "container"
	ArtifactHelmFormat      ArtifactFormat = "helm"
	ArtifactPypiFormat      ArtifactFormat = "pypi"
	ArtifactNpmFormat       ArtifactFormat = "npm"
)

var AllArtifactFormatList = []ArtifactFormat{ArtifactRawFormat, ArtifactMavenFormat, ArtifactContainerFormat, ArtifactHelmFormat}

type AssetKind string

const (
	AssetKindMain AssetKind = "main"
	AssetKindSub  AssetKind = "subordinate"
)

func ValidateAssetKind(kind AssetKind) error {
	if kind != AssetKindMain && kind != AssetKindSub {
		return fmt.Errorf("unidentified kind '%s'", kind)
	}
	return nil
}

const (
	ArtifactNodeRoot = "/"

	ArtifactNodeLevelAsset   = "asset"
	ArtifactNodeLevelVersion = "version"
)

type ArtifactRepository struct {
	ID          int64                 `gorm:"column:repo_id;primaryKey"   json:"-"`
	Identifier  string                `gorm:"column:repo_identifier"      json:"identifier"`
	Description string                `gorm:"column:repo_description"     json:"description"`
	DisplayName string                `gorm:"column:repo_display_name"    json:"display_name"`
	Kind        enum.ArtifactRepoKind `gorm:"column:repo_kind"            json:"kind"`
	RefID       null.Int              `gorm:"column:repo_ref_id"          json:"-"`
	CreatedBy   int64                 `gorm:"column:repo_created_by"      json:"created_by"`
	Created     int64                 `gorm:"column:repo_created"         json:"created"`
	Updated     int64                 `gorm:"column:repo_updated"         json:"updated"`
}

type ArtifactView struct {
	ID          int64  `gorm:"column:view_id;primaryKey" json:"id"`
	Name        string `gorm:"column:view_name"`
	Description string `gorm:"column:view_description"`
	SpaceID     int64  `gorm:"column:view_space_id"`
	Default     bool   `gorm:"column:view_is_default"`
}

type ArtifactPackage struct {
	ID        int64          `gorm:"column:package_id;primaryKey" json:"id"`
	OwnerID   int64          `gorm:"column:package_owner_id" json:"owner_id"`
	Name      string         `gorm:"column:package_name" json:"name"`
	Namespace string         `gorm:"column:package_namespace" json:"namespace"`
	Format    ArtifactFormat `gorm:"column:package_format" json:"format"`
	Created   int64          `gorm:"column:package_created;autoCreateTime:milli" json:"created"`
	Updated   int64          `gorm:"column:package_updated;autoUpdateTime:milli" json:"updated"`
	Deleted   int64          `gorm:"column:package_deleted"`
}

func (p *ArtifactPackage) IsDeleted() bool {
	return p.Deleted > 0
}

type ArtifactVersion struct {
	ID        int64  `gorm:"column:version_id;primaryKey"`
	PackageID int64  `gorm:"column:version_package_id"`
	Version   string `gorm:"column:version"`

	ViewID int64 `gorm:"column:version_view_id"`
	// Metadata store version-level metadata. example: version creator
	Metadata string `gorm:"column:version_metadata"`

	Created int64 `gorm:"column:version_created;autoCreateTime:milli" json:"created"`
	Updated int64 `gorm:"column:version_updated;autoUpdateTime:milli" json:"updated"`
	Deleted int64 `gorm:"column:version_deleted"`
}

type ArtifactVersionInfo struct {
	ID      int64  `gorm:"column:version_id"`
	Version string `gorm:"column:version"`
	ViewID  int64  `gorm:"column:version_view_id"`
	Deleted int64  `gorm:"column:version_deleted"`

	PackageId        int64          `gorm:"column:package_id"`
	PackageName      string         `gorm:"column:package_name"`
	PackageNamespace string         `gorm:"column:package_namespace"`
	PackageFormat    ArtifactFormat `gorm:"column:package_format"`

	SpaceName string `gorm:"column:space_uid"`
}

func (v *ArtifactVersion) IsDeleted() bool {
	return v.Deleted > 0
}

type ArtifactAsset struct {
	ID int64 `gorm:"column:asset_id;primaryKey"`
	// VersionID binding with package(with format), view
	VersionID null.Int `gorm:"column:asset_version_id;"`
	// ViewID used for none-version assets, like index file for a format
	ViewID null.Int `gorm:"column:asset_view_id"`
	// Format is also used for none-version assets
	Format      ArtifactFormat `gorm:"column:asset_format"`
	Path        string         `gorm:"column:asset_path"`
	ContentType string         `gorm:"column:asset_content_type"`
	Kind        AssetKind      `gorm:"column:asset_kind"`
	// Metadata store asset-level metadata. example: helm chart metadata, for rebuild index
	Metadata string `gorm:"column:asset_metadata"`
	BlobID   int64  `gorm:"column:asset_blob_id"`
	CheckSum string `gorm:"column:asset_checksum"`
	Created  int64  `gorm:"column:asset_created;autoCreateTime:milli" json:"created"`
	Updated  int64  `gorm:"column:asset_updated;autoUpdateTime:milli" json:"updated"`
	Deleted  int64  `gorm:"column:asset_deleted"`
}

type ArtifactAssetExtendBlob struct {
	ArtifactAsset
	BlobID    int64  `gorm:"column:blob_id"`
	StorageID int64  `gorm:"column:storage_id"`
	Ref       string `gorm:"column:blob_ref"`
	Size      int64  `gorm:"column:blob_size"`
}

type ArtifactMetaAsset struct {
	ID          int64          `gorm:"column:meta_asset_id;primaryKey"`
	OwnerID     int64          `gorm:"column:meta_asset_owner_id"` // to be removed
	Format      ArtifactFormat `gorm:"column:meta_asset_format"`
	Path        string         `gorm:"column:meta_asset_path"`
	ViewID      int64          `gorm:"column:meta_asset_view_id"`
	ContentType string         `gorm:"column:meta_asset_content_type"`
	Kind        AssetKind      `gorm:"column:meta_asset_kind"`
	BlobID      int64          `gorm:"column:meta_asset_blob_id"`
	CheckSum    string         `gorm:"column:meta_asset_checksum"`
	Created     int64          `gorm:"column:meta_asset_created;autoCreateTime:milli" json:"created"`
	Updated     int64          `gorm:"column:meta_asset_updated;autoUpdateTime:milli" json:"updated"`
}

type ArtifactBlob struct {
	ID        int64  `gorm:"column:blob_id;primaryKey"`
	StorageID int64  `gorm:"column:storage_id"`
	Ref       string `gorm:"column:blob_ref"`
	Size      int64  `gorm:"column:blob_size"`
	Downloads int64  `gorm:"column:blob_downloads"`
	// Metadata store file-level metadata. example: container upload progress hash state
	Metadata string   `gorm:"column:blob_metadata"`
	Deleted  null.Int `gorm:"column:blob_deleted"`
	Created  int64    `gorm:"column:blob_created;autoCreateTime:milli" json:"created"`
	Creator  int64    `gorm:"column:blob_creator"`
}

type ArtifactTreeNode struct {
	ID       int64                `gorm:"column:node_id;primaryKey"`
	ParentID null.Int             `gorm:"column:node_parent_id"`
	OwnerID  int64                `gorm:"column:node_owner_id"`
	Name     string               `gorm:"column:node_name"`
	Path     string               `gorm:"column:node_path"`
	Type     ArtifactTreeNodeType `gorm:"column:node_type"`
	Format   ArtifactFormat       `gorm:"column:node_format"`
}

func (n *ArtifactTreeNode) IsRoot() bool {
	return n.Path == "/"
}

/*
Parent return the object parent node

	inputs:
	/a/b/c -> /a/b
	/a/b -> /a
	/a -> /
*/
func (n *ArtifactTreeNode) Parent() *ArtifactTreeNode {
	p := ArtifactTreeNode{
		OwnerID: n.OwnerID,
		Format:  n.Format,
	}

	parentPath := filepath.Dir(n.Path)
	p.Name = filepath.Base(parentPath)
	p.Path = parentPath

	if p.Path == "/" {
		p.Name = string(n.Format)
		p.Type = ArtifactTreeNodeTypeFormat
	} else {
		p.Name = filepath.Base(parentPath)
		p.Type = ArtifactTreeNodeTypeDirectory
	}
	return &p
}

var (
	ErrPkgNotFound      = errors.New("artifact package not found")
	ErrPkgNoItemDeleted = errors.New("no artifact package was deleted")

	ErrPkgVersionNotFound      = errors.New("artifact package version not found")
	ErrPkgVersionNoItemDeleted = errors.New("no artifact package was deleted")

	ErrAssetNotFound      = errors.New("artifact asset not found")
	ErrAssetNoItemDeleted = errors.New("no artifact asset was deleted")

	ErrMetaAssetNotFound      = errors.New("artifact meta asset not found")
	ErrMetaAssetNoItemDeleted = errors.New("no artifact meta asset was deleted")

	ErrBlobNotFound      = errors.New("artifact blob not found")
	ErrBlobNoItemDeleted = errors.New("no artifact blob was deleted")

	ErrViewDefaultNotFound = errors.New("default view not found")
	ErrViewNotFound        = errors.New("special view not found")
	ErrArgsValueEmpty      = errors.New("args value should not empty")
)

type ArtifactRecycleBlobDesc struct {
	AssetId int64
	Path    string
	BlobId  int64
	BlobRef string
	Size    int64
}

type ArtifactVersionCapacityDesc struct {
	AssetId       int64
	VersionId     int64
	ExclusiveSize int64
	TotalSize     int64
	ExclusiveRefs int
	TotalRefs     int
}

type ArtifactStatisticReport struct {
	DeleteList []*ArtifactRecycleBlobDesc
	TagList    []*ArtifactVersionCapacityDesc
}
