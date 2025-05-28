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

package space

import (
	"encoding/json"

	"github.com/easysoft/gitfox/app/api/controller/limiter"
	"github.com/easysoft/gitfox/app/api/controller/repo"
	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/services/exporter"
	"github.com/easysoft/gitfox/app/services/gitspace"
	"github.com/easysoft/gitfox/app/services/importer"
	"github.com/easysoft/gitfox/app/services/instrument"
	"github.com/easysoft/gitfox/app/services/label"
	"github.com/easysoft/gitfox/app/services/publicaccess"
	"github.com/easysoft/gitfox/app/services/pullreq"
	"github.com/easysoft/gitfox/app/sse"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/audit"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/check"
)

var (
	// TODO (Nested Spaces): Remove once full support is added
	errNestedSpacesNotSupported    = usererror.BadRequestf("Nested spaces are not supported.")
	errPublicSpaceCreationDisabled = usererror.BadRequestf("Public space creation is disabled.")
)

//nolint:revive
type SpaceOutput struct {
	types.Space
	IsPublic bool `json:"is_public" yaml:"is_public"`
}

// TODO [CODE-1363]: remove after identifier migration.
func (s SpaceOutput) MarshalJSON() ([]byte, error) {
	// alias allows us to embed the original object while avoiding an infinite loop of marshaling.
	type alias SpaceOutput
	return json.Marshal(&struct {
		alias
		UID string `json:"uid"`
	}{
		alias: (alias)(s),
		UID:   s.Identifier,
	})
}

type Controller struct {
	nestedSpacesEnabled bool

	tx              dbtx.Transactor
	urlProvider     url.Provider
	sseStreamer     sse.Streamer
	identifierCheck check.SpaceIdentifier
	authorizer      authz.Authorizer
	spacePathStore  store.SpacePathStore
	pipelineStore   store.PipelineStore
	secretStore     store.SecretStore
	connectorStore  store.ConnectorStore
	templateStore   store.TemplateStore
	spaceStore      store.SpaceStore
	repoStore       store.RepoStore
	principalStore  store.PrincipalStore
	artStore        store.ArtifactStore
	repoCtrl        *repo.Controller
	membershipStore store.MembershipStore
	prListService   *pullreq.ListService
	importer        *importer.Repository
	exporter        *exporter.Repository
	resourceLimiter limiter.ResourceLimiter
	publicAccess    publicaccess.Service
	auditService    audit.Service
	gitspaceSvc     *gitspace.Service
	labelSvc        *label.Service
	instrumentation instrument.Service
	aiStore         store.AIStore
	executionStore  store.ExecutionStore
}

func NewController(config *types.Config, tx dbtx.Transactor, urlProvider url.Provider,
	sseStreamer sse.Streamer, identifierCheck check.SpaceIdentifier, authorizer authz.Authorizer,
	spacePathStore store.SpacePathStore, pipelineStore store.PipelineStore, secretStore store.SecretStore,
	connectorStore store.ConnectorStore, templateStore store.TemplateStore, spaceStore store.SpaceStore,
	repoStore store.RepoStore, principalStore store.PrincipalStore, artStore store.ArtifactStore, repoCtrl *repo.Controller,
	membershipStore store.MembershipStore, prListService *pullreq.ListService,
	importer *importer.Repository, exporter *exporter.Repository,
	limiter limiter.ResourceLimiter, publicAccess publicaccess.Service, auditService audit.Service,
	gitspaceSvc *gitspace.Service, labelSvc *label.Service,
	instrumentation instrument.Service, aiStore store.AIStore, executionStore store.ExecutionStore,
) *Controller {
	return &Controller{
		nestedSpacesEnabled: config.NestedSpacesEnabled,
		tx:                  tx,
		urlProvider:         urlProvider,
		sseStreamer:         sseStreamer,
		identifierCheck:     identifierCheck,
		authorizer:          authorizer,
		spacePathStore:      spacePathStore,
		pipelineStore:       pipelineStore,
		secretStore:         secretStore,
		connectorStore:      connectorStore,
		templateStore:       templateStore,
		spaceStore:          spaceStore,
		repoStore:           repoStore,
		principalStore:      principalStore,
		artStore:            artStore,
		repoCtrl:            repoCtrl,
		membershipStore:     membershipStore,
		prListService:       prListService,
		importer:            importer,
		exporter:            exporter,
		resourceLimiter:     limiter,
		publicAccess:        publicAccess,
		auditService:        auditService,
		gitspaceSvc:         gitspaceSvc,
		labelSvc:            labelSvc,
		instrumentation:     instrumentation,
		aiStore:             aiStore,
		executionStore:      executionStore,
	}
}
