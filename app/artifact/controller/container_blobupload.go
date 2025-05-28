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
	"net/url"

	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	storagedriver "github.com/easysoft/gitfox/pkg/storage/driver"

	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
)

func (c *Controller) SaveContainerBlob(
	ctx context.Context, ctnReq *ContainerReq,
	req *http.Request, reference digest.Digest,
) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, ctnReq) != nil {
		return nil, container.ErrDenied
	}

	refDigest := reference.String()
	uploader := container.NewManifestUploader(c.artStore, ctnReq.view, refDigest)

	if e := c.tx.WithTx(ctx, func(ctx context.Context) error {
		return handleUpload(ctx, req, uploader)
	}); e != nil {
		return nil, e
	}

	nextUrl := c.urlProvider.GenerateRegistryURL(ctnReq.FullName(), "blobs", refDigest)

	//desc := uploader.Descriptor()
	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(headerLocation, "/"+nextUrl.RequestURI())
		w.WriteHeader(http.StatusCreated)
	}), nil
}

func (c *Controller) StartBlobUpload(
	ctx context.Context, ctnReq *ContainerReq,
) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, ctnReq) != nil {
		return nil, container.ErrDenied
	}

	req := container.InitResumeRequest(ctnReq.FullName(), c.artStore, ctnReq.view)
	return c.getUploadState(ctx, req)
}

func (c *Controller) getUploadState(ctx context.Context, req *container.ResumeRequest) (HttpResponseWriter, error) {
	nextUrl := c.urlProvider.GenerateRegistryURL(req.RepoName, "blobs", "uploads", req.SessionID)

	token, err := req.State.Pack("")
	if err != nil {
		return nil, err
	}
	vals := url.Values{"_state": []string{token}}
	nextUrl.RawQuery = vals.Encode()

	endSize, err := req.Size(ctx)
	log.Debug().Msgf("get upload size: %d, offset: %d, ref: %s", endSize, req.State.Offset, req.Ref)
	if err != nil {
		// return zero for non-exist file, because init resume request doesn't create any file
		var pathNotFoundError storagedriver.PathNotFoundError
		if !errors.As(err, &pathNotFoundError) {
			return nil, err
		}
	}
	if endSize > 0 {
		endSize = endSize - 1
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(HeadDockerUUID, req.State.SessionID)
		w.Header().Add(headerLocation, "/"+nextUrl.RequestURI())
		w.Header().Add(headerContentLength, "0")
		w.Header().Add(headerRange, fmt.Sprintf("0-%d", endSize))
	}), nil
}

func (c *Controller) GetBlobUploadStatus(ctx context.Context, uploadRequest *ContainerUploadRequest) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, uploadRequest) != nil {
		return nil, container.ErrDenied
	}

	return c.getUploadState(ctx, uploadRequest.resume)
}

func (c *Controller) AppendBlobContent(ctx context.Context, r *http.Request, uploadRequest *ContainerUploadRequest) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, uploadRequest) != nil {
		return nil, container.ErrDenied
	}

	resumeReq := uploadRequest.resume
	if e := c.tx.WithTx(ctx, func(ctx context.Context) error {
		err := resumeReq.AppendWrite(ctx, r)
		return err
	}); e != nil {
		log.Ctx(ctx).Info().Msg("append blob failed")
		return nil, e
	}

	log.Ctx(ctx).Info().Msgf("append blob content successfully")
	return c.getUploadState(ctx, resumeReq)
}

func (c *Controller) FinishBlobUpload(ctx context.Context, r *http.Request, uploadRequest *ContainerUploadRequest) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, uploadRequest) != nil {
		return nil, container.ErrDenied
	}

	dgstStr := r.FormValue("digest")
	if dgstStr == "" {
		// no digest? return error, but allow retry.
		return nil, container.ErrDigestInvalid.WithDetail("digest missing")
	}

	dgst, err := digest.Parse(dgstStr)
	if err != nil {
		return nil, container.ErrDigestInvalid.WithDetail("digest parsing failed")
	}

	resumeReq := uploadRequest.resume
	var nextUrl *url.URL
	if err = c.tx.WithTx(ctx, func(ctx context.Context) error {
		if e := resumeReq.Finish(ctx, r, &dgst); e != nil {
			return e
		}
		nextUrl = c.urlProvider.GenerateRegistryURL(uploadRequest.FullName(), "blobs", "uploads", dgst.String())
		return nil
	}); err != nil {
		return nil, err
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(headerLocation, "/"+nextUrl.RequestURI())
		w.Header().Add(headerContentLength, "0")
		w.Header().Add(headerDigest, dgst.String())
	}), nil
}

func (c *Controller) CancelBlobUpload(ctx context.Context, uploadRequest *ContainerUploadRequest) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, uploadRequest) != nil {
		return nil, container.ErrDenied
	}

	resumeReq := uploadRequest.resume
	if e := c.tx.WithTx(ctx, func(ctx context.Context) error {
		return resumeReq.Remove(ctx)
	}); e != nil {
		return nil, e
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(HeadDockerUUID, resumeReq.SessionID)
		w.WriteHeader(http.StatusNoContent)
	}), nil
}
