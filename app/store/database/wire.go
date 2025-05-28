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

package database

import (
	"context"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	aiorm "github.com/easysoft/gitfox/app/store/database/ai"
	"github.com/easysoft/gitfox/app/store/database/artifacts"
	connectorsorm "github.com/easysoft/gitfox/app/store/database/connectors"
	"github.com/easysoft/gitfox/app/store/database/gitspace"
	infraproviderorm "github.com/easysoft/gitfox/app/store/database/infraprovider"
	labelsorm "github.com/easysoft/gitfox/app/store/database/labels"
	"github.com/easysoft/gitfox/app/store/database/migrate"
	"github.com/easysoft/gitfox/app/store/database/pipeline"
	principalorm "github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/app/store/database/publicaccess"
	publickeyorm "github.com/easysoft/gitfox/app/store/database/publickey"
	"github.com/easysoft/gitfox/app/store/database/pullreq"
	"github.com/easysoft/gitfox/app/store/database/repo"
	spaceorm "github.com/easysoft/gitfox/app/store/database/space"
	"github.com/easysoft/gitfox/app/store/database/system"
	usergrouporm "github.com/easysoft/gitfox/app/store/database/usergroup"
	hookorm "github.com/easysoft/gitfox/app/store/database/webhooks"
	"github.com/easysoft/gitfox/job"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"

	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideDatabase,
	ProvidePrincipalStore,
	ProvideUserGroupStore,
	ProvideUserGroupReviewerStore,
	ProvidePrincipalInfoView,
	ProvideInfraProviderResourceView,
	ProvideSpacePathStore,
	ProvideSpaceStore,
	ProvideRepoStore,
	ProvideRuleStore,
	ProvideJobStore,
	ProvideExecutionStore,
	ProvidePipelineStore,
	ProvideStageStore,
	ProvideStepStore,
	ProvideAIStore,
	ProvideSecretStore,
	ProvideRepoGitInfoView,
	ProvideMembershipStore,
	ProvideTokenStore,
	ProvidePullReqStore,
	ProvidePullReqActivityStore,
	ProvideCodeCommentView,
	ProvidePullReqReviewStore,
	ProvidePullReqReviewerStore,
	ProvidePullReqFileViewStore,
	ProvideWebhookStore,
	ProvideWebhookExecutionStore,
	ProvideSettingsStore,
	ProvidePublicAccessStore,
	ProvideCheckStore,
	ProvideConnectorStore,
	ProvideTemplateStore,
	ProvideTriggerStore,
	ProvidePluginStore,
	ProvidePublicKeyStore,
	ProvideInfraProviderConfigStore,
	ProvideInfraProviderResourceStore,
	ProvideGitspaceConfigStore,
	ProvideGitspaceInstanceStore,
	ProvideGitspaceEventStore,
	ProvideLabelStore,
	ProvideLabelValueStore,
	ProvidePullReqLabelStore,
	ProvideInfraProviderTemplateStore,
	ProvideInfraProvisionedStore,
)

// WireSetOrm provides a wire orm set for this package.
var WireSetOrm = wire.NewSet(
	ProvideGormDatabase,
	ProvideArtifactStore,
)

// migrator is helper function to set up the database by performing automated
// database migration steps.
func migrator(ctx context.Context, db *sqlx.DB) error {
	return migrate.Migrate(ctx, db)
}

// ProvideDatabase provides a database connection.
func ProvideDatabase(ctx context.Context, config database.Config) (*sqlx.DB, error) {
	dsn, err := buildDatasource(ctx, config)
	if err != nil {
		return nil, err
	}
	return database.ConnectAndMigrate(
		ctx,
		config.Driver,
		dsn,
		migrator,
	)
}

func buildDatasource(ctx context.Context, config database.Config) (string, error) {
	var datasource = config.Datasource
	var err error

	switch config.Driver {
	case "sqlite3":
		datasource = config.Datasource
	default:
		if config.Datasource == "gitfox.db" {
			datasource, err = buildDSN(config)
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("build datasource failed")
				return "", err
			}
		}
	}

	return datasource, nil
}

// ProvideGormDatabase provide a gorm database connection
func ProvideGormDatabase(ctx context.Context, dbConfig database.Config, config *types.Config, depend *sqlx.DB) (*gorm.DB, error) {
	opts := make([]database.GormConfigOption, 0)
	if config.Database.Trace {
		opt := database.GormConfigLogger{Level: logger.Info}
		opts = append(opts, opt)
	}

	dsn, err := buildDatasource(ctx, dbConfig)
	if err != nil {
		return nil, err
	}
	if dbConfig.Password != "" && strings.Contains(dsn, dbConfig.Password) {
		log.Ctx(ctx).Debug().Msgf("datasource is %s", strings.ReplaceAll(dsn, dbConfig.Password, "******"))
	}
	return database.ConnectGorm(ctx, dbConfig.Driver, dsn, opts...)
}

// ProvidePrincipalStore provides a principal store.
func ProvidePrincipalStore(db *gorm.DB, uidTransformation store.PrincipalUIDTransformation) store.PrincipalStore {
	return principalorm.NewPrincipalOrmStore(db, uidTransformation)
}

// ProvideUserGroupStore provides a principal store.
func ProvideUserGroupStore(db *gorm.DB) store.UserGroupStore {
	return usergrouporm.NewUserGroupStore(db)
}

// ProvideUserGroupReviewerStore provides a usergroup reviewer store.
func ProvideUserGroupReviewerStore(
	db *gorm.DB,
	pInfoCache store.PrincipalInfoCache,
	userGroupStore store.UserGroupStore,
) store.UserGroupReviewersStore {
	return usergrouporm.NewUsergroupReviewerStore(db, pInfoCache, userGroupStore)
}

// ProvidePrincipalInfoView provides a principal info store.
func ProvidePrincipalInfoView(db *gorm.DB) store.PrincipalInfoView {
	return principalorm.NewPrincipalOrmInfoView(db)
}

// ProvideInfraProviderResourceView provides a principal info store.
func ProvideInfraProviderResourceView(db *gorm.DB) store.InfraProviderResourceView {
	return infraproviderorm.NewInfraProviderResourceView(db)
}

// ProvideSpacePathStore provides a space path store.
func ProvideSpacePathStore(
	db *gorm.DB,
	spacePathTransformation store.SpacePathTransformation,
) store.SpacePathStore {
	return spaceorm.NewSpacePathOrmStore(db, spacePathTransformation)
}

// ProvideSpaceStore provides a space store.
func ProvideSpaceStore(
	db *gorm.DB,
	spacePathCache store.SpacePathCache,
	spacePathStore store.SpacePathStore,
) store.SpaceStore {
	return spaceorm.NewSpaceOrmStore(db, spacePathCache, spacePathStore)
}

// ProvideRepoStore provides a repo store.
func ProvideRepoStore(
	db *gorm.DB,
	spacePathCache store.SpacePathCache,
	spacePathStore store.SpacePathStore,
	spaceStore store.SpaceStore,
) store.RepoStore {
	return repo.NewRepoOrmStore(db, spacePathCache, spacePathStore, spaceStore)
}

// ProvideRuleStore provides a rule store.
func ProvideRuleStore(
	db *gorm.DB,
	principalInfoCache store.PrincipalInfoCache,
) store.RuleStore {
	return repo.NewRuleOrmStore(db, principalInfoCache)
}

// ProvideJobStore provides a job store.
func ProvideJobStore(db *gorm.DB) job.Store {
	return system.NewJobOrmStore(db)
}

// ProvidePipelineStore provides a pipeline store.
func ProvidePipelineStore(db *gorm.DB) store.PipelineStore {
	return pipeline.NewPipelineOrmStore(db)
}

// ProvideInfraProviderConfigStore provides a infraprovider config store.
func ProvideInfraProviderConfigStore(db *gorm.DB) store.InfraProviderConfigStore {
	return infraproviderorm.NewInfraProviderConfigStore(db)
}

// ProvideGitspaceInstanceStore provides a infraprovider resource store.
func ProvideInfraProviderResourceStore(db *gorm.DB) store.InfraProviderResourceStore {
	return infraproviderorm.NewInfraProviderResourceStore(db)
}

// ProvideGitspaceConfigStore provides a gitspace config store.
func ProvideGitspaceConfigStore(
	db *gorm.DB,
	pCache store.PrincipalInfoCache,
	rCache store.InfraProviderResourceCache,
) store.GitspaceConfigStore {
	return gitspace.NewGitspaceConfigStore(db, pCache, rCache)
}

// ProvideGitspaceInstanceStore provides a gitspace instance store.
func ProvideGitspaceInstanceStore(db *gorm.DB) store.GitspaceInstanceStore {
	return gitspace.NewGitspaceInstanceStore(db)
}

// ProvideStageStore provides a stage store.
func ProvideStageStore(db *gorm.DB) store.StageStore {
	return pipeline.NewStageOrmStore(db)
}

// ProvideStepStore provides a step store.
func ProvideStepStore(db *gorm.DB) store.StepStore {
	return pipeline.NewStepOrmStore(db)
}

// ProvideSecretStore provides a secret store.
func ProvideSecretStore(db *gorm.DB) store.SecretStore {
	return spaceorm.NewSecretOrmStore(db)
}

// ProvideConnectorStore provides a connector store.
func ProvideConnectorStore(db *gorm.DB, secretStore store.SecretStore) store.ConnectorStore {
	return connectorsorm.NewConnectorStore(db, secretStore)
}

// ProvideTemplateStore provides a template store.
func ProvideTemplateStore(db *gorm.DB) store.TemplateStore {
	return spaceorm.NewTemplateOrmStore(db)
}

// ProvideTriggerStore provides a trigger store.
func ProvideTriggerStore(db *gorm.DB) store.TriggerStore {
	return pipeline.NewTriggerOrmStore(db)
}

// ProvideExecutionStore provides an execution store.
func ProvideExecutionStore(db *gorm.DB) store.ExecutionStore {
	return pipeline.NewExecutionOrmStore(db)
}

// ProvidePluginStore provides a plugin store.
func ProvidePluginStore(db *gorm.DB) store.PluginStore {
	return system.NewPluginOrmStore(db)
}

// ProvideRepoGitInfoView provides a repo git UID view.
func ProvideRepoGitInfoView(db *gorm.DB) store.RepoGitInfoView {
	return repo.NewRepoGitOrmInfoView(db)
}

func ProvideMembershipStore(
	db *gorm.DB,
	principalInfoCache store.PrincipalInfoCache,
	spacePathStore store.SpacePathStore,
	spaceStore store.SpaceStore,
) store.MembershipStore {
	return spaceorm.NewMembershipOrmStore(db, principalInfoCache, spacePathStore, spaceStore)
}

// ProvideTokenStore provides a token store.
func ProvideTokenStore(db *gorm.DB) store.TokenStore {
	return principalorm.NewTokenOrmStore(db)
}

// ProvidePullReqStore provides a pull request store.
func ProvidePullReqStore(
	db *gorm.DB,
	principalInfoCache store.PrincipalInfoCache,
) store.PullReqStore {
	return pullreq.NewPullReqOrmStore(db, principalInfoCache)
}

// ProvidePullReqActivityStore provides a pull request activity store.
func ProvidePullReqActivityStore(
	db *gorm.DB,
	principalInfoCache store.PrincipalInfoCache,
) store.PullReqActivityStore {
	return pullreq.NewPullReqActivityOrmStore(db, principalInfoCache)
}

// ProvideCodeCommentView provides a code comment view.
func ProvideCodeCommentView(db *gorm.DB) store.CodeCommentView {
	return pullreq.NewCodeCommentOrmView(db)
}

// ProvidePullReqReviewStore provides a pull request review store.
func ProvidePullReqReviewStore(db *gorm.DB) store.PullReqReviewStore {
	return pullreq.NewPullReqReviewOrmStore(db)
}

// ProvidePullReqReviewerStore provides a pull request reviewer store.
func ProvidePullReqReviewerStore(
	db *gorm.DB,
	principalInfoCache store.PrincipalInfoCache,
) store.PullReqReviewerStore {
	return pullreq.NewPullReqReviewerOrmStore(db, principalInfoCache)
}

// ProvidePullReqFileViewStore provides a pull request file view store.
func ProvidePullReqFileViewStore(db *gorm.DB) store.PullReqFileViewStore {
	return pullreq.NewPullReqFileViewOrmStore(db)
}

// ProvideWebhookStore provides a webhook store.
func ProvideWebhookStore(db *gorm.DB) store.WebhookStore {
	return hookorm.NewWebhookOrmStore(db)
}

// ProvideWebhookExecutionStore provides a webhook execution store.
func ProvideWebhookExecutionStore(db *gorm.DB) store.WebhookExecutionStore {
	return hookorm.NewWebhookExecutionOrmStore(db)
}

// ProvideCheckStore provides a status check result store.
func ProvideCheckStore(
	db *gorm.DB,
	principalInfoCache store.PrincipalInfoCache,
) store.CheckStore {
	return repo.NewCheckStoreOrm(db, principalInfoCache)
}

// ProvideArtifactStore provides a artifact store
func ProvideArtifactStore(orm *gorm.DB) store.ArtifactStore {
	return artifacts.NewStore(orm)
}

// ProvideSettingsStore provides a settings store.
func ProvideSettingsStore(db *gorm.DB) store.SettingsStore {
	return system.NewSettingsOrmStore(db)
}

// ProvidePublicAccessStore provides a public access store.
func ProvidePublicAccessStore(db *gorm.DB) store.PublicAccessStore {
	return publicaccess.NewPublicAccessStore(db)
}

// ProvidePublicKeyStore provides a public key store.
func ProvidePublicKeyStore(db *gorm.DB) store.PublicKeyStore {
	return publickeyorm.NewPublicKeyStore(db)
}

// ProvideGitspaceEventStore provides a gitspace event store.
func ProvideGitspaceEventStore(db *gorm.DB) store.GitspaceEventStore {
	return gitspace.NewGitspaceEventStore(db)
}

// ProvideLabelStore provides a label store.
func ProvideLabelStore(db *gorm.DB) store.LabelStore {
	return labelsorm.NewLabelStore(db)
}

// ProvideLabelValueStore provides a label value store.
func ProvideLabelValueStore(db *gorm.DB) store.LabelValueStore {
	return labelsorm.NewLabelValueStore(db)
}

// ProvideLabelValueStore provides a label value store.
func ProvidePullReqLabelStore(db *gorm.DB) store.PullReqLabelAssignmentStore {
	return labelsorm.NewPullReqLabelStore(db)
}

// ProvideInfraProviderTemplateStore provides a infraprovider template store.
func ProvideInfraProviderTemplateStore(db *gorm.DB) store.InfraProviderTemplateStore {
	return infraproviderorm.NewInfraProviderTemplateStore(db)
}

// ProvideInfraProvisionedStore provides a provisioned infra store.
func ProvideInfraProvisionedStore(db *gorm.DB) store.InfraProvisionedStore {
	return infraproviderorm.NewInfraProvisionedStore(db)
}

// ProvideAIStore provides a ai store.
func ProvideAIStore(db *gorm.DB) store.AIStore {
	return aiorm.NewAIStore(db)
}
