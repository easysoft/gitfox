// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package repo

import (
	"github.com/google/wire"
	"github.com/harness/gitness/internal/auth/authz"
	"github.com/harness/gitness/internal/gitrpc"
	"github.com/harness/gitness/internal/store"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	NewController,
)

func ProvideController(authorizer authz.Authorizer, spaceStore store.SpaceStore,
	repoStore store.RepoStore, saStore store.ServiceAccountStore, rpcClient gitrpc.Interface) *Controller {
	return NewController(authorizer, spaceStore, repoStore, saStore, rpcClient)
}
