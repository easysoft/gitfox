// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"path"

	"github.com/easysoft/gitfox/app/artifact/adapter/helm"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
)

type IndexUpdater struct {
	helm bool
}

func (u *IndexUpdater) Run(ctx context.Context, store store.ArtifactStore, req *BaseReq) error {
	if u.helm {
		i := helm.NewHelmIndex(store, req.view)
		if err := i.UpdateRepo(ctx, path.Join("/", url.ArtifactMount, req.spaceName, "helm")); err != nil {
			return err
		}
	}

	return nil
}
