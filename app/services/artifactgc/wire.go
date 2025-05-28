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

package artifactgc

import (
	"github.com/easysoft/gitfox/app/services/settings"
	"github.com/easysoft/gitfox/app/sse"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/job"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/store/database/dbtx"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideArtifactSweepSvc,
)

func ProvideArtifactSweepSvc(
	tx dbtx.Transactor,
	artStore store.ArtifactStore,
	fileStore storage.ContentStorage,
	settings *settings.Service,
	scheduler *job.Scheduler,
	executor *job.Executor,
	sseStreamer sse.Streamer,
) (*Service, error) {
	svc := NewService(tx, artStore, fileStore, settings, scheduler, sseStreamer)

	err := executor.Register(JobTypeArtifactContainerGC, NewContainerJob(svc))
	if err != nil {
		return nil, err
	}

	err = executor.Register(JobTypeArtifactHardRemove, NewSoftRemoveJob(svc))
	if err != nil {
		return nil, err
	}

	return svc, nil
}
