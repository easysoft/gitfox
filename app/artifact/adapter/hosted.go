// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package adapter

import (
	"context"
	"net/http"

	"github.com/easysoft/gitfox/types"
)

type CommitWriter interface {
	Commit(ctx context.Context) error
	Cancel(ctx context.Context) error
}

type ArtifactPackageUploader interface {
	Serve(ctx context.Context, req *http.Request) (int64, error)
	IsValid(ctx context.Context) error
	Cancel(ctx context.Context) error
	Save(ctx context.Context) error
}

// ArtifactIndexUpdater is an index update interface
// every format-adapter should implement it
type ArtifactIndexUpdater interface {
	// UpdatePackage will be called after a new version of package is upload,
	// some format-adapter need to generate package-level index file
	UpdatePackage(ctx context.Context, p *types.ArtifactPackage, v *types.ArtifactView) error

	// UpdateRepo will be called after a new version of package is upload,
	// some format-adapter need to generate repository-level index file
	UpdateRepo(ctx context.Context, v *types.ArtifactView) error
	CommitWriter
}

type PackageDescriptor struct {
	// The artifact package's name
	Name string

	Namespace string

	Version         string
	VersionMetadata MetadataInterface

	Format types.ArtifactFormat

	// Upload file info
	MainAsset *AssetDescriptor

	// auto created sub files
	SubAssets []*AssetDescriptor
}

func NewEmptyPackageDescriptor() *PackageDescriptor {
	return &PackageDescriptor{MainAsset: &AssetDescriptor{}, SubAssets: make([]*AssetDescriptor, 0)}
}

func (p *PackageDescriptor) AddSub(asset *AssetDescriptor) {
	p.SubAssets = append(p.SubAssets, asset)
}

type AssetDescriptor struct {
	Attr ModelAttribute

	Path        string
	Ref         string
	Size        int64
	Metadata    MetadataInterface
	ContentType string
	Format      types.ArtifactFormat
	Kind        types.AssetKind

	Hash *Hash
}

type PackageMetaDescriptor struct {
	Name string
	// Upload file info

	Format types.ArtifactFormat

	MainAsset *AssetDescriptor

	// auto created sub files
	SubAssets []*AssetDescriptor
}

func NewEmptyPackageMetaDescriptor() *PackageMetaDescriptor {
	return &PackageMetaDescriptor{MainAsset: &AssetDescriptor{}, SubAssets: make([]*AssetDescriptor, 0)}
}

func (p *PackageMetaDescriptor) AddSub(asset *AssetDescriptor) {
	p.SubAssets = append(p.SubAssets, asset)
}
