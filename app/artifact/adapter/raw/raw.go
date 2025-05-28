// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package raw

import (
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
)

const (
	_segmentName         = `[a-zA-Z0-9](?:[a-zA-Z0-9\-_]*[a-zA-Z0-9])?`
	_regexPackageName    = `^` + _segmentName + `$`
	_regexPackageVersion = `^[a-zA-Z0-9](?:[a-zA-Z0-9\-_\.]*[a-zA-Z0-9])?$`
	_regexPackageGroup   = `^` + _segmentName + `(?:\.` + _segmentName + `)*` + `$`
)

type uploader struct {
	uploadReq *request.ArtifactUploadRequest
	//base      *base.HostedUploader
	store storage.ContentStorage

	storeLayout adapter.StorageLayout

	descriptor *adapter.PackageDescriptor
	bufReader  io.Reader
}

func NewUploader(contentStore storage.ContentStorage, artStore store.ArtifactStore, view *adapter.ViewDescriptor) adapter.ArtifactPackageUploader {
	return &uploader{
		uploadReq:   request.NewUpload(view, artStore),
		store:       contentStore,
		storeLayout: adapter.StorageLayoutBlob,
		descriptor:  adapter.NewEmptyPackageDescriptor(),
	}
}

func (h *uploader) Serve(ctx context.Context, req *http.Request) (int64, error) {
	_ = req.ParseMultipartForm(32 << 20)
	if err := h.validateForm(ctx, req); err != nil {
		return 0, err
	}

	file, fh, err := req.FormFile("file")
	if err != nil {
		return 0, err
	}

	fw, ref, err := adapter.NewRandomBlobWriter(ctx, h.store)
	if err != nil {
		return 0, err
	}

	h.uploadReq.RegisterWriter(fw)
	size, hash, err := request.Write(file, fw)

	h.descriptor.MainAsset.Path = fh.Filename
	h.descriptor.MainAsset.Size = size
	h.descriptor.MainAsset.Hash = hash
	h.descriptor.MainAsset.Ref = ref
	h.descriptor.MainAsset.ContentType = "application/octet-stream"
	h.descriptor.MainAsset.Kind = types.AssetKindMain
	h.descriptor.MainAsset.Attr = adapter.AttrAssetNormal

	h.descriptor.Format = types.ArtifactRawFormat
	return size, err
}

func (h *uploader) validateForm(ctx context.Context, req *http.Request) error {
	var (
		name      = req.FormValue("name")
		namespace = req.FormValue("group")
		version   = req.FormValue("version")
	)

	if name == "" {
		return adapter.ErrMissFormField.WithDetail(fmt.Sprintf("required field: %s", "name"))
	}

	if ok, _ := regexp.MatchString(_regexPackageName, name); !ok {
		return adapter.ErrInvalidPackageName.WithDetail(fmt.Sprintf("require regex pattern: %s", _regexPackageName))
	}
	h.descriptor.Name = name

	if version == "" {
		return adapter.ErrMissFormField.WithDetail(fmt.Sprintf("required field: %s", "version"))
	}
	if ok, _ := regexp.MatchString(_regexPackageVersion, version); !ok {
		return adapter.ErrInvalidPackageVersion.WithDetail(fmt.Sprintf("require regex pattern: %s", _regexPackageVersion))
	}
	h.descriptor.Version = version

	if namespace != "" {
		if ok, _ := regexp.MatchString(_regexPackageGroup, namespace); !ok {
			return adapter.ErrInvalidGroupName
		}
		h.descriptor.Namespace = namespace
	}

	return nil
}

func (h *uploader) IsValid(ctx context.Context) error {
	u := h.uploadReq.LoadCreator(ctx)
	h.descriptor.VersionMetadata = &adapter.VersionMetadata{CreatorName: u.UID}

	h.uploadReq.Descriptor = h.descriptor
	return nil
}

func (h *uploader) Save(ctx context.Context) error {
	if err := h.uploadReq.Commit(ctx); err != nil {
		_ = h.uploadReq.Cancel(ctx)
		return err
	}
	return nil
}

func (h *uploader) Cancel(ctx context.Context) error {
	return h.uploadReq.Cancel(ctx)
}
