// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/easysoft/gitfox/types"
)

func (c *Controller) ListTree(ctx context.Context, req *BaseReq, filter *types.ArtifactTreeFilter) ([]*types.ArtifactTreeRes, error) {
	var err error
	nodes := make([]*types.ArtifactTreeRes, 0)
	var nodeModules []*types.ArtifactTreeNode

	if filter.Format == types.ArtifactAllFormat {
		nodeModules, err = c.artStore.Nodes().ListFormats(ctx, req.Space().ID)
		if err != nil {
			return nodes, err
		}
		return convertTreeNodes(nodeModules, "", nil), nil
	}

	currNode, err := c.artStore.Nodes().GetByPath(ctx, req.Space().ID, filter.Path, filter.Format)
	if err != nil {
		return nodes, err
	}

	if currNode.Type == types.ArtifactTreeNodeTypeVersion &&
		currNode.Format != types.ArtifactContainerFormat {
		return c.listAssetAsNodes(ctx, req, filter.Path, filter.Format)
	}

	nodeModules, err = c.artStore.Nodes().ListByParentId(ctx, currNode.ID)

	if err != nil {
		return nodes, err
	}

	getVerFunc := func(in *types.ArtifactTreeNodeMeta) *types.ArtifactVersion {
		pkg, e := c.artStore.Packages().GetByName(ctx, in.Name, in.Group, req.Space().ID, filter.Format)
		if e != nil {
			return nil
		}
		ver, e := c.artStore.Versions().GetByVersion(ctx, pkg.ID, req.view.ViewID, in.Version)
		if e != nil {
			return nil
		}
		return ver
	}

	return convertTreeNodes(nodeModules, filter.Level, getVerFunc), nil
}

func (c *Controller) listAssetAsNodes(ctx context.Context, req *BaseReq, path string, format types.ArtifactFormat) ([]*types.ArtifactTreeRes, error) {
	info := parseArtifactPath(path)

	assets, err := c.ListAssets(ctx, req, info.Name, info.Group, info.Version, format)
	if err != nil {
		return []*types.ArtifactTreeRes{}, err
	}

	data := make([]*types.ArtifactTreeRes, 0)
	for _, asset := range assets {
		n := types.ArtifactTreeRes{
			Name:     filepath.Base(asset.Path),
			Path:     path + "/" + filepath.Base(asset.Path),
			Leaf:     true,
			Format:   format,
			Metadata: info,
		}
		assetInfo := *info
		assetInfo.Type = types.ArtifactTreeNodeTypeAsset

		nid := &types.ArtifactTreeNodeId{Type: types.ArtifactTreeNodeTypeAsset, Pk: asset.Id}
		assetInfo.NodeId = nid.String()

		n.Metadata = &assetInfo
		data = append(data, &n)
	}
	return data, nil
}

func convertTreeNodes(inputs []*types.ArtifactTreeNode, level string, getVerObj func(in *types.ArtifactTreeNodeMeta) *types.ArtifactVersion) []*types.ArtifactTreeRes {
	nodes := make([]*types.ArtifactTreeRes, 0)
	sort.Slice(inputs, func(i, j int) bool {
		x := inputs[i]
		y := inputs[j]
		return x.Type == types.ArtifactTreeNodeTypeDirectory && x.Name < y.Name
	})

	for _, i := range inputs {
		n := types.ArtifactTreeRes{
			Name:   i.Name,
			Path:   i.Path,
			Format: i.Format,
		}
		if i.Type == types.ArtifactTreeNodeTypeVersion {
			switch level {
			case types.ArtifactNodeLevelVersion:
				n.Leaf = true
				n.Metadata = parseArtifactPath(i.Path)
				n.Metadata.Type = types.ArtifactTreeNodeTypeVersion
			case types.ArtifactNodeLevelAsset:
				if i.Format == types.ArtifactContainerFormat {
					n.Leaf = true
				}
				n.Metadata = parseArtifactPath(i.Path)
				n.Metadata.Type = types.ArtifactTreeNodeTypeVersion
				ver := getVerObj(n.Metadata)
				if ver == nil {
					continue
				}
				nid := &types.ArtifactTreeNodeId{Type: types.ArtifactTreeNodeTypeVersion, Pk: ver.ID}
				n.Metadata.NodeId = nid.String()
			}
		} else if i.Type == types.ArtifactTreeNodeTypeDirectory {
			nid := &types.ArtifactTreeNodeId{Type: types.ArtifactTreeNodeTypeDirectory, Pk: i.ID}
			n.Metadata = &types.ArtifactTreeNodeMeta{
				Type: types.ArtifactTreeNodeTypeDirectory, NodeId: nid.String(),
			}
		}

		nodes = append(nodes, &n)
	}
	return nodes
}

func parseArtifactPath(path string) *types.ArtifactTreeNodeMeta {
	frames := strings.Split(strings.TrimLeft(path, "/"), "/")
	count := len(frames)
	if count < 2 {
		return nil
	}
	var namespace string
	version := frames[count-1]
	pkgName := frames[count-2]
	if count > 2 {
		namespace = strings.Join(frames[:count-2], ".")
	}

	return &types.ArtifactTreeNodeMeta{
		Name:    pkgName,
		Group:   namespace,
		Version: version,
	}
}
