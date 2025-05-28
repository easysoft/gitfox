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
	"github.com/easysoft/gitfox/app/api/controller/limiter"
	"github.com/easysoft/gitfox/app/api/controller/repo"
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

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideController,
)

func ProvideController(config *types.Config, tx dbtx.Transactor, urlProvider url.Provider, sseStreamer sse.Streamer,
	identifierCheck check.SpaceIdentifier, authorizer authz.Authorizer, spacePathStore store.SpacePathStore,
	pipelineStore store.PipelineStore, secretStore store.SecretStore,
	connectorStore store.ConnectorStore, templateStore store.TemplateStore,
	spaceStore store.SpaceStore, repoStore store.RepoStore, principalStore store.PrincipalStore, artStore store.ArtifactStore,
	repoCtrl *repo.Controller, membershipStore store.MembershipStore, prListService *pullreq.ListService,
	importer *importer.Repository, exporter *exporter.Repository, limiter limiter.ResourceLimiter,
	publicAccess publicaccess.Service, auditService audit.Service, gitspaceService *gitspace.Service,
	labelSvc *label.Service, instrumentation instrument.Service, aiStore store.AIStore, executionStore store.ExecutionStore,
) *Controller {
	return NewController(config, tx, urlProvider,
		sseStreamer, identifierCheck, authorizer,
		spacePathStore, pipelineStore, secretStore,
		connectorStore, templateStore,
		spaceStore, repoStore, principalStore, artStore,
		repoCtrl, membershipStore, prListService,
		importer, exporter, limiter,
		publicAccess, auditService, gitspaceService,
		labelSvc, instrumentation, aiStore, executionStore,
	)
}
