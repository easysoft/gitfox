// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package gc

import (
	"context"
	"fmt"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database/artifacts"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/types"

	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
)

type digestRef struct {
	count int
	obj   *types.ArtifactAssetExtendBlob
}

type manifestRefs struct {
	digests []digest.Digest
	obj     *types.ArtifactAssetExtendBlob
}

type TagDescriptor struct {
	AssetId       int64
	VersionId     int64
	ExclusiveSize int64
	TotalSize     int64
	ExcludeRefs   int
	TotalRefs     int
}

type StatisticReport struct {
	DeleteList []*types.ArtifactRecycleBlobDesc
	TagList    []*TagDescriptor
}

func Statistic(ctx context.Context, artStore store.ArtifactStore, fileStore storage.ContentStorage, rententionTime time.Duration) (*types.ArtifactStatisticReport, error) {
	allAssets, err := artStore.Assets().SearchExtendBlob(ctx, ContainerFormatOption{})
	if err != nil {
		return nil, err
	}
	logger := log.Ctx(ctx)

	digestRefsMap := make(map[digest.Digest]*digestRef)
	for _, asset := range allAssets {
		dgst, e := digest.Parse(asset.Path)
		if e != nil {
			return nil, e
		}
		logger.Debug().Msgf("adding digest to refs map: %s", dgst)
		digestRefsMap[dgst] = &digestRef{count: 0, obj: asset}
	}

	manifestGetter := &manifestStore{
		artStore: artStore, fileStore: fileStore,
		blobGetFn: func(dgst digest.Digest) (*types.ArtifactAssetExtendBlob, bool) {
			ref, ok := digestRefsMap[dgst]
			if !ok {
				return nil, false
			}
			return ref.obj, true
		},
	}

	tagSearchOpts := []store.SearchOption{TagAssetOption{}}
	if rententionTime > 0 {
		tagSearchOpts = append(tagSearchOpts, TagUndeletedAndRecentDeletedOption{After: time.Now().Add(-rententionTime)})
	} else {
		tagSearchOpts = append(tagSearchOpts, artifacts.AssetExcludeDeletedOption{})
	}
	taggedAssets, err := artStore.Assets().SearchExtendBlob(ctx, tagSearchOpts...)
	if err != nil {
		return nil, err
	}

	assetMarkMap := make(map[int64]*manifestRefs)

	for _, tagAsset := range taggedAssets {
		dgst, e := digest.Parse(tagAsset.Path)
		if e != nil {
			return nil, e
		}
		if _, ok := digestRefsMap[dgst]; !ok {
			return nil, fmt.Errorf("unknown tag asset: %s", dgst.String())
		}
		digestRefsMap[dgst].count += 1

		manifestRefObj := &manifestRefs{digests: make([]digest.Digest, 0), obj: tagAsset}
		assetMarkMap[tagAsset.ID] = manifestRefObj
		manifestRefObj.digests = append(manifestRefObj.digests, dgst)

		e = markManifestReferences(ctx, dgst, manifestGetter, func(d digest.Digest) {
			_, exist := digestRefsMap[d]
			if !exist {
				return
			}
			digestRefsMap[d].count += 1
			manifestRefObj.digests = append(manifestRefObj.digests, d)
		})
		if e != nil {
			return nil, e
		}
	}

	tags := make([]*types.ArtifactVersionCapacityDesc, 0)
	for _, assetMark := range assetMarkMap {
		t := &types.ArtifactVersionCapacityDesc{
			AssetId:   assetMark.obj.ID,
			VersionId: assetMark.obj.VersionID.Int64,
		}

		for _, ref := range assetMark.digests {
			d, e := digestRefsMap[ref]
			if !e {
				continue
			}
			if d.count == 1 {
				t.ExclusiveRefs += 1
				t.ExclusiveSize += d.obj.Size
			} else if d.count > 1 {
				d.count--
			}
			t.TotalRefs += 1
			t.TotalSize += d.obj.Size
		}
		tags = append(tags, t)
	}

	deleteList := make([]*types.ArtifactRecycleBlobDesc, 0)
	for dgst, d := range digestRefsMap {
		if d.count != 0 {
			continue
		}
		asset := d.obj
		delBlob := &types.ArtifactRecycleBlobDesc{
			AssetId: asset.ID,
			Path:    dgst.String(),
			BlobId:  asset.BlobID,
			BlobRef: asset.Ref,
			Size:    asset.Size,
		}
		deleteList = append(deleteList, delBlob)
	}

	return &types.ArtifactStatisticReport{DeleteList: deleteList, TagList: tags}, nil
}
