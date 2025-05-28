// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import (
	"context"
	"errors"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/guregu/null"
	"github.com/rs/zerolog/log"
)

func getOrCreatePackage(ctx context.Context,
	artifactStore store.ArtifactStore,
	ownerId int64, name, namespace string, format types.ArtifactFormat,
) (*types.ArtifactPackage, bool, error) {
	var created bool
	pkgModel, err := artifactStore.Packages().GetByName(ctx, name, namespace, ownerId, format)
	if err == nil {
		return pkgModel, created, nil
	}

	if errors.Is(err, gitfox_store.ErrResourceNotFound) {
		pkgModel = &types.ArtifactPackage{
			OwnerID:   ownerId,
			Name:      name,
			Namespace: namespace,
			Format:    format,
		}
		err = artifactStore.Packages().Create(ctx, pkgModel)
		if err != nil {
			return nil, created, err
		}
		created = true
		return pkgModel, created, nil
	}

	return nil, created, err
}

func getOrCreateVersion(ctx context.Context, artifactStore store.ArtifactStore, pkgId, viewId int64, verMeta adapter.MetadataInterface, version string,
) (*types.ArtifactVersion, error) {
	currVerModel, err := artifactStore.Versions().GetByVersion(ctx, pkgId, viewId, version)
	// check existing version need an update

	if err == nil {
		return currVerModel, nil
	}

	if !errors.Is(err, gitfox_store.ErrResourceNotFound) {
		return nil, err
	}

	// create new version
	verModel := &types.ArtifactVersion{
		PackageID: pkgId,
		Version:   version,
		ViewID:    viewId,
	}
	if verMeta != nil {
		metadata, e := verMeta.ToJSON()
		if e == nil {
			verModel.Metadata = string(metadata)
			log.Ctx(ctx).Err(err).Msg("convert version metadata to json failed")
		}
	}

	err = artifactStore.Versions().Create(ctx, verModel)
	if err != nil {
		return nil, err
	}
	return verModel, nil
}

func getOrCreateAsset(ctx context.Context, view *adapter.ViewDescriptor, artifactStore store.ArtifactStore,
	asset *adapter.AssetDescriptor, kind types.AssetKind, ver *types.ArtifactVersion, creator *types.Principal,
) error {
	var blobId int64 = 0
	var blobCreated bool
	if asset.Ref != "" {
		blobModel := &types.ArtifactBlob{
			StorageID: view.StorageID,
			Ref:       asset.Ref,
			Size:      asset.Size,
			Creator:   creator.ID,
		}
		if err := artifactStore.Blobs().Create(ctx, blobModel); err != nil {
			return err
		}
		blobCreated = true
		blobId = blobModel.ID
	}

	// try to find existing asset
	var err error
	var currAssetModel *types.ArtifactAsset
	if ver == nil {
		currAssetModel, err = artifactStore.Assets().GetPath(ctx, asset.Path, asset.Format)
	} else {
		currAssetModel, err = artifactStore.Assets().GetVersionAsset(ctx, asset.Path, ver.ID)
	}

	// got exist asset
	if err == nil {
		newAssetModel := *currAssetModel
		newAssetModel.BlobID = blobId
		if asset.Hash != nil {
			newAssetModel.CheckSum = asset.Hash.String()
		}
		if err = artifactStore.Assets().Update(ctx, currAssetModel); err != nil {
			return err
		}
		if blobCreated {
			if err = artifactStore.Blobs().SoftDeleteById(ctx, currAssetModel.BlobID); err != nil {
				return err
			}
		}
		return nil
	}

	if !errors.Is(err, gitfox_store.ErrResourceNotFound) {
		return err
	}

	newAssetModel := &types.ArtifactAsset{
		Path:        asset.Path,
		Format:      asset.Format,
		Kind:        kind,
		BlobID:      blobId,
		ContentType: asset.ContentType,
	}
	if ver != nil {
		newAssetModel.VersionID = null.IntFrom(ver.ID)
		newAssetModel.ViewID = null.IntFrom(view.ViewID)
	} else {
		newAssetModel.VersionID = null.NewInt(0, false)
		newAssetModel.ViewID = null.NewInt(0, false)
	}

	if asset.Hash != nil {
		newAssetModel.CheckSum = asset.Hash.String()
	}

	if asset.Metadata != nil {
		metadata, e := asset.Metadata.ToJSON()
		if e == nil {
			newAssetModel.Metadata = string(metadata)
		}
	}
	if err = artifactStore.Assets().Create(ctx, newAssetModel); err != nil {
		return err
	}
	return nil
}

func getOrCreateMetaAsset(ctx context.Context, artifactStore store.ArtifactStore, asset *adapter.AssetDescriptor, kind types.AssetKind,
	view *adapter.ViewDescriptor, storageId int64, format types.ArtifactFormat, creator int64,
) error {
	var blobId int64 = 0
	if asset.Ref != "" {
		blobModel := &types.ArtifactBlob{
			StorageID: storageId,
			Ref:       asset.Ref,
			Size:      asset.Size,
			Creator:   creator,
		}
		if err := artifactStore.Blobs().Create(ctx, blobModel); err != nil {
			return err
		}
		blobId = blobModel.ID
	}

	log.Ctx(ctx).Info().Msgf("path: %s, viewId: %d, format: %s", asset.Path, view.ViewID, format)
	currMetaAssetModel, err := artifactStore.Assets().GetMetaAsset(ctx, asset.Path, view.ViewID, format)
	if err == nil {
		newMetaAssetModel := *currMetaAssetModel
		newMetaAssetModel.BlobID = blobId
		if asset.Hash != nil {
			newMetaAssetModel.CheckSum = asset.Hash.String()
		}
		if err = artifactStore.Assets().Update(ctx, &newMetaAssetModel); err != nil {
			return err
		}
		return nil
	}

	if !errors.Is(err, gitfox_store.ErrResourceNotFound) {
		return err
	}

	currMetaAssetModel = &types.ArtifactAsset{
		ViewID:      null.IntFrom(view.ViewID),
		Format:      format,
		Path:        asset.Path,
		Kind:        kind,
		BlobID:      blobId,
		ContentType: asset.ContentType,
	}
	if asset.Hash != nil {
		currMetaAssetModel.CheckSum = asset.Hash.String()
	}
	if err = artifactStore.Assets().Create(ctx, currMetaAssetModel); err != nil {
		return err
	}
	return nil
}
