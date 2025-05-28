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

	"github.com/easysoft/gitfox/app/artifact/model"
	"github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
)

func (c *Controller) SoftRemove(ctx context.Context, req *BaseReq, in *ListNodeInfoRequest) (*types.ArtifactNodeRemoveReport, error) {
	var data []*types.ArtifactNodeRemoveRes
	var err error

	idx := &IndexUpdater{}

	if e := c.tx.WithTx(ctx, func(ctx context.Context) error {
		data, err = c.softRemoveNodes(ctx, req, in, idx)
		if err != nil {
			return err
		}
		err = idx.Run(ctx, c.artStore, req)
		return err
	}); e != nil {
		// sql rollback, nothing executed
		return nil, e
	}

	report := &types.ArtifactNodeRemoveReport{
		Total: len(in.NodeIds),
		Data:  data,
	}

	for _, item := range data {
		if item.Status == types.ArtifactStatusOK {
			report.Success++
		} else {
			report.Failed++
		}
	}
	return report, nil
}

func (c *Controller) softRemoveNodes(ctx context.Context, req *BaseReq, in *ListNodeInfoRequest, idx *IndexUpdater) ([]*types.ArtifactNodeRemoveRes, error) {
	data := make([]*types.ArtifactNodeRemoveRes, len(in.NodeIds))
	errorList := make([]error, 0)

OUTER:
	for id, node := range in.NodeIds {
		res := &types.ArtifactNodeRemoveRes{
			NodeId: node,
			Status: types.ArtifactStatusOK,
		}
		data[id] = res

		if strings.Count(node, ".") != 1 {
			res.Status = types.ArtifactStatusInvalidId
			continue
		}

		frames := strings.Split(node, ".")
		nodeType := types.ArtifactTreeNodeType(frames[0])
		nodePk, e := strconv.ParseInt(frames[1], 10, 64)
		if e != nil {
			res.Status = types.ArtifactStatusInvalidId
			continue
		}

		switch nodeType {
		case types.ArtifactNodeLevelAsset:
			asset, err := c.artStore.Assets().GetById(ctx, nodePk)
			if err != nil {
				if errors.Is(err, store.ErrResourceNotFound) {
					res.Status = types.ArtifactStatusNotFound
				} else {
					errorList = append(errorList, err)
					res.Status = types.ArtifactStatusUnknown
				}
				continue OUTER
			}
			err = c.softRemoveAsset(ctx, asset, res)
			if err != nil {
				errorList = append(errorList, err)
				res.Status = types.ArtifactStatusUnknown
				continue OUTER
			}
		case types.ArtifactNodeLevelVersion:
			ver, err := c.artStore.Versions().GetByID(ctx, nodePk)
			if err != nil {
				if errors.Is(err, store.ErrResourceNotFound) {
					res.Status = types.ArtifactStatusNotFound
				} else {
					errorList = append(errorList, err)
					res.Status = types.ArtifactStatusUnknown
				}
				continue OUTER
			}
			err = c.softRemoveVersion(ctx, ver, res, idx)
			if err != nil {
				errorList = append(errorList, err)
				res.Status = types.ArtifactStatusUnknown
				continue OUTER
			}
		case types.ArtifactTreeNodeTypeDirectory:
			treeNode, err := c.artStore.Nodes().GetById(ctx, nodePk)
			if err != nil {
				if errors.Is(err, store.ErrResourceNotFound) {
					res.Status = types.ArtifactStatusNotFound
				} else {
					errorList = append(errorList, err)
					res.Status = types.ArtifactStatusUnknown
				}
				continue OUTER
			}

			packages, err := c.findPackagesByNode(ctx, treeNode)
			if err != nil {
				errorList = append(errorList, err)
				res.Status = types.ArtifactStatusUnknown
				continue OUTER
			}
			for _, pkg := range packages {
				log.Ctx(ctx).Debug().Msgf("remove package node '%s' '%s'", pkg.Name, pkg.Namespace)
				err = c.softRemovePackage(ctx, pkg, req.view.ViewID, res, idx)
				if err != nil {
					errorList = append(errorList, err)
					res.Status = types.ArtifactStatusUnknown
					continue OUTER
				}
			}
			if err = c.artStore.Nodes().RecurseDeleteById(ctx, nodePk); err != nil {
				errorList = append(errorList, err)
				res.Status = types.ArtifactStatusUnknown
				continue OUTER
			}
			res.Status = types.ArtifactStatusOK
		default:
			res.Status = types.ArtifactStatusInvalidId
		}
	}

	if len(errorList) != 0 {
		return data, errors.New("node deletion is not completed")
	}
	return data, nil
}

func (c *Controller) findPackagesByNode(ctx context.Context, n *types.ArtifactTreeNode) ([]*types.ArtifactPackage, error) {
	if n.Path == "" {
		return nil, errors.New("invalid empty nodePath")
	}

	// [ [name, namespace]]
	querySets := make([][2]string, 0)
	parts := strings.Split(strings.TrimLeft(n.Path, "/"), "/")
	if len(parts) == 1 {
		querySets = append(querySets,
			[2]string{parts[0], ""},
			[2]string{"", parts[0]},
		)
	} else {
		count := len(parts)
		querySets = append(querySets, [2]string{parts[count-1], strings.Join(parts[0:count-1], ".")})
		querySets = append(querySets, [2]string{"", strings.Join(parts, ".")})
	}

	packages := make([]*types.ArtifactPackage, 0)
	for _, q := range querySets {
		name := q[0]
		namespace := q[1]
		if name != "" {
			dbPkg, err := c.artStore.Packages().GetByName(ctx, name, namespace, n.OwnerID, n.Format)
			if err == nil {
				packages = append(packages, dbPkg)
			}
		} else {
			objs, err := c.artStore.Packages().ListByNamespace(ctx, n.OwnerID, namespace, n.Format, true, false)
			if err == nil {
				packages = append(packages, objs...)
			}
		}
	}

	return packages, nil
}

func (c *Controller) softRemoveAsset(ctx context.Context, asset *types.ArtifactAsset, res *types.ArtifactNodeRemoveRes) error {
	var err error
	if err = c.artStore.Assets().SoftDeleteById(ctx, asset.ID); err != nil {
		return err
	}
	if asset.BlobID == 0 {
		return nil
	}

	if err = c.artStore.Blobs().SoftDeleteById(ctx, asset.BlobID); err != nil {
		return err
	}
	res.Assets += 1
	return nil
}

func (c *Controller) softRemoveVersion(ctx context.Context, ver *types.ArtifactVersion, res *types.ArtifactNodeRemoveRes, idx *IndexUpdater) error {
	// find assets of the version and soft remove them
	assets, e := c.artStore.ListAssets(ctx, ver.ID)
	if e != nil {
		return e
	}
	for _, asset := range assets {
		dbAsset, err := c.artStore.Assets().GetById(ctx, asset.Id)
		if err != nil {
			return err
		}
		if err = c.softRemoveAsset(ctx, dbAsset, res); err != nil {
			return err
		}
	}

	// soft remove the version
	if e = c.artStore.Versions().SoftDelete(ctx, ver); e != nil {
		return e
	}
	res.Versions++

	// find treeNode of the version and delete it
	dbPkg, e := c.artStore.Packages().GetByID(ctx, ver.PackageID)
	if e != nil {
		return e
	}
	if dbPkg.Format == types.ArtifactHelmFormat {
		idx.helm = true
	}
	nodePath, err := model.BuildPath(dbPkg.Namespace, dbPkg.Name, ver.Version)
	if err != nil {
		return err
	}
	node, err := c.artStore.Nodes().GetByPath(ctx, dbPkg.OwnerID, nodePath, dbPkg.Format)
	if err != nil {
		return err
	}
	if err = c.artStore.Nodes().DeleteById(ctx, node.ID); err != nil {
		return err
	}
	return nil
}

func (c *Controller) softRemovePackage(ctx context.Context, pkg *types.ArtifactPackage, viewId int64, res *types.ArtifactNodeRemoveRes, idx *IndexUpdater) error {
	// find versions of the package and soft remove them
	versions, err := c.artStore.Versions().Find(ctx, types.SearchVersionOption{PackageId: pkg.ID, ViewId: viewId})
	if err != nil {
		return err
	}

	for _, version := range versions {
		if err = c.softRemoveVersion(ctx, version, res, idx); err != nil {
			return err
		}
	}

	// soft remove the package
	if err = c.artStore.Packages().SoftDelete(ctx, pkg); err != nil {
		return err
	}
	res.Packages++

	if pkg.Format == types.ArtifactHelmFormat {
		idx.helm = true
	}
	return nil
}
