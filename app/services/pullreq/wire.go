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
	"context"

	"github.com/easysoft/gitfox/app/auth/authz"
	gitevents "github.com/easysoft/gitfox/app/events/git"
	pullreqevents "github.com/easysoft/gitfox/app/events/pullreq"
	"github.com/easysoft/gitfox/app/services/codecomments"
	"github.com/easysoft/gitfox/app/services/label"
	"github.com/easysoft/gitfox/app/services/protection"
	"github.com/easysoft/gitfox/app/sse"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/events"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/pubsub"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideService,
	ProvideListService,
)

func ProvideService(ctx context.Context,
	config *types.Config,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullReqEvFactory *events.ReaderFactory[*pullreqevents.Reader],
	pullReqEvReporter *pullreqevents.Reporter,
	git git.Interface,
	repoGitInfoCache store.RepoGitInfoCache,
	repoStore store.RepoStore,
	pullreqStore store.PullReqStore,
	activityStore store.PullReqActivityStore,
	principalInfoCache store.PrincipalInfoCache,
	codeCommentView store.CodeCommentView,
	codeCommentMigrator *codecomments.Migrator,
	fileViewStore store.PullReqFileViewStore,
	pubsub pubsub.PubSub,
	urlProvider url.Provider,
	sseStreamer sse.Streamer,
) (*Service, error) {
	return New(ctx,
		config,
		gitReaderFactory,
		pullReqEvFactory,
		pullReqEvReporter,
		git,
		repoGitInfoCache,
		repoStore,
		pullreqStore,
		activityStore,
		codeCommentView,
		codeCommentMigrator,
		fileViewStore,
		principalInfoCache,
		pubsub,
		urlProvider,
		sseStreamer,
	)
}

func ProvideListService(
	tx dbtx.Transactor,
	git git.Interface,
	authorizer authz.Authorizer,
	spaceStore store.SpaceStore,
	repoStore store.RepoStore,
	repoGitInfoCache store.RepoGitInfoCache,
	pullreqStore store.PullReqStore,
	checkStore store.CheckStore,
	labelSvc *label.Service,
	protectionManager *protection.Manager,
) *ListService {
	return NewListService(
		tx,
		git,
		authorizer,
		spaceStore,
		repoStore,
		repoGitInfoCache,
		pullreqStore,
		checkStore,
		labelSvc,
		protectionManager,
	)
}
