// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/easysoft/gitfox/app/store"
	storagedriver "github.com/easysoft/gitfox/pkg/storage/driver"
	"github.com/easysoft/gitfox/types"
)

type ResumeRequester interface {
	Start(ctx context.Context) error
	Resume(ctx context.Context) error
	Status(ctx context.Context) (ResponseHeaderWriter, error)
	Write(ctx context.Context, req *http.Request) (ResponseHeaderWriter, error)
	Finish(ctx context.Context) (ResponseHeaderWriter, error)
	Cancel(ctx context.Context) error
}

type resumeRequest struct {
	view          *types.ArtifactView
	storageId     int64
	artifactStore store.ArtifactStore
}

type ChunkStatus string

var (
	ChunkStatusStart    ChunkStatus = "start"
	ChunkStatusAppend   ChunkStatus = "append"
	ChunkStatusFinished ChunkStatus = "finished"
)

type ChunkDescriptor struct {
	UniqueID string
	Status   ChunkStatus
}

type ResumeUploader struct {
	resumeRequest
	writer     storagedriver.FileWriter
	Descriptor *ChunkDescriptor
}

func NewResumeUploader(view *types.ArtifactView, artifactStore store.ArtifactStore, storageId int64) *ResumeUploader {
	return &ResumeUploader{
		resumeRequest: resumeRequest{
			view: view, storageId: storageId, artifactStore: artifactStore,
		},
	}
}

func (r *ResumeUploader) RegisterWriter(writer storagedriver.FileWriter) {
	r.writer = writer
}

func (r *ResumeUploader) Commit(ctx context.Context) error {
	switch r.Descriptor.Status {
	case ChunkStatusStart:
		return r.artifactStore.Blobs().Create(ctx, &types.ArtifactBlob{
			StorageID: r.storageId, Ref: r.Descriptor.UniqueID, Created: time.Now().UnixMilli(),
		})
	case ChunkStatusAppend:
		return r.writer.Commit(ctx)
	case ChunkStatusFinished:
		err := r.writer.Commit(ctx)
		if err != nil {
			return err
		}
		//r.artifactStore.Blobs().GetByRef()
	default:
		return fmt.Errorf("unknown status")
	}
	return nil
}

func (r *ResumeUploader) Cancel(ctx context.Context) error {
	return r.writer.Cancel(ctx)
}
