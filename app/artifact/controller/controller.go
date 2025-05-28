// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"strings"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/services/artifactgc"
	"github.com/easysoft/gitfox/app/services/settings"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

type Controller struct {
	tx          dbtx.Transactor
	urlProvider url.Provider
	authorizer  authz.Authorizer
	artStore    store.ArtifactStore
	spaceStore  store.SpaceStore
	fileStore   storage.ContentStorage
	settings    *settings.Service

	gcSvc *artifactgc.Service
}

func NewController(
	tx dbtx.Transactor, urlProvider url.Provider, authorizer authz.Authorizer,
	artStore store.ArtifactStore,
	spaceStore store.SpaceStore,
	fileStore storage.ContentStorage,
	settings *settings.Service,
	gcSvc *artifactgc.Service,
) *Controller {
	return &Controller{
		tx:          tx,
		urlProvider: urlProvider,
		authorizer:  authorizer,
		artStore:    artStore,
		spaceStore:  spaceStore,
		fileStore:   fileStore,
		settings:    settings,
		gcSvc:       gcSvc,
	}
}

type RequestView struct {
	Space     *types.Space
	View      *types.ArtifactView
	StorageId int64
}

func (c *Controller) ParseSpaceView(ctx context.Context, spaceRef string) (*adapter.ViewDescriptor, error) {
	viewName := ""
	spaceKey := spaceRef

	if idx := strings.LastIndex(spaceRef, "@"); idx >= 0 {
		spaceKey = spaceRef[:idx]
		viewName = spaceRef[idx+1:]
	}

	space, err := c.spaceStore.FindByRef(ctx, spaceKey)
	if err != nil {
		return nil, err
	}

	var view *types.ArtifactView

	if viewName == "" {
		view, err = c.artStore.Views().GetDefault(ctx, space.ID)
	} else {
		view, err = c.artStore.Views().GetByName(ctx, space.ID, viewName)
	}

	if err != nil {
		return nil, err
	}

	return &adapter.ViewDescriptor{
		ViewID: view.ID, OwnerID: space.ID, Space: space,
		Store: c.fileStore, StorageID: 1}, nil
}

func (c *Controller) checkAuthArtifactPush(
	ctx context.Context,
	req ArtifactAuthRequest,
) error {
	rdonly, err := settings.SystemGet(ctx, c.settings,
		settings.ContainerReadOnly, false,
	)
	if err != nil {
		return err
	}
	if rdonly {
		return fmt.Errorf("registry storage is readonly, push denied")
	}
	// create is a special case - check permission without specific resource
	err = apiauth.CheckSpaceScope(
		ctx,
		c.authorizer,
		req.Session(),
		req.Space(),
		enum.ResourceTypeArtifact,
		enum.PermissionArtifactPush,
	)
	if err != nil {
		return fmt.Errorf("auth check failed: %w", err)
	}

	return nil
}

func (c *Controller) checkAuthArtifactPull(
	ctx context.Context,
	req ArtifactAuthRequest,
) error {
	// always verified
	if true {
		return nil
	}
	// create is a special case - check permission without specific resource
	err := apiauth.CheckSpaceScope(
		ctx,
		c.authorizer,
		req.Session(),
		req.Space(),
		enum.ResourceTypeArtifact,
		enum.PermissionArtifactPull,
	)
	if err != nil {
		return fmt.Errorf("auth check failed: %w", err)
	}

	return nil
}

func (c *Controller) checkAuthArtifactDelete(
	ctx context.Context,
	req ArtifactAuthRequest,
) error {
	// create is a special case - check permission without specific resource
	err := apiauth.CheckSpaceScope(
		ctx,
		c.authorizer,
		req.Session(),
		req.Space(),
		enum.ResourceTypeArtifact,
		enum.PermissionArtifactDelete,
	)
	if err != nil {
		return fmt.Errorf("auth check failed: %w", err)
	}

	return nil
}
