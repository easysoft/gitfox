// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifactgc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/container/gc"
	"github.com/easysoft/gitfox/app/services/settings"
	"github.com/easysoft/gitfox/app/sse"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database/artifacts"
	"github.com/easysoft/gitfox/job"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Service struct {
	tx        dbtx.Transactor
	artStore  store.ArtifactStore
	fileStore storage.ContentStorage
	settings  *settings.Service

	scheduler   *job.Scheduler
	sseStreamer sse.Streamer
}

func NewService(
	tx dbtx.Transactor,
	artStore store.ArtifactStore,
	fileStore storage.ContentStorage,
	settings *settings.Service,
	scheduler *job.Scheduler,
	sseStreamer sse.Streamer,
) *Service {
	return &Service{
		tx:          tx,
		artStore:    artStore,
		fileStore:   fileStore,
		settings:    settings,
		scheduler:   scheduler,
		sseStreamer: sseStreamer,
	}
}

func (s *Service) StatisticContainer(ctx context.Context, retentionTime time.Duration) (*types.ArtifactStatisticReport, error) {
	return s.statisticContainer(ctx, retentionTime)
}

func (s *Service) statisticContainer(ctx context.Context, retentionTime time.Duration) (*types.ArtifactStatisticReport, error) {
	statisticData, err := gc.Statistic(ctx, s.artStore, s.fileStore, retentionTime)
	if err != nil {
		return nil, err
	}
	return statisticData, nil
}

func (s *Service) GarbageCollectContainer(ctx context.Context, retentionTime time.Duration) (*types.ArtifactGarbageReportRes, error) {
	if err := s.setContainerReadonlySetting(ctx, true); err != nil {
		return nil, err
	}
	defer func() {
		_ = s.setContainerReadonlySetting(ctx, false)
	}()

	statisticData, err := s.statisticContainer(ctx, retentionTime)
	if err != nil {
		return nil, err
	}
	return s.garbageCollectContainer(ctx, statisticData.DeleteList)
}

func (s *Service) garbageCollectContainer(ctx context.Context, delList []*types.ArtifactRecycleBlobDesc) (*types.ArtifactGarbageReportRes, error) {
	recycleRes, err := s.garbageCollect(ctx, KindContainer, delList)
	return recycleRes, err
}
func (s *Service) garbageCollect(ctx context.Context, kind Kind, assets []*types.ArtifactRecycleBlobDesc) (*types.ArtifactGarbageReportRes, error) {
	logger := log.Ctx(ctx)
	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("kind", string(kind))
	})

	gcData := &types.ArtifactGarbageReportRes{}
	for _, asset := range assets {
		err := s.tx.WithTx(ctx, func(ctx context.Context) error {
			if e := s.artStore.Assets().DeleteById(ctx, asset.AssetId); e != nil {
				return e
			}
			if e := s.artStore.Blobs().DeleteById(ctx, asset.BlobId); e != nil {
				return e
			}
			if e := s.fileStore.Delete(ctx, adapter.BlobPath(asset.BlobRef)); e != nil {
				return e
			}
			gcData.Count += 1
			gcData.Size += asset.Size
			logger.Info().Msgf("remove asset %d, blob %d, file %s", asset.AssetId, asset.BlobId, adapter.BlobPath(asset.BlobRef))
			return nil
		})
		if err != nil {
			return gcData, err
		}
	}
	return gcData, nil
}

func (s *Service) setContainerReadonlySetting(ctx context.Context, v bool) error {
	return s.settings.SystemSet(ctx, settings.ContainerReadOnly, v)
}

/*
StatisticExcludeContainer returns:
  - An array of all soft-removed assets, without container format.
  - An array of capacity usage for each version, excluding container format.
*/
func (s *Service) StatisticExcludeContainer(ctx context.Context, retentionTime time.Duration) (*types.ArtifactStatisticReport, error) {
	searchOpts := []store.SearchOption{
		excludeContainerOpt{},
	}

	assetBlobs, err := s.artStore.Assets().SearchExtendBlob(ctx, searchOpts...)
	if err != nil {
		return nil, err
	}

	delList := make([]*types.ArtifactRecycleBlobDesc, 0)
	versionMap := make(map[int64]*types.ArtifactVersionCapacityDesc)

	deleteBefore := time.Now().Add(-retentionTime).UnixMilli()
	for _, asset := range assetBlobs {
		if asset.Deleted > 0 && asset.Deleted < deleteBefore {
			delList = append(delList, &types.ArtifactRecycleBlobDesc{
				AssetId: asset.ID,
				BlobId:  asset.BlobID,
				BlobRef: asset.Ref,
				Size:    asset.Size,
			})
			continue
		}

		verId := asset.VersionID.Int64
		verDesc, ok := versionMap[verId]
		if !ok {
			versionMap[verId] = &types.ArtifactVersionCapacityDesc{
				VersionId:     verId,
				ExclusiveSize: asset.Size,
				TotalSize:     asset.Size,
			}
			continue
		}

		verDesc.ExclusiveSize += asset.Size
		verDesc.TotalSize += asset.Size
	}

	// empty array while retentionTime greater than zero
	verList := make([]*types.ArtifactVersionCapacityDesc, 0)
	for _, item := range versionMap {
		verList = append(verList, item)
	}

	return &types.ArtifactStatisticReport{DeleteList: delList, TagList: verList}, nil
}

func (s *Service) GarbageCollectSoftRemove(ctx context.Context, retentionTime time.Duration) (*types.ArtifactGarbageReportRes, error) {
	recycleBeforeTime := time.Now().Add(-retentionTime)

	statisticData, err := s.StatisticExcludeContainer(ctx, retentionTime)
	if err != nil {
		return nil, err
	}

	assetRecycleRes, err := s.garbageCollect(ctx, KindSoftRemove, statisticData.DeleteList)
	if err != nil {
		return nil, err
	}

	if err = s.recycleSoftRemovedPkgAndVersion(ctx, recycleBeforeTime); err != nil {
		return nil, err
	}

	return assetRecycleRes, nil
}

func (s *Service) recycleSoftRemovedPkgAndVersion(ctx context.Context, before time.Time) error {
	versions, err := s.artStore.Versions().Search(ctx, artifacts.VersionWithDeletedBeforeOption{Before: before})
	if err != nil {
		return err
	}
	vIds := make([]int64, 0, len(versions))
	for id, version := range versions {
		vIds[id] = version.ID
	}

	packages, err := s.artStore.Packages().Search(ctx, artifacts.PackageWithDeletedBeforeOption{Before: before})
	if err != nil {
		return err
	}
	pIds := make([]int64, 0, len(packages))
	for id, pkg := range packages {
		pIds[id] = pkg.ID
	}

	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if len(vIds) > 0 {
			log.Ctx(ctx).Info().Msgf("recycle soft-removed versions %v", vIds)
			if _, e := s.artStore.Versions().DeleteByIds(ctx, vIds...); e != nil {
				return e
			}
		}

		if len(pIds) > 0 {
			log.Ctx(ctx).Info().Msgf("recycle soft-removed packages %v", pIds)
			if _, e := s.artStore.Packages().DeleteByIds(ctx, pIds...); e != nil {
				return e
			}
		}

		return nil
	})

	return err
}

func (s *Service) Trigger(ctx context.Context, input *Input) error {
	j, err := getJobDef(input)
	if err != nil {
		return err
	}
	return s.scheduler.RunJob(ctx, j)
}

func getJobDef(in *Input) (job.Definition, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return job.Definition{}, fmt.Errorf("failed to marshal job input json: %w", err)
	}

	j := job.Definition{
		MaxRetries: 1,
		Data:       base64.StdEncoding.EncodeToString(data),
	}

	if in.JobUID == "" {
		in.JobUID = fmt.Sprintf("artifact-gc-%d", time.Now().UnixNano())
	}
	j.UID = in.JobUID

	switch in.Kind {
	case KindContainer:
		j.Timeout = JobMaxDurationDArtifactContainerGC
		j.Type = JobTypeArtifactContainerGC
	default:
		j.Timeout = JobMaxDurationDArtifactSoftRemove
		j.Type = JobTypeArtifactHardRemove
	}

	return j, nil
}
