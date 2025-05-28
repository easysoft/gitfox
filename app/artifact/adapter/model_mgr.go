// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/easysoft/gitfox/app/artifact/model"
	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/guregu/null"
	"github.com/rs/zerolog/log"
)

type ModelAttribute string

const (
	// AttrAssetNormal should be bind to a version
	AttrAssetNormal ModelAttribute = "normal"

	// AttrAssetIsolated does not bind to version nor view, used for container blobs
	AttrAssetIsolated ModelAttribute = "isolated"

	// AttrAssetExclusive require the version has only one asset, others should be removed
	AttrAssetExclusive ModelAttribute = "exclusive"

	// AttrAssetIndex only bind to view, used for generated index
	AttrAssetIndex ModelAttribute = "index"
)

type ModelManager struct {
	store   store.ArtifactStore
	view    *ViewDescriptor
	creator *types.Principal

	pkgDescriptor *PackageDescriptor
	pkgModel      *types.ArtifactPackage
	versionModel  *types.ArtifactVersion
}

func NewModelManager(artStore store.ArtifactStore, view *ViewDescriptor, pkg *PackageDescriptor, creator *types.Principal) *ModelManager {
	return &ModelManager{
		store:         artStore,
		view:          view,
		creator:       creator,
		pkgDescriptor: pkg,
	}
}

func (mm *ModelManager) Sync(ctx context.Context, assetDesc *AssetDescriptor) (*types.ArtifactAsset, error) {
	switch assetDesc.Attr {
	case AttrAssetIsolated:
		return mm.createIsolatedAsset(ctx, assetDesc)
	case AttrAssetExclusive:
		ver, err := mm.getOrCreateVersion(ctx)
		if err != nil {
			return nil, err
		}
		asset, err := mm.createExclusiveAsset(ctx, ver, assetDesc)
		if err != nil {
			return nil, nil
		}
		if err = mm.updateVerTime(ctx, ver, asset); err != nil {
			return nil, err
		}
		return asset, nil
	default:
		ver, err := mm.getOrCreateVersion(ctx)
		if err != nil {
			return nil, err
		}
		asset, err := mm.createNormalAsset(ctx, ver, assetDesc)
		if err != nil {
			return nil, nil
		}
		if err = mm.updateVerTime(ctx, ver, asset); err != nil {
			return nil, err
		}
		return asset, nil
	}
}

func (mm *ModelManager) createIsolatedAsset(ctx context.Context, descriptor *AssetDescriptor) (*types.ArtifactAsset, error) {
	currAsset, assetExist, err := mm.getExistAsset(func() (*types.ArtifactAsset, error) {
		return mm.store.Assets().GetPath(ctx, descriptor.Path, descriptor.Format)
	})
	if err != nil {
		return nil, err
	}

	if assetExist {
		return mm.updateAsset(ctx, currAsset, descriptor)
	}

	newAsset := &types.ArtifactAsset{}
	if err = mm.createAsset(ctx, newAsset, descriptor); err != nil {
		return nil, err
	}
	return newAsset, nil
}

func (mm *ModelManager) createExclusiveAsset(ctx context.Context, ver *types.ArtifactVersion, descriptor *AssetDescriptor) (*types.ArtifactAsset, error) {
	currAsset, assetExist, err := mm.getExistAsset(func() (*types.ArtifactAsset, error) {
		return mm.store.Assets().GetVersionAsset(ctx, descriptor.Path, ver.ID)
	})
	if err != nil {
		return nil, err
	}

	if assetExist {
		if _, err = mm.updateAsset(ctx, currAsset, descriptor); err != nil {
			return nil, err
		}
		if err = mm.store.Assets().SoftDeleteExcludeById(ctx, ver.ID, currAsset.ID); err != nil {
			return nil, err
		}
		return currAsset, err
	}

	newAsset := &types.ArtifactAsset{
		ViewID:    null.IntFrom(mm.view.ViewID),
		VersionID: null.IntFrom(ver.ID),
	}
	if err = mm.createAsset(ctx, newAsset, descriptor); err != nil {
		return nil, err
	}

	if err = mm.store.Assets().SoftDeleteExcludeById(ctx, ver.ID, newAsset.ID); err != nil {
		return nil, err
	}
	return newAsset, nil
}

func (mm *ModelManager) createNormalAsset(ctx context.Context, ver *types.ArtifactVersion, descriptor *AssetDescriptor) (*types.ArtifactAsset, error) {
	currAsset, assetExist, err := mm.getExistAsset(func() (*types.ArtifactAsset, error) {
		return mm.store.Assets().GetVersionAsset(ctx, descriptor.Path, ver.ID)
	})
	if err != nil {
		return nil, err
	}

	if assetExist {
		return mm.updateAsset(ctx, currAsset, descriptor)
	}

	newAsset := &types.ArtifactAsset{
		ViewID:    null.IntFrom(mm.view.ViewID),
		VersionID: null.IntFrom(ver.ID),
	}
	if err = mm.createAsset(ctx, newAsset, descriptor); err != nil {
		return nil, err
	}
	return newAsset, nil
}

func (mm *ModelManager) getExistAsset(mutateFunc func() (*types.ArtifactAsset, error)) (*types.ArtifactAsset, bool, error) {
	var assetExist bool
	currAsset, err := mutateFunc()
	if err == nil {
		assetExist = true
	} else if errors.Is(err, gitfox_store.ErrResourceNotFound) {
		assetExist = false
	} else {
		return nil, assetExist, err
	}
	return currAsset, assetExist, nil
}

func (mm *ModelManager) updateAsset(ctx context.Context, current *types.ArtifactAsset, descriptor *AssetDescriptor) (*types.ArtifactAsset, error) {
	now := time.Now().UnixMilli()
	if err := mm.validateAsset(ctx, current, descriptor); err != nil {
		return nil, err
	}

	newBlob, err := mm.createBlob(ctx, descriptor, now)
	if err != nil {
		return nil, err
	}

	previousBlobId := current.BlobID

	current.BlobID = newBlob.ID
	if descriptor.Hash != nil {
		current.CheckSum = descriptor.Hash.String()
	}

	if err = mm.store.Assets().Update(ctx, current); err != nil {
		return nil, err
	}

	if err = mm.store.Blobs().SoftDeleteById(ctx, previousBlobId); err != nil {
		return nil, err
	}
	return current, nil
}

func (mm *ModelManager) createAsset(ctx context.Context, newObj *types.ArtifactAsset, descriptor *AssetDescriptor) error {
	now := time.Now().UnixMilli()
	newBlob, err := mm.createBlob(ctx, descriptor, now)
	if err != nil {
		return err
	}

	if err = fillUpAsset(newObj, descriptor, newBlob.ID, now); err != nil {
		return err
	}

	if err = mm.store.Assets().Create(ctx, newObj); err != nil {
		return err
	}
	return nil
}

func fillUpAsset(asset *types.ArtifactAsset, descriptor *AssetDescriptor, blobId, createTime int64) error {
	asset.Path = descriptor.Path
	asset.Format = descriptor.Format
	asset.Kind = descriptor.Kind
	asset.ContentType = descriptor.ContentType
	asset.BlobID = blobId
	asset.Created = createTime

	if descriptor.Hash != nil {
		asset.CheckSum = descriptor.Hash.String()
	}

	if descriptor.Metadata != nil {
		metadata, err := descriptor.Metadata.ToJSON()
		if err != nil {
			return err
		}
		asset.Metadata = string(metadata)
	}
	return nil
}

func (mm *ModelManager) createBlob(ctx context.Context, descriptor *AssetDescriptor, createTime int64) (*types.ArtifactBlob, error) {
	blobModel := &types.ArtifactBlob{
		StorageID: mm.view.StorageID,
		Ref:       descriptor.Ref,
		Size:      descriptor.Size,
		Creator:   mm.creator.ID,
		Created:   createTime,
	}
	if err := mm.store.Blobs().Create(ctx, blobModel); err != nil {
		return nil, err
	}
	return blobModel, nil
}

// validateAsset will check whether the uploaded file is consistent with the existing file
func (mm *ModelManager) validateAsset(ctx context.Context, asset *types.ArtifactAsset, descriptor *AssetDescriptor) error {
	var h Hash
	if err := json.Unmarshal([]byte(asset.CheckSum), &h); err != nil {
		log.Ctx(ctx).Err(err).Msg("unmarshal1 checksum failed")
		return nil
	}
	if h.Sha256 == descriptor.Hash.Sha256 {
		return ErrStorageFileNotChanged
	}
	return nil
}

func (mm *ModelManager) getOrCreateVersion(ctx context.Context) (*types.ArtifactVersion, error) {
	if mm.versionModel != nil {
		return mm.versionModel, nil
	}

	if mm.pkgModel == nil {
		pkg, err := mm.getOrCreatePackage(ctx)
		if err != nil {
			return nil, err
		}
		mm.pkgModel = pkg
	}

	defer func() {
		if mm.pkgModel != nil && mm.versionModel != nil {
			_ = model.AddTreeNode(ctx, mm.store, mm.pkgModel, mm.versionModel)
		}
	}()

	currVersion, err := mm.store.Versions().GetByVersion(ctx, mm.pkgModel.ID, mm.view.ViewID, mm.pkgDescriptor.Version)
	if err == nil {
		if currVersion.IsDeleted() {
			if e := mm.store.Versions().UnDelete(ctx, currVersion); e != nil {
				return nil, e
			}
		}
		mm.versionModel = currVersion
		return currVersion, nil
	}

	if !errors.Is(err, gitfox_store.ErrResourceNotFound) {
		return nil, err
	}

	// create new version
	verModel := &types.ArtifactVersion{
		PackageID: mm.pkgModel.ID,
		Version:   mm.pkgDescriptor.Version,
		ViewID:    mm.view.ViewID,
	}

	if mm.pkgDescriptor.VersionMetadata != nil {
		metadata, e := mm.pkgDescriptor.VersionMetadata.ToJSON()
		if e == nil {
			verModel.Metadata = string(metadata)
		} else {
			log.Ctx(ctx).Err(err).Msg("convert version metadata to json failed")
		}
	}

	if err = mm.store.Versions().Create(ctx, verModel); err != nil {
		return nil, err
	}
	mm.versionModel = verModel
	return verModel, nil
}

func (mm *ModelManager) updateVerTime(ctx context.Context, verModel *types.ArtifactVersion, assetModel *types.ArtifactAsset) error {
	if verModel.ID != assetModel.VersionID.Int64 {
		return nil
	}

	verModel.Updated = assetModel.Updated
	return mm.store.Versions().Update(ctx, verModel)
}

func (mm *ModelManager) getOrCreatePackage(ctx context.Context) (*types.ArtifactPackage, error) {
	if mm.pkgModel != nil {
		return mm.pkgModel, nil
	}

	desc := mm.pkgDescriptor
	currPkg, err := mm.store.Packages().GetByName(ctx, desc.Name, desc.Namespace, mm.view.OwnerID, desc.Format)
	if err == nil {
		if currPkg.IsDeleted() {
			if e := mm.store.Packages().UnDelete(ctx, currPkg); e != nil {
				return nil, e
			}
		}
		mm.pkgModel = currPkg
		return currPkg, nil
	}

	if !errors.Is(err, gitfox_store.ErrResourceNotFound) {
		return nil, err
	}

	pkgModel := &types.ArtifactPackage{
		OwnerID:   mm.view.OwnerID,
		Name:      desc.Name,
		Namespace: desc.Namespace,
		Format:    desc.Format,
	}
	if err = mm.store.Packages().Create(ctx, pkgModel); err != nil {
		return nil, err
	}
	mm.pkgModel = pkgModel
	return pkgModel, nil
}
