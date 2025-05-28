// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import (
	"context"

	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/store"
	storagedriver "github.com/easysoft/gitfox/pkg/storage/driver"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
)

type ArtifactRequest struct {
	view          *adapter.ViewDescriptor
	storageId     int64
	artifactStore store.ArtifactStore
	modelMgr      *adapter.ModelManager

	writers []storagedriver.FileWriter
}

func newRequest(view *adapter.ViewDescriptor, artifactStore store.ArtifactStore) *ArtifactRequest {
	return &ArtifactRequest{
		view:          view,
		storageId:     view.StorageID,
		artifactStore: artifactStore,
		writers:       make([]storagedriver.FileWriter, 0),
	}
}

type ArtifactUploadRequest struct {
	*ArtifactRequest
	creator    *types.Principal
	Descriptor *adapter.PackageDescriptor
}

func NewUpload(view *adapter.ViewDescriptor, artifactStore store.ArtifactStore) *ArtifactUploadRequest {
	req := ArtifactUploadRequest{
		ArtifactRequest: newRequest(view, artifactStore),
	}

	return &req
}

func (r *ArtifactUploadRequest) RegisterWriter(writer storagedriver.FileWriter) {
	r.writers = append(r.writers, writer)
}

func (r *ArtifactUploadRequest) LoadCreator(ctx context.Context) *types.Principal {
	session, ok := request.AuthSessionFrom(ctx)
	if !ok {
		return &auth.AnonymousPrincipal
	} else {
		return session.User
	}
}

func (r *ArtifactUploadRequest) Commit(ctx context.Context) error {
	logger := log.Ctx(ctx)
	var err error

	modelMgr := adapter.NewModelManager(r.artifactStore, r.view, r.Descriptor, r.LoadCreator(ctx))
	if _, err = modelMgr.Sync(ctx, r.Descriptor.MainAsset); err != nil {
		return err
	}

	for _, subAsset := range r.Descriptor.SubAssets {
		if _, err = modelMgr.Sync(ctx, subAsset); err != nil {
			return err
		}
	}

	// commit file writer
	logger.Debug().Msgf("commit file writer for artifact %s", r.Descriptor.Name)
	for _, fw := range r.writers {
		if err = fw.Commit(ctx); err != nil {
			logger.Warn().Err(err).Msg("commit file upload failed")
			return err
		}
		if err = fw.Close(); err != nil {
			logger.Warn().Err(err).Msg("commit file close failed")
		}
	}

	return nil
}

func (r *ArtifactUploadRequest) Cancel(ctx context.Context) error {
	logger := log.Ctx(ctx)

	// clean files
	for _, fw := range r.writers {
		if err := fw.Cancel(ctx); err != nil {
			logger.Warn().Err(err).Msg("cancel file upload failed")
		}
	}
	return nil
}
