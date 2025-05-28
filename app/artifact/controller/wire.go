// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/services/artifactgc"
	"github.com/easysoft/gitfox/app/services/settings"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/store/database/dbtx"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideArtifactController,
)

func ProvideArtifactController(tx dbtx.Transactor,
	urlProvider url.Provider, authorizer authz.Authorizer,
	artStore store.ArtifactStore, spaceStore store.SpaceStore, fileStore storage.ContentStorage,
	settings *settings.Service, gcSvc *artifactgc.Service,
) *Controller {
	return NewController(tx, urlProvider, authorizer, artStore, spaceStore, fileStore, settings, gcSvc)
}
