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

package githook

import (
	"github.com/easysoft/gitfox/app/api/controller/limiter"
	"github.com/easysoft/gitfox/app/auth/authz"
	eventsgit "github.com/easysoft/gitfox/app/events/git"
	eventspr "github.com/easysoft/gitfox/app/events/pullreq"
	eventsrepo "github.com/easysoft/gitfox/app/events/repo"
	"github.com/easysoft/gitfox/app/services/codeowners"
	"github.com/easysoft/gitfox/app/services/protection"
	"github.com/easysoft/gitfox/app/services/settings"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/git/hook"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideController,
	ProvideFactory,
)

func ProvideFactory() hook.ClientFactory {
	return &ControllerClientFactory{
		githookCtrl: nil,
	}
}

func ProvideController(
	authorizer authz.Authorizer,
	principalStore store.PrincipalStore,
	repoStore store.RepoStore,
	gitReporter *eventsgit.Reporter,
	repoReporter *eventsrepo.Reporter,
	prReporter *eventspr.Reporter,
	git git.Interface,
	pullreqStore store.PullReqStore,
	urlProvider url.Provider,
	protectionManager *protection.Manager,
	githookFactory hook.ClientFactory,
	limiter limiter.ResourceLimiter,
	settings *settings.Service,
	preReceiveExtender PreReceiveExtender,
	updateExtender UpdateExtender,
	postReceiveExtender PostReceiveExtender,
	codeOwners *codeowners.Service,
	reviewerStore store.PullReqReviewerStore,
) *Controller {
	ctrl := NewController(
		authorizer,
		principalStore,
		repoStore,
		gitReporter,
		repoReporter,
		prReporter,
		git,
		pullreqStore,
		urlProvider,
		protectionManager,
		limiter,
		settings,
		preReceiveExtender,
		updateExtender,
		postReceiveExtender,
		codeOwners,
		reviewerStore,
	)

	// TODO: improve wiring if possible
	if fct, ok := githookFactory.(*ControllerClientFactory); ok {
		fct.githookCtrl = ctrl
		fct.git = git
	}

	return ctrl
}

var ExtenderWireSet = wire.NewSet(
	ProvidePreReceiveExtender,
	ProvideUpdateExtender,
	ProvidePostReceiveExtender,
)

func ProvidePreReceiveExtender() (PreReceiveExtender, error) {
	return NewPreReceiveExtender(), nil
}

func ProvideUpdateExtender() (UpdateExtender, error) {
	return NewUpdateExtender(), nil
}

func ProvidePostReceiveExtender() (PostReceiveExtender, error) {
	return NewPostReceiveExtender(), nil
}
