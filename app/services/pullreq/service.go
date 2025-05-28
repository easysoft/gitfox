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
	"fmt"
	"sync"
	"time"

	"github.com/easysoft/gitfox/app/bootstrap"
	gitevents "github.com/easysoft/gitfox/app/events/git"
	pullreqevents "github.com/easysoft/gitfox/app/events/pullreq"
	"github.com/easysoft/gitfox/app/githook"
	"github.com/easysoft/gitfox/app/services/codecomments"
	"github.com/easysoft/gitfox/app/sse"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/events"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/pubsub"
	"github.com/easysoft/gitfox/stream"
	"github.com/easysoft/gitfox/types"
)

type Service struct {
	pullreqEvReporter   *pullreqevents.Reporter
	git                 git.Interface
	repoGitInfoCache    store.RepoGitInfoCache
	repoStore           store.RepoStore
	pullreqStore        store.PullReqStore
	activityStore       store.PullReqActivityStore
	codeCommentView     store.CodeCommentView
	principalInfoCache  store.PrincipalInfoCache
	codeCommentMigrator *codecomments.Migrator
	fileViewStore       store.PullReqFileViewStore
	sseStreamer         sse.Streamer
	urlProvider         url.Provider

	cancelMutex        sync.Mutex
	cancelMergeability map[string]context.CancelFunc

	pubsub pubsub.PubSub
}

//nolint:funlen // needs refactoring
func New(ctx context.Context,
	config *types.Config,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullreqEvReaderFactory *events.ReaderFactory[*pullreqevents.Reader],
	pullreqEvReporter *pullreqevents.Reporter,
	git git.Interface,
	repoGitInfoCache store.RepoGitInfoCache,
	repoStore store.RepoStore,
	pullreqStore store.PullReqStore,
	activityStore store.PullReqActivityStore,
	codeCommentView store.CodeCommentView,
	codeCommentMigrator *codecomments.Migrator,
	fileViewStore store.PullReqFileViewStore,
	principalInfoCache store.PrincipalInfoCache,
	bus pubsub.PubSub,
	urlProvider url.Provider,
	sseStreamer sse.Streamer,
) (*Service, error) {
	service := &Service{
		pullreqEvReporter:   pullreqEvReporter,
		git:                 git,
		repoGitInfoCache:    repoGitInfoCache,
		repoStore:           repoStore,
		pullreqStore:        pullreqStore,
		activityStore:       activityStore,
		principalInfoCache:  principalInfoCache,
		codeCommentView:     codeCommentView,
		urlProvider:         urlProvider,
		codeCommentMigrator: codeCommentMigrator,
		fileViewStore:       fileViewStore,
		cancelMergeability:  make(map[string]context.CancelFunc),
		pubsub:              bus,
		sseStreamer:         sseStreamer,
	}

	var err error

	// handle git branch events to trigger specific pull request events

	const groupGit = "gitfox:pullreq:git"
	_, err = gitReaderFactory.Launch(ctx, groupGit, config.InstanceID,
		func(r *gitevents.Reader) error {
			const idleTimeout = 15 * time.Second
			r.Configure(
				stream.WithConcurrency(1),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(3),
				))

			_ = r.RegisterBranchUpdated(service.triggerPREventOnBranchUpdate)
			_ = r.RegisterBranchDeleted(service.closePullReqOnBranchDelete)

			return nil
		})
	if err != nil {
		return nil, err
	}

	// pull request ref maintenance

	const groupPullReqHeadRef = "gitfox:pullreq:headref"
	_, err = pullreqEvReaderFactory.Launch(ctx, groupPullReqHeadRef, config.InstanceID,
		func(r *pullreqevents.Reader) error {
			const idleTimeout = 10 * time.Second
			r.Configure(
				stream.WithConcurrency(1),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(3),
				))

			_ = r.RegisterCreated(service.createHeadRefOnCreated)
			_ = r.RegisterBranchUpdated(service.updateHeadRefOnBranchUpdate)
			_ = r.RegisterReopened(service.updateHeadRefOnReopen)

			return nil
		})
	if err != nil {
		return nil, err
	}

	// pull request file viewed maintenance

	const groupPullReqFileViewed = "gitfox:pullreq:fileviewed"
	_, err = pullreqEvReaderFactory.Launch(ctx, groupPullReqFileViewed, config.InstanceID,
		func(r *pullreqevents.Reader) error {
			const idleTimeout = 30 * time.Second
			r.Configure(
				stream.WithConcurrency(3),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(1),
				))

			_ = r.RegisterBranchUpdated(service.handleFileViewedOnBranchUpdate)

			return nil
		})
	if err != nil {
		return nil, err
	}

	const groupPullReqCounters = "gitfox:pullreq:counters"
	_, err = pullreqEvReaderFactory.Launch(ctx, groupPullReqCounters, config.InstanceID,
		func(r *pullreqevents.Reader) error {
			const idleTimeout = 10 * time.Second
			r.Configure(
				stream.WithConcurrency(1),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(2),
				))

			_ = r.RegisterCreated(service.updatePRCountersOnCreated)
			_ = r.RegisterReopened(service.updatePRCountersOnReopened)
			_ = r.RegisterClosed(service.updatePRCountersOnClosed)
			_ = r.RegisterMerged(service.updatePRCountersOnMerged)

			return nil
		})
	if err != nil {
		return nil, err
	}

	// mergeability check
	const groupPullReqMergeable = "gitfox:pullreq:mergeable"
	_, err = pullreqEvReaderFactory.Launch(ctx, groupPullReqMergeable, config.InstanceID,
		func(r *pullreqevents.Reader) error {
			const idleTimeout = 30 * time.Second
			r.Configure(
				stream.WithConcurrency(3),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(2),
				))

			_ = r.RegisterCreated(service.mergeCheckOnCreated)
			_ = r.RegisterBranchUpdated(service.mergeCheckOnBranchUpdate)
			_ = r.RegisterReopened(service.mergeCheckOnReopen)
			_ = r.RegisterClosed(service.mergeCheckOnClosed)
			_ = r.RegisterMerged(service.mergeCheckOnMerged)

			return nil
		})
	if err != nil {
		return nil, err
	}

	// cancel any previous pr mergeability check
	// payload is oldsha.
	_ = bus.Subscribe(ctx, cancelMergeCheckKey, func(payload []byte) error {
		oldSHA := string(payload)
		if oldSHA == "" {
			return nil
		}

		service.cancelMutex.Lock()
		defer service.cancelMutex.Unlock()

		cancel := service.cancelMergeability[oldSHA]
		if cancel != nil {
			cancel()
		}

		delete(service.cancelMergeability, oldSHA)

		return nil
	}, pubsub.WithChannelNamespace("pullreq"))

	// mergeability check
	const groupPullReqCodeComments = "gitfox:pullreq:codecomments"
	_, err = pullreqEvReaderFactory.Launch(ctx, groupPullReqCodeComments, config.InstanceID,
		func(r *pullreqevents.Reader) error {
			const idleTimeout = 10 * time.Second
			r.Configure(
				stream.WithConcurrency(3),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(2),
				))

			_ = r.RegisterBranchUpdated(service.updateCodeCommentsOnBranchUpdate)
			_ = r.RegisterReopened(service.updateCodeCommentsOnReopen)

			return nil
		})
	if err != nil {
		return nil, err
	}

	return service, nil
}

// createSystemRPCWriteParams creates base write parameters for write operations.
func createSystemRPCWriteParams(
	ctx context.Context,
	urlProvider url.Provider,
	repoID int64,
	repoGITUID string,
) (git.WriteParams, error) {
	principal := bootstrap.NewSystemServiceSession().Principal

	// generate envars (add everything githook CLI needs for execution)
	envVars, err := githook.GenerateEnvironmentVariables(
		ctx,
		urlProvider.GetInternalAPIURL(ctx),
		repoID,
		principal.ID,
		false,
		true,
	)
	if err != nil {
		return git.WriteParams{}, fmt.Errorf("failed to generate git hook environment variables: %w", err)
	}

	return git.WriteParams{
		Actor: git.Identity{
			Name:  principal.DisplayName,
			Email: principal.Email,
		},
		RepoUID: repoGITUID,
		EnvVars: envVars,
	}, nil
}
