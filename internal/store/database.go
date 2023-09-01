// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

// Package store defines the data storage interfaces.
package store

import (
	"context"
	"time"

	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"
)

type (
	// PrincipalStore defines the principal data storage.
	PrincipalStore interface {
		/*
		 * PRINCIPAL RELATED OPERATIONS.
		 */
		// Find finds the principal by id.
		Find(ctx context.Context, id int64) (*types.Principal, error)

		// FindByUID finds the principal by uid.
		FindByUID(ctx context.Context, uid string) (*types.Principal, error)

		// FindManyByUID returns all principals found for the provided UIDs.
		// If a UID isn't found, it's not returned in the list.
		FindManyByUID(ctx context.Context, uids []string) ([]*types.Principal, error)

		// FindByEmail finds the principal by email.
		FindByEmail(ctx context.Context, email string) (*types.Principal, error)

		/*
		 * USER RELATED OPERATIONS.
		 */

		// FindUser finds the user by id.
		FindUser(ctx context.Context, id int64) (*types.User, error)

		// List lists the principals matching the provided filter.
		List(ctx context.Context, fetchQuery *types.PrincipalFilter) ([]*types.Principal, error)

		// FindUserByUID finds the user by uid.
		FindUserByUID(ctx context.Context, uid string) (*types.User, error)

		// FindUserByEmail finds the user by email.
		FindUserByEmail(ctx context.Context, email string) (*types.User, error)

		// CreateUser saves the user details.
		CreateUser(ctx context.Context, user *types.User) error

		// UpdateUser updates an existing user.
		UpdateUser(ctx context.Context, user *types.User) error

		// DeleteUser deletes the user.
		DeleteUser(ctx context.Context, id int64) error

		// ListUsers returns a list of users.
		ListUsers(ctx context.Context, params *types.UserFilter) ([]*types.User, error)

		// CountUsers returns a count of users which match the given filter.
		CountUsers(ctx context.Context, opts *types.UserFilter) (int64, error)

		/*
		 * SERVICE ACCOUNT RELATED OPERATIONS.
		 */

		// FindServiceAccount finds the service account by id.
		FindServiceAccount(ctx context.Context, id int64) (*types.ServiceAccount, error)

		// FindServiceAccountByUID finds the service account by uid.
		FindServiceAccountByUID(ctx context.Context, uid string) (*types.ServiceAccount, error)

		// CreateServiceAccount saves the service account.
		CreateServiceAccount(ctx context.Context, sa *types.ServiceAccount) error

		// UpdateServiceAccount updates the service account details.
		UpdateServiceAccount(ctx context.Context, sa *types.ServiceAccount) error

		// DeleteServiceAccount deletes the service account.
		DeleteServiceAccount(ctx context.Context, id int64) error

		// ListServiceAccounts returns a list of service accounts for a specific parent.
		ListServiceAccounts(ctx context.Context,
			parentType enum.ParentResourceType, parentID int64) ([]*types.ServiceAccount, error)

		// CountServiceAccounts returns a count of service accounts for a specific parent.
		CountServiceAccounts(ctx context.Context,
			parentType enum.ParentResourceType, parentID int64) (int64, error)

		/*
		 * SERVICE RELATED OPERATIONS.
		 */

		// FindService finds the service by id.
		FindService(ctx context.Context, id int64) (*types.Service, error)

		// FindServiceByUID finds the service by uid.
		FindServiceByUID(ctx context.Context, uid string) (*types.Service, error)

		// CreateService saves the service.
		CreateService(ctx context.Context, sa *types.Service) error

		// UpdateService updates the service.
		UpdateService(ctx context.Context, sa *types.Service) error

		// DeleteService deletes the service.
		DeleteService(ctx context.Context, id int64) error

		// ListServices returns a list of service for a specific parent.
		ListServices(ctx context.Context) ([]*types.Service, error)

		// CountServices returns a count of service for a specific parent.
		CountServices(ctx context.Context) (int64, error)
	}

	// PrincipalInfoView defines helper utility for fetching types.PrincipalInfo objects.
	// It uses the same underlying data storage as PrincipalStore.
	PrincipalInfoView interface {
		Find(ctx context.Context, id int64) (*types.PrincipalInfo, error)
		FindMany(ctx context.Context, ids []int64) ([]*types.PrincipalInfo, error)
	}

	// PathStore defines the path data storage.
	// It is used to store routing paths for repos & spaces.
	PathStore interface {
		// Create creates a new path.
		Create(ctx context.Context, path *types.Path) error

		// Find finds the path for the given id.
		Find(ctx context.Context, id int64) (*types.Path, error)

		// FindWithLock finds the path for the given id and locks the entry.
		FindWithLock(ctx context.Context, id int64) (*types.Path, error)

		// FindValue finds the path for the given value.
		FindValue(ctx context.Context, value string) (*types.Path, error)

		// FindPrimary finds the primary path for a target.
		FindPrimary(ctx context.Context, targetType enum.PathTargetType, targetID int64) (*types.Path, error)

		// FindPrimaryWithLock finds the primary path for a target and locks the db entry.
		FindPrimaryWithLock(ctx context.Context, targetType enum.PathTargetType, targetID int64) (*types.Path, error)

		// Update updates an existing path.
		Update(ctx context.Context, path *types.Path) error

		// Delete deletes a specific path.
		Delete(ctx context.Context, id int64) error

		// Count returns the count of paths for a target.
		Count(ctx context.Context, targetType enum.PathTargetType, targetID int64,
			opts *types.PathFilter) (int64, error)

		// List lists all paths for a target.
		List(ctx context.Context, targetType enum.PathTargetType, targetID int64,
			opts *types.PathFilter) ([]*types.Path, error)

		// ListPrimaryDescendantsWithLock lists all primary paths that are descendants of the given path and locks them.
		ListPrimaryDescendantsWithLock(ctx context.Context, value string) ([]*types.Path, error)
	}

	// SpaceStore defines the space data storage.
	SpaceStore interface {
		// Find the space by id.
		Find(ctx context.Context, id int64) (*types.Space, error)

		// FindByRef finds the space using the spaceRef as either the id or the space path.
		FindByRef(ctx context.Context, spaceRef string) (*types.Space, error)

		// Create creates a new space
		Create(ctx context.Context, space *types.Space) error

		// Update updates the space details.
		Update(ctx context.Context, space *types.Space) error

		// UpdateOptLock updates the space using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, space *types.Space,
			mutateFn func(space *types.Space) error) (*types.Space, error)

		// Delete deletes the space.
		Delete(ctx context.Context, id int64) error

		// Count the child spaces of a space.
		Count(ctx context.Context, id int64, opts *types.SpaceFilter) (int64, error)

		// List returns a list of child spaces in a space.
		List(ctx context.Context, id int64, opts *types.SpaceFilter) ([]types.Space, error)
	}

	// RepoStore defines the repository data storage.
	RepoStore interface {
		// Find the repo by id.
		Find(ctx context.Context, id int64) (*types.Repository, error)

		// FindByRef finds the repo using the repoRef as either the id or the repo path.
		FindByRef(ctx context.Context, repoRef string) (*types.Repository, error)

		// Create a new repo.
		Create(ctx context.Context, repo *types.Repository) error

		// Update the repo details.
		Update(ctx context.Context, repo *types.Repository) error

		// UpdateOptLock the repo details using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, repo *types.Repository,
			mutateFn func(repository *types.Repository) error) (*types.Repository, error)

		// Delete the repo.
		Delete(ctx context.Context, id int64) error

		// Count of repos in a space.
		Count(ctx context.Context, parentID int64, opts *types.RepoFilter) (int64, error)

		// List returns a list of repos in a space.
		List(ctx context.Context, parentID int64, opts *types.RepoFilter) ([]*types.Repository, error)
	}

	// RepoGitInfoView defines the repository GitUID view.
	RepoGitInfoView interface {
		Find(ctx context.Context, id int64) (*types.RepositoryGitInfo, error)
	}

	// MembershipStore defines the membership data storage.
	MembershipStore interface {
		Find(ctx context.Context, key types.MembershipKey) (*types.Membership, error)
		FindUser(ctx context.Context, key types.MembershipKey) (*types.MembershipUser, error)
		Create(ctx context.Context, membership *types.Membership) error
		Update(ctx context.Context, membership *types.Membership) error
		Delete(ctx context.Context, key types.MembershipKey) error
		CountUsers(ctx context.Context, spaceID int64, filter types.MembershipUserFilter) (int64, error)
		ListUsers(ctx context.Context, spaceID int64, filter types.MembershipUserFilter) ([]types.MembershipUser, error)
		CountSpaces(ctx context.Context, userID int64, filter types.MembershipSpaceFilter) (int64, error)
		ListSpaces(ctx context.Context, userID int64, filter types.MembershipSpaceFilter) ([]types.MembershipSpace, error)
	}

	// TokenStore defines the token data storage.
	TokenStore interface {
		// Find finds the token by id
		Find(ctx context.Context, id int64) (*types.Token, error)

		// FindByUID finds the token by principalId and tokenUID
		FindByUID(ctx context.Context, principalID int64, tokenUID string) (*types.Token, error)

		// Create saves the token details.
		Create(ctx context.Context, token *types.Token) error

		// Delete deletes the token with the given id.
		Delete(ctx context.Context, id int64) error

		// DeleteForPrincipal deletes all tokens for a specific principal
		DeleteForPrincipal(ctx context.Context, principalID int64) error

		// List returns a list of tokens of a specific type for a specific principal.
		List(ctx context.Context, principalID int64, tokenType enum.TokenType) ([]*types.Token, error)

		// Count returns a count of tokens of a specifc type for a specific principal.
		Count(ctx context.Context, principalID int64, tokenType enum.TokenType) (int64, error)
	}

	// PullReqStore defines the pull request data storage.
	PullReqStore interface {
		// Find the pull request by id.
		Find(ctx context.Context, id int64) (*types.PullReq, error)

		// FindByNumberWithLock finds the pull request by repo ID and the pull request number
		// and acquires an exclusive lock of the pull request database row for the duration of the transaction.
		FindByNumberWithLock(ctx context.Context, repoID, number int64) (*types.PullReq, error)

		// FindByNumber finds the pull request by repo ID and the pull request number.
		FindByNumber(ctx context.Context, repoID, number int64) (*types.PullReq, error)

		// Create a new pull request.
		Create(ctx context.Context, pullreq *types.PullReq) error

		// Update the pull request. It will set new values to the Version and Updated fields.
		Update(ctx context.Context, pr *types.PullReq) error

		// UpdateOptLock the pull request details using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, pr *types.PullReq,
			mutateFn func(pr *types.PullReq) error) (*types.PullReq, error)

		// UpdateActivitySeq the pull request's activity sequence number.
		// It will set new values to the ActivitySeq, Version and Updated fields.
		UpdateActivitySeq(ctx context.Context, pr *types.PullReq) (*types.PullReq, error)

		// Update all PR where target branch points to new SHA
		UpdateMergeCheckStatus(ctx context.Context, targetRepo int64, targetBranch string, status enum.MergeCheckStatus) error

		// Delete the pull request.
		Delete(ctx context.Context, id int64) error

		// Count of pull requests in a space.
		Count(ctx context.Context, opts *types.PullReqFilter) (int64, error)

		// List returns a list of pull requests in a space.
		List(ctx context.Context, opts *types.PullReqFilter) ([]*types.PullReq, error)
	}

	PullReqActivityStore interface {
		// Find the pull request activity by id.
		Find(ctx context.Context, id int64) (*types.PullReqActivity, error)

		// Create a new pull request activity. Value of the Order field should be fetched with UpdateActivitySeq.
		// Value of the SubOrder field (for replies) should be the incremented ReplySeq field (non-replies have 0).
		Create(ctx context.Context, act *types.PullReqActivity) error

		// CreateWithPayload create a new system activity from the provided payload.
		CreateWithPayload(ctx context.Context,
			pr *types.PullReq, principalID int64, payload types.PullReqActivityPayload) (*types.PullReqActivity, error)

		// Update the pull request activity. It will set new values to the Version and Updated fields.
		Update(ctx context.Context, act *types.PullReqActivity) error

		// UpdateOptLock updates the pull request activity using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context,
			act *types.PullReqActivity,
			mutateFn func(act *types.PullReqActivity) error,
		) (*types.PullReqActivity, error)

		// Count returns number of pull request activities in a pull request.
		Count(ctx context.Context, prID int64, opts *types.PullReqActivityFilter) (int64, error)

		// CountUnresolved returns number of unresolved comments.
		CountUnresolved(ctx context.Context, prID int64) (int, error)

		// List returns a list of pull request activities in a pull request (a timeline).
		List(ctx context.Context, prID int64, opts *types.PullReqActivityFilter) ([]*types.PullReqActivity, error)
	}

	// CodeCommentView is to manipulate only code-comment subset of PullReqActivity.
	// It's used by internal service that migrates code comment line numbers after new commits.
	CodeCommentView interface {
		// ListNotAtSourceSHA loads code comments that need to be updated after a new commit.
		// Resulting list is ordered by the file name and the relevant line number.
		ListNotAtSourceSHA(ctx context.Context, prID int64, sourceSHA string) ([]*types.CodeComment, error)

		// ListNotAtMergeBaseSHA loads code comments that need to be updated after merge base update.
		// Resulting list is ordered by the file name and the relevant line number.
		ListNotAtMergeBaseSHA(ctx context.Context, prID int64, targetSHA string) ([]*types.CodeComment, error)

		// UpdateAll updates code comments (pull request activity of types code-comment).
		// entities coming from the input channel.
		UpdateAll(ctx context.Context, codeComments []*types.CodeComment) error
	}

	// PullReqReviewStore defines the pull request review storage.
	PullReqReviewStore interface {
		// Find returns the pull request review entity or an error if it doesn't exist.
		Find(ctx context.Context, id int64) (*types.PullReqReview, error)

		// Create creates a new pull request review.
		Create(ctx context.Context, v *types.PullReqReview) error
	}

	// PullReqReviewerStore defines the pull request reviewer storage.
	PullReqReviewerStore interface {
		// Find returns the pull request reviewer or an error if it doesn't exist.
		Find(ctx context.Context, prID, principalID int64) (*types.PullReqReviewer, error)

		// Create creates the new pull request reviewer.
		Create(ctx context.Context, v *types.PullReqReviewer) error

		// Update updates the pull request reviewer.
		Update(ctx context.Context, v *types.PullReqReviewer) error

		// Delete the Pull request reviewer
		Delete(ctx context.Context, prID, principalID int64) error

		// List returns all pull request reviewers for the pull request.
		List(ctx context.Context, prID int64) ([]*types.PullReqReviewer, error)
	}

	// WebhookStore defines the webhook data storage.
	WebhookStore interface {
		// Find finds the webhook by id.
		Find(ctx context.Context, id int64) (*types.Webhook, error)

		// Create creates a new webhook.
		Create(ctx context.Context, hook *types.Webhook) error

		// Update updates an existing webhook.
		Update(ctx context.Context, hook *types.Webhook) error

		// UpdateOptLock updates the webhook using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, hook *types.Webhook,
			mutateFn func(hook *types.Webhook) error) (*types.Webhook, error)

		// Delete deletes the webhook for the given id.
		Delete(ctx context.Context, id int64) error

		// Count counts the webhooks for a given parent type and id.
		Count(ctx context.Context, parentType enum.WebhookParent, parentID int64,
			opts *types.WebhookFilter) (int64, error)

		// List lists the webhooks for a given parent type and id.
		List(ctx context.Context, parentType enum.WebhookParent, parentID int64,
			opts *types.WebhookFilter) ([]*types.Webhook, error)
	}

	// WebhookExecutionStore defines the webhook execution data storage.
	WebhookExecutionStore interface {
		// Find finds the webhook execution by id.
		Find(ctx context.Context, id int64) (*types.WebhookExecution, error)

		// Create creates a new webhook execution entry.
		Create(ctx context.Context, hook *types.WebhookExecution) error

		// ListForWebhook lists the webhook executions for a given webhook id.
		ListForWebhook(ctx context.Context, webhookID int64,
			opts *types.WebhookExecutionFilter) ([]*types.WebhookExecution, error)

		// ListForTrigger lists the webhook executions for a given trigger id.
		ListForTrigger(ctx context.Context, triggerID string) ([]*types.WebhookExecution, error)
	}

	CheckStore interface {
		// Upsert creates new or updates an existing status check result.
		Upsert(ctx context.Context, check *types.Check) error

		// Count counts status check results for a specific commit in a repo.
		Count(ctx context.Context, repoID int64, commitSHA string, opts types.CheckListOptions) (int, error)

		// List returns a list of status check results for a specific commit in a repo.
		List(ctx context.Context, repoID int64, commitSHA string, opts types.CheckListOptions) ([]types.Check, error)

		// ListRecent returns a list of recently executed status checks in a repository.
		ListRecent(ctx context.Context, repoID int64, since time.Time) ([]string, error)
	}

	ReqCheckStore interface {
		// Create creates new required status check.
		Create(ctx context.Context, reqCheck *types.ReqCheck) error

		// List returns a list of required status checks for a repo.
		List(ctx context.Context, repoID int64) ([]*types.ReqCheck, error)

		// Delete removes a required status checks for a repo.
		Delete(ctx context.Context, repoID, reqCheckID int64) error
	}

	JobStore interface {
		// Find fetches a job by its unique identifier.
		Find(ctx context.Context, uid string) (*types.Job, error)

		// Create is used to create a new job.
		Create(ctx context.Context, job *types.Job) error

		// Upsert will insert the job in the database if the job didn't already exist,
		// or it will update the existing one but only if its definition has changed.
		Upsert(ctx context.Context, job *types.Job) error

		// UpdateDefinition is used to update a job definition.
		UpdateDefinition(ctx context.Context, job *types.Job) error

		// UpdateExecution is used to update a job before and after execution.
		UpdateExecution(ctx context.Context, job *types.Job) error

		// UpdateProgress is used to update a job progress data.
		UpdateProgress(ctx context.Context, job *types.Job) error

		// CountRunning returns number of jobs that are currently being run.
		CountRunning(ctx context.Context) (int, error)

		// ListReady returns a list of jobs that are ready for execution.
		ListReady(ctx context.Context, now time.Time, limit int) ([]*types.Job, error)

		// ListDeadlineExceeded returns a list of jobs that have exceeded their execution deadline.
		ListDeadlineExceeded(ctx context.Context, now time.Time) ([]*types.Job, error)

		// NextScheduledTime returns a scheduled time of the next ready job.
		NextScheduledTime(ctx context.Context, now time.Time) (time.Time, error)

		// DeleteOld removes non-recurring jobs that have finished execution or have failed.
		DeleteOld(ctx context.Context, olderThan time.Time) (int64, error)
	}

	PipelineStore interface {
		// Find returns a pipeline given a pipeline ID from the datastore.
		Find(ctx context.Context, id int64) (*types.Pipeline, error)

		// FindByUID returns a pipeline with a given UID in a space
		FindByUID(ctx context.Context, id int64, uid string) (*types.Pipeline, error)

		// Create creates a new pipeline in the datastore.
		Create(ctx context.Context, pipeline *types.Pipeline) error

		// Update tries to update a pipeline in the datastore
		Update(ctx context.Context, pipeline *types.Pipeline) error

		// List lists the pipelines present in a repository in the datastore.
		List(ctx context.Context, repoID int64, pagination types.ListQueryFilter) ([]*types.Pipeline, error)

		// ListLatest lists the pipelines present in a repository in the datastore.
		// It also returns latest build information for all the returned entries.
		ListLatest(ctx context.Context, repoID int64, pagination types.ListQueryFilter) ([]*types.Pipeline, error)

		// UpdateOptLock updates the pipeline using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, pipeline *types.Pipeline,
			mutateFn func(pipeline *types.Pipeline) error) (*types.Pipeline, error)

		// Delete deletes a pipeline ID from the datastore.
		Delete(ctx context.Context, id int64) error

		// Count the number of pipelines in a repository matching the given filter.
		Count(ctx context.Context, repoID int64, filter types.ListQueryFilter) (int64, error)

		// DeleteByUID deletes a pipeline with a given UID under a repo.
		DeleteByUID(ctx context.Context, repoID int64, uid string) error

		// IncrementSeqNum increments the sequence number of the pipeline
		IncrementSeqNum(ctx context.Context, pipeline *types.Pipeline) (*types.Pipeline, error)
	}

	SecretStore interface {
		// Find returns a secret given an ID
		Find(ctx context.Context, id int64) (*types.Secret, error)

		// FindByUID returns a secret given a space ID and a UID
		FindByUID(ctx context.Context, spaceID int64, uid string) (*types.Secret, error)

		// Create creates a new secret
		Create(ctx context.Context, secret *types.Secret) error

		// Count the number of secrets in a space matching the given filter.
		Count(ctx context.Context, spaceID int64, pagination types.ListQueryFilter) (int64, error)

		// UpdateOptLock updates the secret using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, secret *types.Secret,
			mutateFn func(secret *types.Secret) error) (*types.Secret, error)

		// Update tries to update a secret.
		Update(ctx context.Context, secret *types.Secret) error

		// Delete deletes a secret given an ID.
		Delete(ctx context.Context, id int64) error

		// DeleteByUID deletes a secret given a space ID and a uid
		DeleteByUID(ctx context.Context, spaceID int64, uid string) error

		// List lists the secrets in a given space
		List(ctx context.Context, spaceID int64, filter types.ListQueryFilter) ([]*types.Secret, error)
	}

	ExecutionStore interface {
		// Find returns a execution given a pipeline and an execution number
		Find(ctx context.Context, pipelineID int64, num int64) (*types.Execution, error)

		// Create creates a new execution in the datastore.
		Create(ctx context.Context, execution *types.Execution) error

		// Update tries to update an execution.
		Update(ctx context.Context, execution *types.Execution) error

		// UpdateOptLock updates the execution using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, execution *types.Execution,
			mutateFn func(execution *types.Execution) error) (*types.Execution, error)

		// List lists the executions for a given pipeline ID
		List(ctx context.Context, pipelineID int64, pagination types.Pagination) ([]*types.Execution, error)

		// Delete deletes an execution given a pipeline ID and an execution number
		Delete(ctx context.Context, pipelineID int64, num int64) error

		// Count the number of executions in a space
		Count(ctx context.Context, parentID int64) (int64, error)
	}

	StageStore interface {
		// List returns a build stage list from the datastore
		// where the stage is incomplete (pending or running).
		ListIncomplete(ctx context.Context) ([]*types.Stage, error)

		// ListWithSteps returns a stage list from the datastore corresponding to an execution,
		// with the individual steps included.
		ListWithSteps(ctx context.Context, executionID int64) ([]*types.Stage, error)

		// Find returns a build stage from the datastore by ID.
		Find(ctx context.Context, stageID int64) (*types.Stage, error)

		// FindByNumber returns a stage from the datastore by number.
		FindByNumber(ctx context.Context, executionID int64, stageNum int) (*types.Stage, error)
	}

	StepStore interface {
		// FindByNumber returns a step from the datastore by number.
		FindByNumber(ctx context.Context, stageID int64, stepNum int) (*types.Step, error)
	}

	ConnectorStore interface {
		// Find returns a connector given an ID.
		Find(ctx context.Context, id int64) (*types.Connector, error)

		// FindByUID returns a connector given a space ID and a UID.
		FindByUID(ctx context.Context, spaceID int64, uid string) (*types.Connector, error)

		// Create creates a new connector.
		Create(ctx context.Context, connector *types.Connector) error

		// Count the number of connectors in a space matching the given filter.
		Count(ctx context.Context, spaceID int64, pagination types.ListQueryFilter) (int64, error)

		// UpdateOptLock updates the connector using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, connector *types.Connector,
			mutateFn func(connector *types.Connector) error) (*types.Connector, error)

		// Update tries to update a connector.
		Update(ctx context.Context, connector *types.Connector) error

		// Delete deletes a connector given an ID.
		Delete(ctx context.Context, id int64) error

		// DeleteByUID deletes a connector given a space ID and a uid.
		DeleteByUID(ctx context.Context, spaceID int64, uid string) error

		// List lists the connectors in a given space.
		List(ctx context.Context, spaceID int64, filter types.ListQueryFilter) ([]*types.Connector, error)
	}

	TemplateStore interface {
		// Find returns a template given an ID.
		Find(ctx context.Context, id int64) (*types.Template, error)

		// FindByUID returns a template given a space ID and a UID.
		FindByUID(ctx context.Context, spaceID int64, uid string) (*types.Template, error)

		// Create creates a new template.
		Create(ctx context.Context, template *types.Template) error

		// Count the number of templates in a space matching the given filter.
		Count(ctx context.Context, spaceID int64, pagination types.ListQueryFilter) (int64, error)

		// UpdateOptLock updates the template using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, template *types.Template,
			mutateFn func(template *types.Template) error) (*types.Template, error)

		// Update tries to update a template.
		Update(ctx context.Context, template *types.Template) error

		// Delete deletes a template given an ID.
		Delete(ctx context.Context, id int64) error

		// DeleteByUID deletes a template given a space ID and a uid.
		DeleteByUID(ctx context.Context, spaceID int64, uid string) error

		// List lists the templates in a given space.
		List(ctx context.Context, spaceID int64, filter types.ListQueryFilter) ([]*types.Template, error)
	}

	TriggerStore interface {
		// FindByUID returns a trigger given a pipeline and a trigger UID.
		FindByUID(ctx context.Context, pipelineID int64, uid string) (*types.Trigger, error)

		// Create creates a new trigger in the datastore.
		Create(ctx context.Context, trigger *types.Trigger) error

		// Update tries to update an trigger.
		Update(ctx context.Context, trigger *types.Trigger) error

		// UpdateOptLock updates the trigger using the optimistic locking mechanism.
		UpdateOptLock(ctx context.Context, trigger *types.Trigger,
			mutateFn func(trigger *types.Trigger) error) (*types.Trigger, error)

		// List lists the triggers for a given pipeline ID.
		List(ctx context.Context, pipelineID int64, filter types.ListQueryFilter) ([]*types.Trigger, error)

		// Delete deletes an trigger given a pipeline ID and a trigger UID.
		DeleteByUID(ctx context.Context, pipelineID int64, uid string) error

		// Count the number of triggers in a pipeline.
		Count(ctx context.Context, pipelineID int64, filter types.ListQueryFilter) (int64, error)
	}

	PluginStore interface {
		// List returns back the list of plugins matching the given filter
		// along with their associated schemas.
		List(ctx context.Context, filter types.ListQueryFilter) ([]*types.Plugin, error)

		// Create creates a new entry in the plugin datastore.
		Create(ctx context.Context, plugin *types.Plugin) error

		// Count counts the number of plugins matching the given filter.
		Count(ctx context.Context, filter types.ListQueryFilter) (int64, error)
	}
)
