// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package adapter

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/easysoft/gitfox/pkg/storage"
	storagedriver "github.com/easysoft/gitfox/pkg/storage/driver"
	"github.com/easysoft/gitfox/types"

	uuid "github.com/satori/go.uuid"
)

type StorageFileWriter struct {
	Layout StorageLayout
	Writer storagedriver.FileWriter

	Key  string
	Path string
}

func (s *StorageFileWriter) SetPath(p string) {
	s.Path = p
}

type StorageFileWriterList []*StorageFileWriter

func NewStorageFileWriter(ctx context.Context, path string, append bool, store storage.ContentStorage) (*StorageFileWriter, error) {
	fileWriter, err := store.GetWriter(ctx, path, append)
	if err != nil {
		return nil, err
	}
	return &StorageFileWriter{
		Writer: fileWriter,
		Key:    path,
	}, nil
}

func NewStorageBlobWriter(ctx context.Context, store storage.ContentStorage, ref string, append bool) (*StorageFileWriter, error) {
	fileWriter, err := store.GetWriter(ctx, BlobPath(ref), append)
	if err != nil {
		return nil, err
	}
	return &StorageFileWriter{
		Writer: fileWriter,
		Key:    ref,
	}, nil
}

func NewRandomBlobWriter(ctx context.Context, store storage.ContentStorage) (storagedriver.FileWriter, string, error) {
	ref := GenerateSha256()
	storePath := BlobPath(ref)

	_, err := store.Stat(ctx, storePath)
	if err == nil {
		return nil, "", ErrStorageFileAlreadyExists
	}

	var notFoundErr storagedriver.PathNotFoundError
	if errors.As(err, &notFoundErr) {
		fileWriter, e := store.GetWriter(ctx, storePath, false)
		if e != nil {
			return nil, ref, e
		}
		return fileWriter, ref, nil
	}

	return nil, ref, err
}

func NewTemporaryWriter(ctx context.Context, store storage.ContentStorage, temporaryId string, append bool) (storagedriver.FileWriter, error) {
	temporaryPath := TemporaryPath(temporaryId)
	fileWriter, err := store.GetWriter(ctx, temporaryPath, append)
	if err != nil {
		return nil, err
	}
	return fileWriter, nil
}

type StorageLayout string

const (
	StorageLayoutBlob StorageLayout = "blob"
	StorageLayoutPath StorageLayout = "path"
)

func BlobPath(p string) string {
	if len(p) < 4 {
		return p
	}

	if strings.Contains(p, "/") {
		return p
	}

	return path.Join(p[0:2], p[2:4], p)
}

func GenerateSha256() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

func TemporaryPath(p string) string {
	return path.Join("_uploads", BlobPath(strings.ReplaceAll(p, "-", "")))
}

type ViewDescriptor struct {
	ViewID    int64
	OwnerID   int64
	StorageID int64
	Store     storage.ContentStorage
	Space     *types.Space
}
