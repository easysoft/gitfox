// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"

	"github.com/easysoft/gitfox/app/artifact/adapter/container"

	"github.com/opencontainers/go-digest"
)

func (c *Controller) ListContainerImages(ctx context.Context, req *BaseReq, packageName, version string) (*container.TagMetadata, error) {
	t, err := findContainerTag(ctx, c.artStore, req.view.ViewID, req.view.OwnerID, packageName, version)
	if err != nil {
		return nil, err
	}

	dgst, err := digest.Parse(t.asset.Path)
	if err != nil {
		return nil, err
	}
	tagMetaReader := container.NewTagMetadataReader(c.artStore, req.view)
	data, err := tagMetaReader.Parse(ctx, t.asset.ContentType, dgst)
	if err != nil {
		return nil, err
	}
	return data, nil
}
