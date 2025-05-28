// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package helm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/request"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
)

const (
	_chartPkgType = "application/x-compressed-tar"

	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names
	_regexChartName = `^[a-z][a-z0-9\-]+[a-z0-9]$`
)

type helmUploader struct {
	uploadReq *request.ArtifactUploadRequest
	//base      *base.HostedUploader
	store storage.ContentStorage

	storeLayout adapter.StorageLayout

	descriptor *adapter.PackageDescriptor
	bufReader  io.Reader
}

func NewUploader(contentStore storage.ContentStorage, artStore store.ArtifactStore, view *adapter.ViewDescriptor) adapter.ArtifactPackageUploader {
	return &helmUploader{
		uploadReq:   request.NewUpload(view, artStore),
		store:       contentStore,
		storeLayout: adapter.StorageLayoutBlob,
		descriptor:  adapter.NewEmptyPackageDescriptor(),
	}
}

func (h *helmUploader) Serve(ctx context.Context, req *http.Request) (int64, error) {
	_ = req.ParseMultipartForm(32 << 20)
	file, _, err := req.FormFile("chart")
	if err != nil {
		return 0, err
	}

	fw, ref, err := adapter.NewRandomBlobWriter(ctx, h.store)
	if err != nil {
		return 0, err
	}

	buf := bytes.NewBuffer(nil)
	h.bufReader = buf

	h.uploadReq.RegisterWriter(fw)
	size, hash, err := request.Write(file, fw, buf)

	h.descriptor.MainAsset.Size = size
	h.descriptor.MainAsset.Hash = hash
	h.descriptor.MainAsset.Ref = ref
	h.descriptor.MainAsset.ContentType = _chartPkgType
	h.descriptor.MainAsset.Kind = types.AssetKindMain
	h.descriptor.MainAsset.Attr = adapter.AttrAssetNormal
	return size, err
}

func (h *helmUploader) IsValid(ctx context.Context) error {
	chart, err := ParseChart(h.bufReader)
	if err != nil {
		return adapter.ErrInvalidPackageContent.WithDetail(err.Error())
	}
	log.Ctx(ctx).Info().Msgf("chart: %+v", chart)

	if ok, _ := regexp.MatchString(_regexChartName, chart.Name()); !ok {
		return adapter.ErrInvalidPackageName
	}

	h.descriptor.MainAsset.Path = fmt.Sprintf("charts/%s-%s.tgz", chart.Metadata.Name, chart.Metadata.Version)

	h.descriptor.Name = chart.Metadata.Name
	h.descriptor.Namespace = ""
	h.descriptor.Version = chart.Metadata.Version
	h.descriptor.Format = types.ArtifactHelmFormat
	h.descriptor.MainAsset.Metadata = &AssetMetadata{data: chart.Metadata}
	uploader := h.uploadReq.LoadCreator(ctx)
	h.descriptor.VersionMetadata = &adapter.VersionMetadata{CreatorName: uploader.UID}
	h.uploadReq.Descriptor = h.descriptor
	return nil
}

func (h *helmUploader) Save(ctx context.Context) error {
	if err := h.uploadReq.Commit(ctx); err != nil {
		_ = h.uploadReq.Cancel(ctx)
		return err
	}
	return nil
}

func (h *helmUploader) Cancel(ctx context.Context) error {
	return h.uploadReq.Cancel(ctx)
}
