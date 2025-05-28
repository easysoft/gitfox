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

	repoevents "github.com/easysoft/gitfox/app/events/repo"
	"github.com/easysoft/gitfox/app/services/locker"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/events"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/job"
	"github.com/easysoft/gitfox/types"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideCalculator,
	ProvideService,
)

func ProvideCalculator(
	config *types.Config,
	git git.Interface,
	repoStore store.RepoStore,
	scheduler *job.Scheduler,
	executor *job.Executor,
) (*SizeCalculator, error) {
	job := &SizeCalculator{
		enabled:    config.RepoSize.Enabled,
		cron:       config.RepoSize.CRON,
		maxDur:     config.RepoSize.MaxDuration,
		numWorkers: config.RepoSize.NumWorkers,
		git:        git,
		repoStore:  repoStore,
		scheduler:  scheduler,
	}

	err := executor.Register(jobType, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func ProvideService(ctx context.Context,
	config *types.Config,
	repoEvReporter *repoevents.Reporter,
	repoReaderFactory *events.ReaderFactory[*repoevents.Reader],
	repoStore store.RepoStore,
	urlProvider url.Provider,
	git git.Interface,
	locker *locker.Locker,
) (*Service, error) {
	return NewService(ctx, config, repoEvReporter, repoReaderFactory,
		repoStore, urlProvider, git, locker)
}
