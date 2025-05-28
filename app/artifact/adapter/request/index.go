// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import (
	"context"
	"errors"

	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/store"
	storagedriver "github.com/easysoft/gitfox/pkg/storage/driver"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
)

type ArtifactIndexRequest struct {
	*ArtifactRequest

	Descriptor *adapter.PackageMetaDescriptor
}

func NewIndex(view *adapter.ViewDescriptor, artifactStore store.ArtifactStore) *ArtifactIndexRequest {
	return &ArtifactIndexRequest{
		ArtifactRequest: newRequest(view, artifactStore),
	}
}

func (r *ArtifactIndexRequest) RegisterWriter(writer storagedriver.FileWriter) {
	r.writers = append(r.writers, writer)
}

func (r *ArtifactIndexRequest) Commit(ctx context.Context) error {
	logger := log.Ctx(ctx)

	// AnonymousUser is only used for unittest
	var creatorId = auth.AnonymousPrincipal.ID
	session, ok := request.AuthSessionFrom(ctx)
	if ok {
		creatorId = session.User.ID
	}

	if r.Descriptor.MainAsset == nil {
		return errors.New("index main asset can't be nil")
	}
	// setup main asset model
	mainMetaAsset := r.Descriptor.MainAsset
	err := getOrCreateMetaAsset(ctx, r.artifactStore, mainMetaAsset, types.AssetKindMain, r.view, r.storageId, r.Descriptor.Format, creatorId)
	if err != nil {
		return err
	}

	// setup sub asset models
	for _, subAsset := range r.Descriptor.SubAssets {
		if err = getOrCreateMetaAsset(ctx, r.artifactStore, subAsset, types.AssetKindSub, r.view, r.storageId, r.Descriptor.Format, creatorId); err != nil {
			return err
		}
	}

	// commit file writer
	for _, fw := range r.writers {
		if err = fw.Commit(ctx); err != nil {
			logger.Warn().Err(err).Msg("commit file upload failed")
			return err
		}
	}
	return nil
}

func (r *ArtifactIndexRequest) Cancel(ctx context.Context) error {
	logger := log.Ctx(ctx)

	// clean files
	for _, fw := range r.writers {
		if err := fw.Cancel(ctx); err != nil {
			logger.Warn().Err(err).Msg("cancel file upload failed")
		}
	}
	return nil
}
