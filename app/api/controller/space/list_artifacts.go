// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package space

import (
	"context"
	"fmt"
	"strings"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

func (c *Controller) ListArtifacts(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	filter *types.ArtifactFilter,
) ([]*types.ArtifactListItem, int64, error) {
	space, view, err := parseView(ctx, spaceRef, c.spaceStore, c.artStore)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find parent space: %w", err)
	}

	err = apiauth.CheckSpace(ctx, c.authorizer, session, space, enum.PermissionSecretView)
	if err != nil {
		return nil, 0, fmt.Errorf("could not authorize: %w", err)
	}

	var packages []*types.ArtifactListItem
	var count int64

	packages, err = c.artStore.FindPackages(ctx, space.ID, view.ID, filter)
	if err != nil {
		return nil, 0, err
	}

	return packages, count, nil
}

func parseView(ctx context.Context, spaceRef string, spaceStore store.SpaceStore, artStore store.ArtifactStore) (*types.Space, *types.ArtifactView, error) {
	viewName := ""
	spaceKey := spaceRef

	if idx := strings.LastIndex(spaceRef, "@"); idx >= 0 {
		spaceKey = spaceRef[:idx]
		viewName = spaceRef[idx+1:]
	}

	space, err := spaceStore.FindByRef(ctx, spaceKey)
	if err != nil {
		return nil, nil, err
	}

	var view *types.ArtifactView

	if viewName == "" {
		view, err = artStore.Views().GetDefault(ctx, space.ID)
	} else {
		view, err = artStore.Views().GetByName(ctx, space.ID, viewName)
	}
	if err != nil {
		return nil, nil, err
	}
	return space, view, nil
}
