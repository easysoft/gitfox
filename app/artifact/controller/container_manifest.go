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
	"regexp"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	"github.com/easysoft/gitfox/types"
)

var (
	regexTag = regexp.MustCompile("^[a-zA-Z0-9_][a-zA-Z0-9._-]{0,127}$")
)

func validateTag(s string) bool {
	return regexTag.MatchString(s)
}

func (c *Controller) HeadContainerManifest(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	return c.getContainerManifest(ctx, req, manifestReq, true)
}

func (c *Controller) GetContainerManifest(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	return c.getContainerManifest(ctx, req, manifestReq, false)
}

func (c *Controller) PutManifest(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest,
) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, manifestReq) != nil {
		return nil, container.ErrDenied
	}

	refDigest := manifestReq.Digest.String()
	uploader := container.NewManifestUploader(c.artStore, manifestReq.view, refDigest)

	if e := c.tx.WithTx(ctx, func(ctx context.Context) error {
		err := handleUpload(ctx, req, uploader)
		if errors.Is(err, adapter.ErrStorageFileNotChanged) {
			return nil
		}
		return err
	}); e != nil {
		return nil, e
	}

	nextUrl := c.urlProvider.GenerateRegistryURL(manifestReq.FullName(), "reference", refDigest)

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(headerLocation, "/"+nextUrl.RequestURI())
		w.Header().Add(headerContentLength, "0")
		w.Header().Add(headerDigest, refDigest)
		w.WriteHeader(http.StatusCreated)
	}), nil
}

func (c *Controller) DeleteContainerManifest(ctx context.Context, manifestReq *ContainerManifestRequest,
) (HttpResponseWriter, error) {
	return c.DeleteContainerBlob(ctx, manifestReq.ContainerReq, manifestReq.Digest.String())
}

func (c *Controller) getContainerManifest(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest, head bool) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPull(ctx, manifestReq) != nil {
		return nil, container.ErrDenied
	}

	meta, err := c.GetAssetInfo(ctx, manifestReq.Digest.String(), types.ArtifactContainerFormat, nil)
	if err != nil {
		return nil, err
	}

	for _, headerVal := range req.Header["If-None-Match"] {
		if headerVal == meta.Path || headerVal == fmt.Sprintf(`"%s"`, meta.Path) { // allow quoted or unquoted
			return NewResponseWriter(func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotModified)
			}), nil
		}
	}

	return c.getManifestWriter(ctx, manifestReq.view, meta, !head)
}

func (c *Controller) getManifestWriter(ctx context.Context, view *adapter.ViewDescriptor, desc *AssetMeta, sendContent bool) (HttpResponseWriter, error) {
	content, err := view.Store.Get(ctx, adapter.BlobPath(desc.Ref))
	if err != nil {
		return nil, err
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(headerEtag, fmt.Sprintf(`"%s"`, desc.Path))
		w.Header().Add(headerContentType, desc.ContentType)
		w.Header().Add(headerContentLength, fmt.Sprintf("%d", desc.Size))
		w.Header().Add(headerDigest, desc.Path)
		w.WriteHeader(http.StatusOK)

		if sendContent {
			w.Write(content)
		}
	}), nil
}
