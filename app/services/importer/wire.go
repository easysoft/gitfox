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

package importer

import (
	"github.com/easysoft/gitfox/app/services/keywordsearch"
	"github.com/easysoft/gitfox/app/services/publicaccess"
	"github.com/easysoft/gitfox/app/sse"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/audit"
	"github.com/easysoft/gitfox/encrypt"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/job"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideRepoImporter,
)

func ProvideRepoImporter(
	config *types.Config,
	urlProvider url.Provider,
	git git.Interface,
	tx dbtx.Transactor,
	repoStore store.RepoStore,
	pipelineStore store.PipelineStore,
	triggerStore store.TriggerStore,
	encrypter encrypt.Encrypter,
	scheduler *job.Scheduler,
	executor *job.Executor,
	sseStreamer sse.Streamer,
	indexer keywordsearch.Indexer,
	publicAccess publicaccess.Service,
	auditService audit.Service,
) (*Repository, error) {
	importer := &Repository{
		defaultBranch: config.Git.DefaultBranch,
		urlProvider:   urlProvider,
		git:           git,
		tx:            tx,
		repoStore:     repoStore,
		pipelineStore: pipelineStore,
		triggerStore:  triggerStore,
		encrypter:     encrypter,
		scheduler:     scheduler,
		sseStreamer:   sseStreamer,
		indexer:       indexer,
		publicAccess:  publicAccess,
		auditService:  auditService,
	}

	err := executor.Register(jobType, importer)
	if err != nil {
		return nil, err
	}

	return importer, nil
}
