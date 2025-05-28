// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"sort"

	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

func (c *Controller) StatisticContainerCapacity(ctx context.Context) (*types.ArtifactStatisticResponse, error) {
	// statistic asset usage
	report, err := c.gcSvc.StatisticContainer(ctx, 0)
	if err != nil {
		return nil, err
	}

	//
	return c.statistic(ctx, report)
}

func (c *Controller) StatisticCapacity(ctx context.Context) (*types.ArtifactStatisticResponse, error) {
	report, err := c.gcSvc.StatisticExcludeContainer(ctx, 0)
	if err != nil {
		return nil, err
	}

	return c.statistic(ctx, report)
}

func (c *Controller) statistic(ctx context.Context, report *types.ArtifactStatisticReport) (*types.ArtifactStatisticResponse, error) {
	res := &types.ArtifactStatisticResponse{}

	delData := &types.ArtifactGarbageReportRes{Ids: make([]int64, 0)}
	for _, d := range report.DeleteList {
		delData.Count += 1
		delData.Size += d.Size
		delData.Ids = append(delData.Ids, d.AssetId)
	}
	res.GarbageCollect = delData

	if len(report.TagList) == 0 {
		res.Capacity = make([]*types.ArtifactPackageCapacityRes, 0)
		return res, nil
	}

	sort.Slice(report.TagList, func(i, j int) bool {
		return report.TagList[i].VersionId < report.TagList[j].VersionId
	})

	vids := make([]int64, 0)
	for _, t := range report.TagList {
		vids = append(vids, t.VersionId)
	}

	verInfoList, err := c.artStore.Versions().Search(ctx, versionIdsOpt{ids: vids})
	if err != nil {
		return nil, err
	}

	pkgCapMap := make(map[int64]*types.ArtifactPackageCapacityRes)
	for id, verInfo := range verInfoList {
		pkgCap, ok := pkgCapMap[verInfo.PackageId]
		if !ok {
			pkgCap = &types.ArtifactPackageCapacityRes{
				Space: verInfo.SpaceName, Name: verInfo.PackageName, Namespace: verInfo.PackageNamespace,
				Format: verInfo.PackageFormat, Versions: make([]types.ArtifactVersionCapacityRes, 0),
			}
			pkgCapMap[verInfo.PackageId] = pkgCap
		}

		tagCap := report.TagList[id]

		verCap := types.ArtifactVersionCapacityRes{
			Version: verInfo.Version,
			Size:    tagCap.TotalSize, ExclusiveSize: tagCap.ExclusiveSize,
		}
		pkgCap.Versions = append(pkgCap.Versions, verCap)
		pkgCap.Size += tagCap.TotalSize
		pkgCap.ExclusiveSize += tagCap.ExclusiveSize
	}

	pkgCapList := make([]*types.ArtifactPackageCapacityRes, 0)
	for _, pkgCap := range pkgCapMap {
		sort.Slice(pkgCap.Versions, func(i, j int) bool {
			return pkgCap.Versions[i].ExclusiveSize > pkgCap.Versions[j].ExclusiveSize
		})
		pkgCapList = append(pkgCapList, pkgCap)
	}
	sort.Slice(pkgCapList, func(i, j int) bool {
		return pkgCapList[i].Size > pkgCapList[j].Size
	})

	res.Capacity = pkgCapList

	return res, nil
}

type versionIdsOpt struct {
	ids []int64
}

func (o versionIdsOpt) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("version_id IN ?", o.ids).Order("version_id asc")
}
