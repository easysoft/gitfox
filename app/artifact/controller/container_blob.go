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
	"time"

	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"
)

const (
	HeadDockerUUID = "Docker-Upload-UUID"
	blobCache      = time.Hour * 24 * 365
)

func (c *Controller) HeadBlob(ctx context.Context, ctnReq *ContainerReq, digest string) (HttpResponseWriter, error) {
	return c.getContainerBlob(ctx, ctnReq, digest, true)
}

func (c *Controller) GetContainerBlob(ctx context.Context, ctnReq *ContainerReq, digest string) (HttpResponseWriter, error) {
	return c.getContainerBlob(ctx, ctnReq, digest, false)
}

func (c *Controller) getContainerBlob(ctx context.Context, ctnReq *ContainerReq, digest string, head bool) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPull(ctx, ctnReq) != nil {
		return nil, container.ErrDenied
	}

	fr, meta, err := c.GetAssetReader(ctx, digest, types.ArtifactContainerFormat, nil)
	if err != nil {
		return nil, err
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Set(headerEtag, fmt.Sprintf(`"%s"`, meta.Path))
		w.Header().Set(headerCacheControl, fmt.Sprintf("max-age=%.f", blobCache.Seconds()))
		w.Header().Add(headerContentType, meta.ContentType)
		w.Header().Add(headerContentLength, fmt.Sprintf("%d", meta.Size))
		w.Header().Add(headerDigest, meta.Path)
		defer fr.Close()

		if !head {
			render.Reader(ctx, w, http.StatusOK, fr)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}), nil
}

func (c *Controller) DeleteContainerBlob(ctx context.Context, ctnReq *ContainerReq, digest string) (HttpResponseWriter, error) {
	if c.checkAuthArtifactDelete(ctx, ctnReq) != nil {
		return nil, container.ErrDenied
	}

	meta, err := c.GetAssetInfo(ctx, digest, types.ArtifactContainerFormat, nil)
	if err != nil {
		return nil, err
	}
	if e := c.tx.WithTx(ctx, func(ctx context.Context) error {
		if err = c.artStore.Blobs().DeleteById(ctx, meta.BlobId); err != nil {
			return err
		}
		if err = c.artStore.Assets().DeleteById(ctx, meta.Id); err != nil {
			return err
		}
		_ = c.fileStore.Delete(ctx, adapter.BlobPath(meta.Ref))
		return nil
	}); e != nil {
		if errors.Is(err, gitfox_store.ErrResourceNotFound) {
			return nil, container.ErrManifestBlobUnknown.WithDetail(e.Error())
		}
		return nil, container.ErrUnknown.WithDetail(e.Error())
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusAccepted)
	}), nil
}
