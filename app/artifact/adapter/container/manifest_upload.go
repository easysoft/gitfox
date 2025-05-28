// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/request"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
)

type manifestUploader struct {
	digest    string
	tag       string
	uploadReq *request.ArtifactUploadRequest
	//base      *base.HostedUploader
	store storage.ContentStorage

	storeLayout adapter.StorageLayout

	descriptor *adapter.PackageDescriptor
	bufReader  io.Reader
}

func NewManifestUploader(artStore store.ArtifactStore, view *adapter.ViewDescriptor, dgst string) *manifestUploader {
	up := &manifestUploader{
		digest:      dgst,
		uploadReq:   request.NewUpload(view, artStore),
		store:       view.Store,
		storeLayout: adapter.StorageLayoutBlob,
		descriptor:  adapter.NewEmptyPackageDescriptor(),
	}
	return up
}

func (h *manifestUploader) Prepare(fn func(desc *adapter.PackageDescriptor)) {
	h.descriptor.Format = types.ArtifactContainerFormat
	fn(h.descriptor)
}

func (h *manifestUploader) Descriptor() *adapter.PackageDescriptor {
	return h.descriptor
}

func (h *manifestUploader) Serve(ctx context.Context, req *http.Request) (int64, error) {
	mediaType := req.Header.Get("Content-Type")
	fw, ref, err := adapter.NewRandomBlobWriter(ctx, h.store)
	if err != nil {
		return 0, err
	}

	buf := bytes.NewBuffer(nil)
	h.bufReader = buf

	h.uploadReq.RegisterWriter(fw)
	size, hash, err := request.Write(req.Body, fw, buf)
	if err != nil {
		return 0, err
	}

	h.descriptor.MainAsset.Size = size
	h.descriptor.MainAsset.Hash = hash
	h.descriptor.MainAsset.Ref = ref
	h.descriptor.MainAsset.ContentType = mediaType
	h.descriptor.MainAsset.Kind = types.AssetKindMain
	if h.digest != "" {
		// saved manifest as a blob, shared cross spaces
		h.descriptor.MainAsset.Attr = adapter.AttrAssetIsolated
	} else {
		// manifest is a tag, should contain only one asset
		h.descriptor.MainAsset.Attr = adapter.AttrAssetExclusive

		// save metadata for container tag
		uploader := h.uploadReq.LoadCreator(ctx)
		h.descriptor.VersionMetadata = &adapter.VersionMetadata{CreatorName: uploader.UID}
	}

	h.descriptor.Format = types.ArtifactContainerFormat
	return size, err
}

func (h *manifestUploader) IsValid(ctx context.Context) error {
	log.Ctx(ctx).Info().Msgf("digest: %s", h.digest)
	log.Ctx(ctx).Info().Msgf("hash sha256: %s", h.descriptor.MainAsset.Hash.Sha256)
	h.descriptor.MainAsset.Path = fmt.Sprintf("sha256:%s", h.descriptor.MainAsset.Hash.Sha256)
	h.descriptor.MainAsset.Format = types.ArtifactContainerFormat
	h.uploadReq.Descriptor = h.descriptor
	return nil
}

func (h *manifestUploader) Save(ctx context.Context) error {
	if err := h.uploadReq.Commit(ctx); err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to commit upload request")
		_ = h.uploadReq.Cancel(ctx)
		return err
	}
	return nil
}

func (h *manifestUploader) Cancel(ctx context.Context) error {
	return h.uploadReq.Cancel(ctx)
}
