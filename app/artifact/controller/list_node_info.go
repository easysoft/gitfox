// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	"github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/opencontainers/go-digest"
)

type ListNodeInfoRequest struct {
	NodeIds []string `json:"node_ids"`
}

func (c *Controller) ListNodeInfo(ctx context.Context, in *ListNodeInfoRequest) ([]*types.ArtifactNodeInfo, error) {
	result := make([]*types.ArtifactNodeInfo, 0)

	for _, node := range in.NodeIds {
		if strings.Count(node, ".") != 1 {
			return nil, errors.New("invalid node_id")
		}

		frames := strings.Split(node, ".")
		nodeType := types.ArtifactTreeNodeType(frames[0])
		nodePk, err := strconv.ParseInt(frames[1], 10, 64)
		if err != nil {
			return nil, err
		}

		var space *types.Space
		var pkg *types.ArtifactPackage
		var ver *types.ArtifactVersion
		var assetRes *types.ArtifactAssetsRes
		info := types.ArtifactNodeInfo{}

		if nodeType == types.ArtifactTreeNodeTypeVersion {
			ver, err = c.artStore.Versions().GetByID(ctx, nodePk)
			if err != nil {
				if errors.Is(err, store.ErrResourceNotFound) {
					result = append(result, notFoundAsset())
					continue
				}
				return nil, err
			}
		} else if nodeType == types.ArtifactTreeNodeTypeAsset {
			assetRes, err = c.artStore.GetAsset(ctx, nodePk)
			if err != nil {
				if errors.Is(err, store.ErrResourceNotFound) {
					result = append(result, notFoundAsset())
					continue
				}
				return nil, err
			}
			asset, e := c.artStore.Assets().GetById(ctx, assetRes.Id)
			if e != nil {
				return nil, err
			}
			ver, err = c.artStore.Versions().GetByID(ctx, asset.VersionID.Int64)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("invalid node_type")
		}

		pkg, err = c.artStore.Packages().GetByID(ctx, ver.PackageID)
		if err != nil {
			return nil, err
		}

		space, err = c.spaceStore.Find(ctx, pkg.OwnerID)
		if err != nil {
			return nil, err
		}

		info.Format = pkg.Format
		info.Metadata = &types.ArtifactPkgVerMetadata{
			Space:   space.Identifier,
			Name:    pkg.Name,
			Group:   pkg.Namespace,
			Version: ver.Version,
		}

		if pkg.Format == types.ArtifactContainerFormat {
			t, e := findContainerTag(ctx, c.artStore, ver.ViewID, pkg.OwnerID, pkg.Name, ver.Version)
			if e != nil {
				return nil, e
			}

			dgst, e := digest.Parse(t.asset.Path)
			if e != nil {
				return nil, err
			}
			view, e := c.ParseSpaceView(ctx, space.Identifier)
			if e != nil {
				return nil, err
			}
			tagMetaReader := container.NewTagMetadataReader(c.artStore, view)
			data, e := tagMetaReader.Parse(ctx, t.asset.ContentType, dgst)
			if e != nil {
				return nil, err
			}
			for _, i := range data.Images {
				if i.OS == "linux" && i.Arch == "amd64" {
					info.Size = i.Size
					info.Created = t.version.Created
					info.Updated = t.asset.Created
				}
			}
		} else {
			if err = addLink([]*types.ArtifactAssetsRes{assetRes}, space.Identifier,
				pkg.Name, pkg.Namespace, ver.Version, pkg.Format); err != nil {
				return nil, err
			}
			info.Path = assetRes.Path
			info.ContentType = assetRes.ContentType
			info.Size = assetRes.Size
			info.Created = assetRes.Created
			info.Updated = assetRes.Updated
			info.CreatorName = assetRes.CreatorName
			info.Link = assetRes.Link
			info.CheckSum = assetRes.CheckSum
		}

		info.Status = "ok"
		result = append(result, &info)
	}

	return result, nil
}

func notFoundAsset() *types.ArtifactNodeInfo {
	return &types.ArtifactNodeInfo{
		Status: "not found",
	}
}
