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

package trigger

import (
	"context"

	gitevents "github.com/easysoft/gitfox/app/events/git"
	pullreqevents "github.com/easysoft/gitfox/app/events/pullreq"
	"github.com/easysoft/gitfox/app/pipeline/commit"
	"github.com/easysoft/gitfox/app/pipeline/triggerer"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/events"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideService,
)

func ProvideService(
	ctx context.Context,
	config Config,
	triggerStore store.TriggerStore,
	commitSvc commit.Service,
	pullReqStore store.PullReqStore,
	repoStore store.RepoStore,
	pipelineStore store.PipelineStore,
	triggerSvc triggerer.Triggerer,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullReqEvFactory *events.ReaderFactory[*pullreqevents.Reader],
) (*Service, error) {
	return New(ctx, config, triggerStore, pullReqStore, repoStore, pipelineStore, triggerSvc,
		commitSvc, gitReaderFactory, pullReqEvFactory)
}
