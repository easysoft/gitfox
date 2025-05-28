// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/store/database/dbtx"
)

type Controller struct {
	tx              dbtx.Transactor
	urlProvider     url.Provider
	authorizer      authz.Authorizer
	artifactStore   store.ArtifactStore
	spaceStore      store.SpaceStore
	repoStore       store.RepoStore
	principalStore  store.PrincipalStore
	membershipStore store.MembershipStore
}

func NewController(
	tx dbtx.Transactor,
	urlProvider url.Provider,
	authorizer authz.Authorizer,
	artifactStore store.ArtifactStore,
	spaceStore store.SpaceStore,
	repoStore store.RepoStore,
	principalStore store.PrincipalStore,
	membershipStore store.MembershipStore,
) *Controller {
	return &Controller{
		tx:              tx,
		urlProvider:     urlProvider,
		authorizer:      authorizer,
		artifactStore:   artifactStore,
		spaceStore:      spaceStore,
		repoStore:       repoStore,
		principalStore:  principalStore,
		membershipStore: membershipStore,
	}
}
