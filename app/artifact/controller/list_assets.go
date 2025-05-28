// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"net/url"
	"strings"

	url2 "github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/types"
)

func (c *Controller) ListAssets(ctx context.Context, req *BaseReq,
	packageName, group, version string, format types.ArtifactFormat,
) ([]*types.ArtifactAssetsRes, error) {
	dbVer, err := c.artStore.GetVersion(ctx, req.view.Space.ID, req.view.ViewID, packageName, group, version, format)
	if err != nil {
		return nil, err
	}

	data, err := c.artStore.ListAssets(ctx, dbVer.ID)
	if err != nil {
		return nil, err
	}

	err = addLink(data, req.spaceName, packageName, group, version, format)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type linkFunc func(format types.ArtifactFormat, spaceName, pkgName, group, version, filePath string) (string, error)

var (
	linkRawFunc linkFunc = func(format types.ArtifactFormat, spaceName, pkgName, group, version, filePath string) (string, error) {
		pathSegments := []string{url2.ArtifactMount, spaceName, string(format)}

		if group != "" {
			groups := strings.Split(group, ".")
			pathSegments = append(pathSegments, groups...)
		}

		pathSegments = append(pathSegments, pkgName, version, filePath)
		return url.JoinPath("/", pathSegments...)
	}

	linkHelmFunc linkFunc = func(format types.ArtifactFormat, spaceName, pkgName, group, version, filePath string) (string, error) {
		pathSegments := []string{url2.ArtifactMount, spaceName, string(format)}
		pathSegments = append(pathSegments, filePath)
		return url.JoinPath("/", pathSegments...)
	}
)

func addLink(data []*types.ArtifactAssetsRes, spaceName, packageName, group, version string, format types.ArtifactFormat) error {
	var fn linkFunc
	switch format {
	case types.ArtifactRawFormat:
		fn = linkRawFunc
	case types.ArtifactHelmFormat:
		fn = linkHelmFunc
	default:
		fn = linkRawFunc
	}

	for _, res := range data {
		link, err := fn(format, spaceName, packageName, group, version, res.Path)
		if err != nil {
			return err
		}
		res.Link = link
	}
	return nil
}
