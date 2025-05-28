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

package pullreq

import (
	"github.com/easysoft/gitfox/app/auth/authz"
	pullreqevents "github.com/easysoft/gitfox/app/events/pullreq"
	"github.com/easysoft/gitfox/app/services/codecomments"
	"github.com/easysoft/gitfox/app/services/codeowners"
	"github.com/easysoft/gitfox/app/services/instrument"
	"github.com/easysoft/gitfox/app/services/label"
	"github.com/easysoft/gitfox/app/services/locker"
	"github.com/easysoft/gitfox/app/services/migrate"
	"github.com/easysoft/gitfox/app/services/protection"
	"github.com/easysoft/gitfox/app/services/pullreq"
	"github.com/easysoft/gitfox/app/services/usergroup"
	"github.com/easysoft/gitfox/app/sse"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/audit"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/store/database/dbtx"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideController,
)

func ProvideController(
	tx dbtx.Transactor,
	urlProvider url.Provider,
	authorizer authz.Authorizer,
	auditService audit.Service,
	pullReqStore store.PullReqStore, pullReqActivityStore store.PullReqActivityStore,
	codeCommentsView store.CodeCommentView,
	pullReqReviewStore store.PullReqReviewStore, pullReqReviewerStore store.PullReqReviewerStore,
	repoStore store.RepoStore,
	principalStore store.PrincipalStore,
	userGroupStore store.UserGroupStore,
	userGroupReviewerStore store.UserGroupReviewersStore,
	principalInfoCache store.PrincipalInfoCache,
	fileViewStore store.PullReqFileViewStore,
	membershipStore store.MembershipStore,
	checkStore store.CheckStore,
	aiStore store.AIStore,
	rpcClient git.Interface, eventReporter *pullreqevents.Reporter, codeCommentMigrator *codecomments.Migrator,
	pullreqService *pullreq.Service, pullreqListService *pullreq.ListService,
	ruleManager *protection.Manager, sseStreamer sse.Streamer,
	codeOwners *codeowners.Service, locker *locker.Locker, importer *migrate.PullReq,
	labelSvc *label.Service,
	instrumentation instrument.Service,
	userGroupService usergroup.SearchService,
) *Controller {
	return NewController(tx,
		urlProvider,
		authorizer,
		auditService,
		pullReqStore,
		pullReqActivityStore,
		codeCommentsView,
		pullReqReviewStore,
		pullReqReviewerStore,
		repoStore,
		principalStore,
		userGroupStore,
		userGroupReviewerStore,
		principalInfoCache,
		fileViewStore,
		membershipStore,
		checkStore,
		aiStore,
		rpcClient,
		eventReporter,
		codeCommentMigrator,
		pullreqService,
		pullreqListService,
		ruleManager,
		sseStreamer,
		codeOwners,
		locker,
		importer,
		labelSvc,
		instrumentation,
		userGroupService,
	)
}
