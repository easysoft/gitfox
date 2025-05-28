// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/types"
)

const (
	headerContentType   = "Content-Type"
	headerContentLength = "Content-Length"
	headerDisposition   = "Content-Disposition"
	headerCacheControl  = "Cache-Control"
	headerEtag          = "ETag"
	headerLastModify    = "Last-Modified"
	headerLocation      = "Location"
	headerRange         = "Range"
	headerDigest        = "Docker-Content-Digest"
)

type AssetMeta struct {
	Id          int64
	BlobId      int64
	Size        int64
	Path        string
	ContentType string
	ETag        string
	LastModify  int64

	Ref string
}

func newAssetMeta(asset *types.ArtifactAsset, blob *types.ArtifactBlob) *AssetMeta {
	meta := &AssetMeta{
		Id:          asset.ID,
		BlobId:      blob.ID,
		Size:        blob.Size,
		Path:        asset.Path,
		ContentType: asset.ContentType,
		ETag:        "",
		LastModify:  blob.Created,
		Ref:         blob.Ref,
	}
	return meta
}

type VersionAssetMeta struct {
	*AssetMeta
	VersionId int64
}

func (r *AssetMeta) Write(w http.ResponseWriter) {
	w.Header().Set(headerContentLength, fmt.Sprintf("%d", r.Size))
	w.Header().Set(headerContentType, r.ContentType)
	if r.ETag != "" {
		w.Header().Set(headerEtag, r.ETag)
	}

	w.Header().Set(headerLastModify, time.UnixMilli(r.LastModify).Format(time.RFC1123))
	w.Header().Set(headerDisposition, fmt.Sprintf("filename=%s", filepath.Base(r.Path)))
}

func (c *Controller) GetVersionAssetReader(ctx context.Context, viewId, pkgId int64, version, path string) (io.ReadCloser, *VersionAssetMeta, error) {
	meta, err := c.GetVersionAssetInfo(ctx, viewId, pkgId, version, path)
	if err != nil {
		return nil, nil, err
	}
	r, err := c.fileStore.Open(ctx, adapter.BlobPath(meta.Ref))
	return r, meta, err
}

func (c *Controller) GetVersionAssetInfo(ctx context.Context, viewId, pkgId int64, version, path string) (*VersionAssetMeta, error) {
	ver, err := c.artStore.Versions().GetByVersion(ctx, pkgId, viewId, version)
	if err != nil {
		return nil, err
	}

	asset, err := c.artStore.Assets().GetVersionAsset(ctx, path, ver.ID)
	if err != nil {
		return nil, err
	}

	blob, err := c.artStore.Blobs().GetById(ctx, asset.BlobID)
	if err != nil {
		return nil, err
	}

	assetMeta := newAssetMeta(asset, blob)
	verMeta := &VersionAssetMeta{
		AssetMeta: assetMeta,
	}
	return verMeta, nil
}

func (c *Controller) GetMetaReader(ctx context.Context, path string, format types.ArtifactFormat, viewId int64) (io.ReadCloser, *AssetMeta, error) {
	meta, err := c.GetMetaInfo(ctx, path, format, viewId)
	if err != nil {
		return nil, nil, err
	}
	r, err := c.fileStore.Open(ctx, adapter.BlobPath(meta.Ref))
	return r, meta, err
}

func (c *Controller) GetMetaInfo(ctx context.Context, path string, format types.ArtifactFormat, viewId int64) (*AssetMeta, error) {
	asset, err := c.artStore.Assets().GetMetaAsset(ctx, path, viewId, format)
	if err != nil {
		return nil, err
	}

	blob, err := c.artStore.Blobs().GetById(ctx, asset.BlobID)
	if err != nil {
		return nil, err
	}

	return newAssetMeta(asset, blob), nil
}

func (c *Controller) GetAssetInfo(ctx context.Context, path string, format types.ArtifactFormat, viewId *int64) (*AssetMeta, error) {
	asset, err := c.artStore.Assets().GetPath(ctx, path, format)
	if err != nil {
		return nil, err
	}

	blob, err := c.artStore.Blobs().GetById(ctx, asset.BlobID)
	if err != nil {
		return nil, err
	}

	fInfo, err := c.fileStore.Stat(ctx, adapter.BlobPath(blob.Ref))
	if err != nil {
		return nil, err
	}
	if fInfo.Size() != blob.Size {
		return nil, fmt.Errorf("file size mismatch")
	}

	// todo: set a default etag

	return newAssetMeta(asset, blob), nil
}

func (c *Controller) GetAssetReader(ctx context.Context, path string, format types.ArtifactFormat, viewId *int64) (io.ReadCloser, *AssetMeta, error) {
	meta, err := c.GetAssetInfo(ctx, path, format, viewId)
	if err != nil {
		return nil, nil, err
	}

	r, err := c.fileStore.Open(ctx, adapter.BlobPath(meta.Ref))
	return r, meta, err
}

func (c *Controller) GetAssetContent(ctx context.Context, path string, format types.ArtifactFormat, viewId *int64) ([]byte, *AssetMeta, error) {
	meta, err := c.GetAssetInfo(ctx, path, format, viewId)
	if err != nil {
		return nil, nil, err
	}

	content, err := c.fileStore.Get(ctx, adapter.BlobPath(meta.Ref))
	return content, meta, err
}
