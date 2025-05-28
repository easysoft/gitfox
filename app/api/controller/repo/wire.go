// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"github.com/easysoft/gitfox/app/api/controller/limiter"
	"github.com/easysoft/gitfox/app/auth/authz"
	repoevents "github.com/easysoft/gitfox/app/events/repo"
	"github.com/easysoft/gitfox/app/services/codeowners"
	"github.com/easysoft/gitfox/app/services/importer"
	"github.com/easysoft/gitfox/app/services/instrument"
	"github.com/easysoft/gitfox/app/services/keywordsearch"
	"github.com/easysoft/gitfox/app/services/label"
	"github.com/easysoft/gitfox/app/services/locker"
	"github.com/easysoft/gitfox/app/services/protection"
	"github.com/easysoft/gitfox/app/services/publicaccess"
	"github.com/easysoft/gitfox/app/services/settings"
	"github.com/easysoft/gitfox/app/services/usergroup"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/audit"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/lock"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/check"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideController,
)

func ProvideController(
	config *types.Config,
	tx dbtx.Transactor,
	urlProvider url.Provider,
	authorizer authz.Authorizer,
	repoStore store.RepoStore,
	spaceStore store.SpaceStore,
	memberShipStore store.MembershipStore,
	pipelineStore store.PipelineStore,
	principalStore store.PrincipalStore,
	executionStore store.ExecutionStore,
	ruleStore store.RuleStore,
	checkStore store.CheckStore,
	pullReqStore store.PullReqStore,
	settings *settings.Service,
	principalInfoCache store.PrincipalInfoCache,
	protectionManager *protection.Manager,
	rpcClient git.Interface,
	importer *importer.Repository,
	codeOwners *codeowners.Service,
	reporeporter *repoevents.Reporter,
	indexer keywordsearch.Indexer,
	limiter limiter.ResourceLimiter,
	locker *locker.Locker,
	auditService audit.Service,
	mtxManager lock.MutexManager,
	identifierCheck check.RepoIdentifier,
	repoChecks Check,
	publicAccess publicaccess.Service,
	labelSvc *label.Service,
	instrumentation instrument.Service,
	userGroupStore store.UserGroupStore,
	userGroupService usergroup.SearchService,
) *Controller {
	return NewController(config, tx, urlProvider,
		authorizer,
		repoStore, spaceStore, memberShipStore, pipelineStore, executionStore,
		principalStore, ruleStore, checkStore, pullReqStore, settings,
		principalInfoCache, protectionManager, rpcClient, importer,
		codeOwners, reporeporter, indexer, limiter, locker, auditService, mtxManager, identifierCheck,
		repoChecks, publicAccess, labelSvc, instrumentation, userGroupStore, userGroupService)
}

func ProvideRepoCheck() Check {
	return NewNoOpRepoChecks()
}
