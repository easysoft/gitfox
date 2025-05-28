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
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/api/controller/limiter"
	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/auth"
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
	"github.com/easysoft/gitfox/types/enum"
)

var errPublicRepoCreationDisabled = usererror.BadRequestf("Public repository creation is disabled.")

type RepositoryOutput struct {
	types.Repository
	IsPublic  bool `json:"is_public" yaml:"is_public"`
	Importing bool `json:"importing" yaml:"-"`
}

// TODO [CODE-1363]: remove after identifier migration.
func (r RepositoryOutput) MarshalJSON() ([]byte, error) {
	// alias allows us to embed the original object while avoiding an infinite loop of marshaling.
	type alias RepositoryOutput
	return json.Marshal(&struct {
		alias
		UID string `json:"uid"`
	}{
		alias: (alias)(r),
		UID:   r.Identifier,
	})
}

type Controller struct {
	defaultBranch string

	tx                 dbtx.Transactor
	urlProvider        url.Provider
	authorizer         authz.Authorizer
	repoStore          store.RepoStore
	spaceStore         store.SpaceStore
	memberShipStore    store.MembershipStore
	pipelineStore      store.PipelineStore
	executionStore     store.ExecutionStore
	principalStore     store.PrincipalStore
	ruleStore          store.RuleStore
	checkStore         store.CheckStore
	pullReqStore       store.PullReqStore
	settings           *settings.Service
	principalInfoCache store.PrincipalInfoCache
	userGroupStore     store.UserGroupStore
	userGroupService   usergroup.SearchService
	protectionManager  *protection.Manager
	git                git.Interface
	importer           *importer.Repository
	codeOwners         *codeowners.Service
	eventReporter      *repoevents.Reporter
	indexer            keywordsearch.Indexer
	resourceLimiter    limiter.ResourceLimiter
	locker             *locker.Locker
	auditService       audit.Service
	mtxManager         lock.MutexManager
	identifierCheck    check.RepoIdentifier
	repoCheck          Check
	publicAccess       publicaccess.Service
	labelSvc           *label.Service
	instrumentation    instrument.Service
}

func NewController(
	config *types.Config,
	tx dbtx.Transactor,
	urlProvider url.Provider,
	authorizer authz.Authorizer,
	repoStore store.RepoStore,
	spaceStore store.SpaceStore,
	memberShipStore store.MembershipStore,
	pipelineStore store.PipelineStore,
	executionStore store.ExecutionStore,
	principalStore store.PrincipalStore,
	ruleStore store.RuleStore,
	checkStore store.CheckStore,
	pullReqStore store.PullReqStore,
	settings *settings.Service,
	principalInfoCache store.PrincipalInfoCache,
	protectionManager *protection.Manager,
	git git.Interface,
	importer *importer.Repository,
	codeOwners *codeowners.Service,
	eventReporter *repoevents.Reporter,
	indexer keywordsearch.Indexer,
	limiter limiter.ResourceLimiter,
	locker *locker.Locker,
	auditService audit.Service,
	mtxManager lock.MutexManager,
	identifierCheck check.RepoIdentifier,
	repoCheck Check,
	publicAccess publicaccess.Service,
	labelSvc *label.Service,
	instrumentation instrument.Service,
	userGroupStore store.UserGroupStore,
	userGroupService usergroup.SearchService,
) *Controller {
	return &Controller{
		defaultBranch:      config.Git.DefaultBranch,
		tx:                 tx,
		urlProvider:        urlProvider,
		authorizer:         authorizer,
		repoStore:          repoStore,
		spaceStore:         spaceStore,
		memberShipStore:    memberShipStore,
		pipelineStore:      pipelineStore,
		executionStore:     executionStore,
		principalStore:     principalStore,
		ruleStore:          ruleStore,
		checkStore:         checkStore,
		pullReqStore:       pullReqStore,
		settings:           settings,
		principalInfoCache: principalInfoCache,
		protectionManager:  protectionManager,
		git:                git,
		importer:           importer,
		codeOwners:         codeOwners,
		eventReporter:      eventReporter,
		indexer:            indexer,
		resourceLimiter:    limiter,
		locker:             locker,
		auditService:       auditService,
		mtxManager:         mtxManager,
		identifierCheck:    identifierCheck,
		repoCheck:          repoCheck,
		publicAccess:       publicAccess,
		labelSvc:           labelSvc,
		instrumentation:    instrumentation,
		userGroupStore:     userGroupStore,
		userGroupService:   userGroupService,
	}
}

// getRepo fetches an active repo (not one that is currently being imported).
func (c *Controller) getRepo(
	ctx context.Context,
	repoRef string,
) (*types.Repository, error) {
	return GetRepo(
		ctx,
		c.repoStore,
		repoRef,
		ActiveRepoStates,
	)
}

// getRepoCheckAccess fetches an active repo (not one that is currently being imported)
// and checks if the current user has permission to access it.
func (c *Controller) getRepoCheckAccess(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	reqPermission enum.Permission,
) (*types.Repository, error) {
	return GetRepoCheckAccess(
		ctx,
		c.repoStore,
		c.authorizer,
		session,
		repoRef,
		reqPermission,
		ActiveRepoStates,
	)
}

// getRepoCheckAccessForGit fetches a repo
// and checks if the current user has permission to access it.
func (c *Controller) getRepoCheckAccessForGit(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	reqPermission enum.Permission,
) (*types.Repository, error) {
	return GetRepoCheckAccess(
		ctx,
		c.repoStore,
		c.authorizer,
		session,
		repoRef,
		reqPermission,
		nil, // Any state allowed - we'll block in the pre-receive hook.
	)
}

func ValidateParentRef(parentRef string) error {
	parentRefAsID, err := strconv.ParseInt(parentRef, 10, 64)
	if (err == nil && parentRefAsID <= 0) || (len(strings.TrimSpace(parentRef)) == 0) {
		return errRepositoryRequiresParent
	}

	return nil
}

func (c *Controller) fetchRules(
	ctx context.Context,
	session *auth.Session,
	repo *types.Repository,
) (protection.Protection, bool, error) {
	isRepoOwner, err := apiauth.IsRepoOwner(ctx, c.authorizer, session, repo)
	if err != nil {
		return nil, false, fmt.Errorf("failed to determine if user is repo owner: %w", err)
	}

	protectionRules, err := c.protectionManager.ForRepository(ctx, repo.ID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch protection rules for the repository: %w", err)
	}

	return protectionRules, isRepoOwner, nil
}

func (c *Controller) getRuleUserAndUserGroups(
	ctx context.Context,
	r *types.Rule,
) (map[int64]*types.PrincipalInfo, map[int64]*types.UserGroupInfo, error) {
	rule, err := c.parseRule(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse rule: %w", err)
	}

	userMap, err := c.getRuleUsers(ctx, rule)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rule users: %w", err)
	}
	userGroupMap, err := c.getRuleUserGroups(ctx, rule)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rule user groups: %w", err)
	}

	return userMap, userGroupMap, nil
}

func (c *Controller) getRuleUsers(
	ctx context.Context,
	rule protection.Protection,
) (map[int64]*types.PrincipalInfo, error) {
	userIDs, err := rule.UserIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID from rule: %w", err)
	}

	userMap, err := c.principalInfoCache.Map(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get principal infos: %w", err)
	}

	return userMap, nil
}

func (c *Controller) getRuleUserGroups(
	ctx context.Context,
	rule protection.Protection,
) (map[int64]*types.UserGroupInfo, error) {
	groupIDs, err := rule.UserGroupIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get group IDs from rule: %w", err)
	}

	userGroupInfoMap := make(map[int64]*types.UserGroupInfo)

	if len(groupIDs) == 0 {
		return userGroupInfoMap, nil
	}

	groupMap, err := c.userGroupStore.Map(ctx, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get userGroup infos: %w", err)
	}

	for k, v := range groupMap {
		userGroupInfoMap[k] = v.ToUserGroupInfo()
	}
	return userGroupInfoMap, nil
}

func (c *Controller) parseRule(r *types.Rule) (protection.Protection, error) {
	rule, err := c.protectionManager.FromJSON(r.Type, r.Definition, false)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json rule definition: %w", err)
	}

	return rule, nil
}
