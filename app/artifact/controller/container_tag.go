// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
)

type tagInfo struct {
	name    string
	version *types.ArtifactVersion
	asset   *types.ArtifactAsset
	blob    *types.ArtifactBlob
}

func findContainerTag(ctx context.Context, artStore store.ArtifactStore, viewId, ownerId int64, pkgName, tag string) (*tagInfo, error) {
	dbPkg, err := artStore.Packages().GetByName(ctx, pkgName, "", ownerId, types.ArtifactContainerFormat)
	if err != nil {
		return nil, err
	}
	dbVer, err := artStore.Versions().GetByVersion(ctx, dbPkg.ID, viewId, tag)
	if err != nil {
		return nil, err
	}
	dbAssets, err := artStore.Assets().FindMain(ctx, dbVer.ID)
	if err != nil {
		return nil, err
	}
	if len(dbAssets) != 1 {
		return nil, fmt.Errorf("expected 1 asset, found %d", len(dbAssets))
	}
	dbAsset := dbAssets[0]
	dbBlob, err := artStore.Blobs().GetById(ctx, dbAsset.BlobID)
	if err != nil {
		return nil, err
	}

	return &tagInfo{
		name:    tag,
		version: dbVer,
		asset:   dbAsset,
		blob:    dbBlob,
	}, nil
}

func (c *Controller) HeadContainerTag(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	return c.getContainerTag(ctx, req, manifestReq, true)
}

func (c *Controller) GetContainerTag(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	return c.getContainerTag(ctx, req, manifestReq, false)
}

func (c *Controller) getContainerTag(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest, head bool) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPull(ctx, manifestReq) != nil {
		return nil, container.ErrDenied
	}

	view := manifestReq.view
	t, err := findContainerTag(ctx, c.artStore, view.ViewID, view.OwnerID, manifestReq.repoName, manifestReq.Tag)
	if err != nil {
		return nil, err
	}

	meta := newAssetMeta(t.asset, t.blob)
	for _, headerVal := range req.Header["If-None-Match"] {
		if headerVal == meta.Path || headerVal == fmt.Sprintf(`"%s"`, meta.Path) { // allow quoted or unquoted
			return NewResponseWriter(func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotModified)
			}), nil
		}
	}

	return c.getManifestWriter(ctx, view, meta, !head)
}

func (c *Controller) PutContainerTag(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, manifestReq) != nil {
		return nil, container.ErrDenied
	}

	view := manifestReq.view
	uploader := container.NewManifestUploader(c.artStore, view, "")
	uploader.Prepare(func(desc *adapter.PackageDescriptor) {
		desc.Name = manifestReq.repoName
		desc.Version = manifestReq.Tag
	})

	if e := c.tx.WithTx(ctx, func(ctx context.Context) error {
		err := handleUpload(ctx, req, uploader)
		if errors.Is(err, adapter.ErrStorageFileNotChanged) {
			return nil
		}
		return err
	}); e != nil {
		log.Ctx(ctx).Err(e).Msg("save container tag failed")
		return nil, e
	}
	nextUrl := c.urlProvider.GenerateRegistryURL(manifestReq.FullName(), "reference", manifestReq.Tag)

	desc := uploader.Descriptor()
	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(headerLocation, "/"+nextUrl.RequestURI())
		w.Header().Add(headerContentLength, "0")
		w.Header().Add(headerDigest, desc.MainAsset.Path)
		w.WriteHeader(http.StatusCreated)
	}), nil
}

func (c *Controller) DeleteContainerTag(ctx context.Context, req *ContainerManifestRequest) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, req) != nil {
		return nil, container.ErrDenied
	}

	t, err := findContainerTag(ctx, c.artStore, req.view.ViewID, req.view.OwnerID, req.repoName, req.Tag)
	if err != nil {
		return nil, err
	}
	if e := c.tx.WithTx(ctx, func(ctx context.Context) error {
		if err = c.artStore.Versions().DeleteById(ctx, t.version.ID); err != nil {
			return err
		}
		if err = c.artStore.Assets().DeleteById(ctx, t.asset.ID); err != nil {
			return err
		}

		// todo: a blob is linked to multi asset
		if err = c.artStore.Blobs().DeleteById(ctx, t.blob.ID); err != nil {
			return err
		}
		_ = c.fileStore.Delete(ctx, adapter.BlobPath(t.blob.Ref))
		return nil
	}); e != nil {
		log.Ctx(ctx).Err(e).Msg("delete container tag failed")
		return nil, e
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusAccepted)
	}), nil
}

type tagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (c *Controller) ListTags(ctx context.Context, req *ContainerReq) (*tagList, error) {
	if c.checkAuthArtifactPull(ctx, req) != nil {
		return nil, container.ErrDenied
	}

	var result = tagList{
		Name: req.repoName,
		Tags: make([]string, 0),
	}

	pkgModel, err := c.artStore.Packages().GetByName(ctx, req.repoName, "", req.view.OwnerID, types.ArtifactContainerFormat)
	if err != nil {
		return &result, nil
	}

	objects, err := c.artStore.Versions().Find(ctx, types.SearchVersionOption{
		PackageId: pkgModel.ID, ViewId: req.view.ViewID,
	})
	if err != nil {
		return &result, nil
	}

	for _, object := range objects {
		result.Tags = append(result.Tags, object.Version)
	}
	return &result, nil
}
