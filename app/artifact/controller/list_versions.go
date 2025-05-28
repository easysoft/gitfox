// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
)

func (c *Controller) ListVersions(ctx context.Context, req *BaseReq, format types.ArtifactFormat,
	filter *types.ArtifactVersionFilter,
) ([]*types.ArtifactVersionsRes, error) {
	dbPkg, err := c.artStore.Packages().GetByName(ctx, filter.Package, filter.Group, req.view.OwnerID, format)
	if err != nil {
		return nil, err
	}

	verFilter := types.SearchVersionOption{
		PackageId: dbPkg.ID,
		ViewId:    req.view.ViewID,
		Page:      filter.Page,
		Size:      filter.Size,
		Query:     filter.Query,
	}

	versions, err := c.artStore.Versions().Find(ctx, verFilter)
	if err != nil {
		return nil, err
	}

	data := make([]*types.ArtifactVersionsRes, 0)
	for _, version := range versions {
		item := &types.ArtifactVersionsRes{
			Version: version.Version, CreatorName: auth.AnonymousPrincipal.UID, Updated: version.Updated,
		}
		if version.Metadata != "" {
			var meta adapter.VersionMetadata
			err = json.NewDecoder(bytes.NewReader([]byte(version.Metadata))).Decode(&meta)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("parse version metadata failed")
			} else {
				item.CreatorName = meta.CreatorName
			}
		}
		data = append(data, item)
	}

	return data, nil
}
