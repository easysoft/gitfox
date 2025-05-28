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

//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/easysoft/gitfox/app/api/controller/aiagent"
	"github.com/easysoft/gitfox/app/api/controller/capabilities"
	checkcontroller "github.com/easysoft/gitfox/app/api/controller/check"
	"github.com/easysoft/gitfox/app/api/controller/connector"
	"github.com/easysoft/gitfox/app/api/controller/execution"
	githookCtrl "github.com/easysoft/gitfox/app/api/controller/githook"
	gitspaceCtrl "github.com/easysoft/gitfox/app/api/controller/gitspace"
	infraproviderCtrl "github.com/easysoft/gitfox/app/api/controller/infraprovider"
	controllerkeywordsearch "github.com/easysoft/gitfox/app/api/controller/keywordsearch"
	"github.com/easysoft/gitfox/app/api/controller/limiter"
	controllerlogs "github.com/easysoft/gitfox/app/api/controller/logs"
	"github.com/easysoft/gitfox/app/api/controller/migrate"
	"github.com/easysoft/gitfox/app/api/controller/pipeline"
	"github.com/easysoft/gitfox/app/api/controller/plugin"
	"github.com/easysoft/gitfox/app/api/controller/principal"
	"github.com/easysoft/gitfox/app/api/controller/pullreq"
	"github.com/easysoft/gitfox/app/api/controller/repo"
	"github.com/easysoft/gitfox/app/api/controller/reposettings"
	runnerCtrl "github.com/easysoft/gitfox/app/api/controller/runner"
	"github.com/easysoft/gitfox/app/api/controller/secret"
	"github.com/easysoft/gitfox/app/api/controller/service"
	"github.com/easysoft/gitfox/app/api/controller/serviceaccount"
	"github.com/easysoft/gitfox/app/api/controller/space"
	"github.com/easysoft/gitfox/app/api/controller/system"
	"github.com/easysoft/gitfox/app/api/controller/template"
	controllertrigger "github.com/easysoft/gitfox/app/api/controller/trigger"
	"github.com/easysoft/gitfox/app/api/controller/upload"
	"github.com/easysoft/gitfox/app/api/controller/user"
	"github.com/easysoft/gitfox/app/api/controller/usergroup"
	controllerwebhook "github.com/easysoft/gitfox/app/api/controller/webhook"
	"github.com/easysoft/gitfox/app/api/openapi"
	"github.com/easysoft/gitfox/app/artifact"
	controllerartifact "github.com/easysoft/gitfox/app/artifact/controller"
	"github.com/easysoft/gitfox/app/auth/authn"
	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/bootstrap"
	connectorservice "github.com/easysoft/gitfox/app/connector"
	gitevents "github.com/easysoft/gitfox/app/events/git"
	gitspaceevents "github.com/easysoft/gitfox/app/events/gitspace"
	gitspaceinfraevents "github.com/easysoft/gitfox/app/events/gitspaceinfra"
	pipelineevents "github.com/easysoft/gitfox/app/events/pipeline"
	pullreqevents "github.com/easysoft/gitfox/app/events/pullreq"
	repoevents "github.com/easysoft/gitfox/app/events/repo"
	infrastructure "github.com/easysoft/gitfox/app/gitspace/infrastructure"
	"github.com/easysoft/gitfox/app/gitspace/logutil"
	"github.com/easysoft/gitfox/app/gitspace/orchestrator"
	containerorchestrator "github.com/easysoft/gitfox/app/gitspace/orchestrator/container"
	containerGit "github.com/easysoft/gitfox/app/gitspace/orchestrator/git"
	"github.com/easysoft/gitfox/app/gitspace/orchestrator/ide"
	containerUser "github.com/easysoft/gitfox/app/gitspace/orchestrator/user"
	"github.com/easysoft/gitfox/app/gitspace/scm"
	gitspacesecret "github.com/easysoft/gitfox/app/gitspace/secret"
	"github.com/easysoft/gitfox/app/pipeline/canceler"
	"github.com/easysoft/gitfox/app/pipeline/commit"
	"github.com/easysoft/gitfox/app/pipeline/converter"
	"github.com/easysoft/gitfox/app/pipeline/file"
	"github.com/easysoft/gitfox/app/pipeline/manager"
	"github.com/easysoft/gitfox/app/pipeline/resolver"
	"github.com/easysoft/gitfox/app/pipeline/runner"
	"github.com/easysoft/gitfox/app/pipeline/scheduler"
	pipelinestorage "github.com/easysoft/gitfox/app/pipeline/storage"
	"github.com/easysoft/gitfox/app/pipeline/triggerer"
	"github.com/easysoft/gitfox/app/router"
	"github.com/easysoft/gitfox/app/server"
	"github.com/easysoft/gitfox/app/services"
	aiagentservice "github.com/easysoft/gitfox/app/services/aiagent"
	"github.com/easysoft/gitfox/app/services/artifactgc"
	capabilitiesservice "github.com/easysoft/gitfox/app/services/capabilities"
	"github.com/easysoft/gitfox/app/services/cleanup"
	"github.com/easysoft/gitfox/app/services/codecomments"
	"github.com/easysoft/gitfox/app/services/codeowners"
	"github.com/easysoft/gitfox/app/services/exporter"
	"github.com/easysoft/gitfox/app/services/gitspaceevent"
	"github.com/easysoft/gitfox/app/services/gitspaceservice"
	"github.com/easysoft/gitfox/app/services/importer"
	"github.com/easysoft/gitfox/app/services/instrument"
	"github.com/easysoft/gitfox/app/services/keywordsearch"
	svclabel "github.com/easysoft/gitfox/app/services/label"
	locker "github.com/easysoft/gitfox/app/services/locker"
	messagingservice "github.com/easysoft/gitfox/app/services/messaging"
	"github.com/easysoft/gitfox/app/services/metric"
	migrateservice "github.com/easysoft/gitfox/app/services/migrate"
	"github.com/easysoft/gitfox/app/services/notification"
	"github.com/easysoft/gitfox/app/services/notification/mailer"
	"github.com/easysoft/gitfox/app/services/protection"
	"github.com/easysoft/gitfox/app/services/publicaccess"
	"github.com/easysoft/gitfox/app/services/publickey"
	pullreqservice "github.com/easysoft/gitfox/app/services/pullreq"
	reposervice "github.com/easysoft/gitfox/app/services/repo"
	"github.com/easysoft/gitfox/app/services/settings"
	systemsvc "github.com/easysoft/gitfox/app/services/system"
	"github.com/easysoft/gitfox/app/services/trigger"
	usergroupservice "github.com/easysoft/gitfox/app/services/usergroup"
	"github.com/easysoft/gitfox/app/services/webhook"
	"github.com/easysoft/gitfox/app/sse"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/cache"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/logs"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/audit"
	"github.com/easysoft/gitfox/blob"
	cliserver "github.com/easysoft/gitfox/cli/operations/server"
	"github.com/easysoft/gitfox/encrypt"
	"github.com/easysoft/gitfox/events"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/git/api"
	"github.com/easysoft/gitfox/git/storage"
	infraproviderpkg "github.com/easysoft/gitfox/infraprovider"
	"github.com/easysoft/gitfox/internal/runner/extend"
	"github.com/easysoft/gitfox/job"
	"github.com/easysoft/gitfox/livelog"
	"github.com/easysoft/gitfox/lock"
	"github.com/easysoft/gitfox/pubsub"
	"github.com/easysoft/gitfox/ssh"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/check"

	"github.com/google/wire"
)

func initSystem(ctx context.Context, config *types.Config) (*cliserver.System, error) {
	wire.Build(
		artifactgc.WireSet,
		cliserver.NewSystem,
		cliserver.ProvideRedis,
		bootstrap.WireSet,
		cliserver.ProvideDatabaseConfig,
		database.WireSet,
		database.WireSetOrm,
		cliserver.ProvideBlobStoreConfig,
		mailer.WireSet,
		notification.WireSet,
		blob.WireSet,
		dbtx.WireSet,
		cache.WireSet,
		router.WireSet,
		pullreqservice.WireSet,
		services.WireSet,
		services.ProvideGitspaceServices,
		server.WireSet,
		url.WireSet,
		space.WireSet,
		limiter.WireSet,
		publicaccess.WireSet,
		repo.WireSet,
		runnerCtrl.WireSet,
		reposettings.WireSet,
		pullreq.WireSet,
		controllerwebhook.WireSet,
		controllerwebhook.ProvidePreprocessor,
		svclabel.WireSet,
		serviceaccount.WireSet,
		user.WireSet,
		upload.WireSet,
		service.WireSet,
		principal.WireSet,
		usergroupservice.WireSet,
		system.WireSet,
		authn.WireSet,
		authz.WireSet,
		infrastructure.WireSet,
		infraproviderpkg.WireSet,
		gitspaceevents.WireSet,
		pipelineevents.WireSet,
		infraproviderCtrl.WireSet,
		gitspaceCtrl.WireSet,
		gitevents.WireSet,
		pullreqevents.WireSet,
		repoevents.WireSet,
		storage.WireSet,
		api.WireSet,
		cliserver.ProvideGitConfig,
		git.WireSet,
		store.WireSet,
		check.WireSet,
		encrypt.WireSet,
		cliserver.ProvideEventsConfig,
		events.WireSet,
		cliserver.ProvideWebhookConfig,
		cliserver.ProvideNotificationConfig,
		webhook.WireSet,
		cliserver.ProvideTriggerConfig,
		trigger.WireSet,
		githookCtrl.ExtenderWireSet,
		githookCtrl.WireSet,
		cliserver.ProvideLockConfig,
		lock.WireSet,
		locker.WireSet,
		cliserver.ProvidePubsubConfig,
		pubsub.WireSet,
		cliserver.ProvideJobsConfig,
		job.WireSet,
		cliserver.ProvideCleanupConfig,
		cleanup.WireSet,
		codecomments.WireSet,
		protection.WireSet,
		checkcontroller.WireSet,
		execution.WireSet,
		pipeline.WireSet,
		logs.WireSet,
		livelog.WireSet,
		controllerlogs.WireSet,
		secret.WireSet,
		connector.WireSet,
		connectorservice.WireSet,
		template.WireSet,
		manager.WireSet,
		triggerer.WireSet,
		file.WireSet,
		converter.WireSet,
		runner.WireSet,
		sse.WireSet,
		scheduler.WireSet,
		commit.WireSet,
		controllertrigger.WireSet,
		plugin.WireSet,
		resolver.WireSet,
		importer.WireSet,
		migrateservice.WireSet,
		canceler.WireSet,
		exporter.WireSet,
		metric.WireSet,
		reposervice.WireSet,
		cliserver.ProvideCodeOwnerConfig,
		codeowners.WireSet,
		gitspaceevent.WireSet,
		cliserver.ProvideKeywordSearchConfig,
		keywordsearch.WireSet,
		controllerkeywordsearch.WireSet,
		controllerartifact.WireSet,
		settings.WireSet,
		systemsvc.WireSet,
		usergroup.WireSet,
		artifact.WireSet,
		cliserver.ProvideArtifactConfig,
		cliserver.ProvideStorageConfig,
		openapi.WireSet,
		repo.ProvideRepoCheck,
		audit.WireSet,
		pipelinestorage.WireSet,
		extend.WireSet,
		ssh.WireSet,
		publickey.WireSet,
		migrate.WireSet,
		scm.WireSet,
		gitspacesecret.WireSet,
		orchestrator.WireSet,
		containerorchestrator.WireSet,
		cliserver.ProvideIDEVSCodeWebConfig,
		cliserver.ProvideDockerConfig,
		cliserver.ProvideGitspaceEventConfig,
		logutil.WireSet,
		cliserver.ProvideGitspaceOrchestratorConfig,
		ide.WireSet,
		gitspaceinfraevents.WireSet,
		gitspaceservice.WireSet,
		cliserver.ProvideGitspaceInfraProvisionerConfig,
		cliserver.ProvideIDEVSCodeConfig,
		instrument.WireSet,
		aiagentservice.WireSet,
		aiagent.WireSet,
		capabilities.WireSet,
		capabilitiesservice.WireSet,
		// secretservice.WireSet,
		containerGit.WireSet,
		containerUser.WireSet,
		messagingservice.WireSet,
	)
	return &cliserver.System{}, nil
}
