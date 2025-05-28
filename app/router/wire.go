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

package router

import (
	"context"
	"strings"

	"github.com/easysoft/gitfox/app/api/controller/aiagent"
	"github.com/easysoft/gitfox/app/api/controller/capabilities"
	"github.com/easysoft/gitfox/app/api/controller/check"
	"github.com/easysoft/gitfox/app/api/controller/connector"
	"github.com/easysoft/gitfox/app/api/controller/execution"
	"github.com/easysoft/gitfox/app/api/controller/githook"
	"github.com/easysoft/gitfox/app/api/controller/gitspace"
	"github.com/easysoft/gitfox/app/api/controller/infraprovider"
	"github.com/easysoft/gitfox/app/api/controller/keywordsearch"
	"github.com/easysoft/gitfox/app/api/controller/logs"
	"github.com/easysoft/gitfox/app/api/controller/migrate"
	"github.com/easysoft/gitfox/app/api/controller/pipeline"
	"github.com/easysoft/gitfox/app/api/controller/plugin"
	"github.com/easysoft/gitfox/app/api/controller/principal"
	"github.com/easysoft/gitfox/app/api/controller/pullreq"
	"github.com/easysoft/gitfox/app/api/controller/repo"
	"github.com/easysoft/gitfox/app/api/controller/reposettings"
	"github.com/easysoft/gitfox/app/api/controller/runner"
	"github.com/easysoft/gitfox/app/api/controller/secret"
	"github.com/easysoft/gitfox/app/api/controller/serviceaccount"
	"github.com/easysoft/gitfox/app/api/controller/space"
	"github.com/easysoft/gitfox/app/api/controller/system"
	"github.com/easysoft/gitfox/app/api/controller/template"
	"github.com/easysoft/gitfox/app/api/controller/trigger"
	"github.com/easysoft/gitfox/app/api/controller/upload"
	"github.com/easysoft/gitfox/app/api/controller/user"
	"github.com/easysoft/gitfox/app/api/controller/usergroup"
	"github.com/easysoft/gitfox/app/api/controller/webhook"
	"github.com/easysoft/gitfox/app/api/openapi"
	artctl "github.com/easysoft/gitfox/app/artifact/controller"
	"github.com/easysoft/gitfox/app/auth/authn"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/types"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideRouter,
)

func GetGitRoutingHost(ctx context.Context, urlProvider url.Provider) string {
	// use url provider as it has the latest data.
	gitHostname := urlProvider.GetGITHostname(ctx)
	apiHostname := urlProvider.GetAPIHostname(ctx)

	// only use host name to identify git traffic if it differs from api hostname.
	// TODO: Can we make this even more flexible - aka use the full base urls to route traffic?
	gitRoutingHost := ""
	if !strings.EqualFold(gitHostname, apiHostname) {
		gitRoutingHost = gitHostname
	}
	return gitRoutingHost
}

// ProvideRouter provides ordered list of routers.
func ProvideRouter(
	appCtx context.Context,
	config *types.Config,
	principalStore store.PrincipalStore,
	authenticator authn.Authenticator,
	repoCtrl *repo.Controller,
	repoSettingsCtrl *reposettings.Controller,
	executionCtrl *execution.Controller,
	logCtrl *logs.Controller,
	spaceCtrl *space.Controller,
	pipelineCtrl *pipeline.Controller,
	secretCtrl *secret.Controller,
	triggerCtrl *trigger.Controller,
	connectorCtrl *connector.Controller,
	templateCtrl *template.Controller,
	pluginCtrl *plugin.Controller,
	pullreqCtrl *pullreq.Controller,
	webhookCtrl *webhook.Controller,
	githookCtrl *githook.Controller,
	git git.Interface,
	saCtrl *serviceaccount.Controller,
	userCtrl *user.Controller,
	principalCtrl principal.Controller,
	userGroupCtrl *usergroup.Controller,
	checkCtrl *check.Controller,
	sysCtrl *system.Controller,
	blobCtrl *upload.Controller,
	searchCtrl *keywordsearch.Controller,
	artifactCtrl *artctl.Controller,
	runnerCtrl *runner.Controller,
	infraProviderCtrl *infraprovider.Controller,
	gitspaceCtrl *gitspace.Controller,
	migrateCtrl *migrate.Controller,
	aiagentCtrl *aiagent.Controller,
	capabilitiesCtrl *capabilities.Controller,
	urlProvider url.Provider,
	openapi openapi.Service,
	artStore store.ArtifactStore,
	repoStore store.RepoStore,
	fileStore storage.ContentStorage,
) *Router {
	routers := make([]Interface, 4)

	gitRoutingHost := GetGitRoutingHost(appCtx, urlProvider)
	gitHandler := NewGitHandler(
		config,
		urlProvider,
		authenticator,
		repoCtrl,
	)
	routers[0] = NewGitRouter(gitHandler, gitRoutingHost)

	apiHandler := NewAPIHandler(
		appCtx, config, principalStore,
		authenticator, repoCtrl, repoSettingsCtrl, executionCtrl, logCtrl, spaceCtrl, pipelineCtrl,
		secretCtrl, triggerCtrl, connectorCtrl, templateCtrl, pluginCtrl, pullreqCtrl, webhookCtrl,
		githookCtrl, git, saCtrl, userCtrl, principalCtrl, userGroupCtrl, checkCtrl, sysCtrl, blobCtrl, searchCtrl,
		artifactCtrl, runnerCtrl,
		infraProviderCtrl, migrateCtrl, gitspaceCtrl, aiagentCtrl, capabilitiesCtrl)
	routers[1] = NewAPIRouter(apiHandler)

	artifactHandler := NewArtifactHandler(appCtx, urlProvider, config, authenticator, artifactCtrl, artStore, repoStore, fileStore)
	routers[2] = NewArtifactRouter(artifactHandler)
	webHandler := NewWebHandler(config, authenticator, openapi)
	routers[3] = NewWebRouter(webHandler)

	return NewRouter(routers)
}
