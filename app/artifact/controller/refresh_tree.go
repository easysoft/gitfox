// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"

	"github.com/easysoft/gitfox/app/artifact/model"
	"github.com/easysoft/gitfox/types"
)

func (c *Controller) RefreshTree(ctx context.Context) error {
	dbSpaces, err := c.spaceStore.List(ctx, 0, &types.SpaceFilter{})
	if err != nil {
		return err
	}

	for _, space := range dbSpaces {
		dbPackages, e1 := c.artStore.Packages().List(ctx, space.ID, false)
		if e1 != nil {
			return e1
		}
		for _, pkg := range dbPackages {
			dbVersions, e2 := c.artStore.Versions().Find(ctx, types.SearchVersionOption{PackageId: pkg.ID})
			if e2 != nil {
				return e2
			}
			for _, pkgVersion := range dbVersions {
				if err = model.AddTreeNode(ctx, c.artStore, pkg, pkgVersion); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
