// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pullreq

import (
	"context"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ store.PullReqFileViewStore = (*FileViewOrmStore)(nil)

// NewPullReqFileViewOrmStore returns a new PullReqFileViewOrmStore.
func NewPullReqFileViewOrmStore(
	db *gorm.DB,
) *FileViewOrmStore {
	return &FileViewOrmStore{
		db: db,
	}
}

// FileViewOrmStore implements store.PullReqFileViewStore backed by a relational database.
type FileViewOrmStore struct {
	db *gorm.DB
}

type pullReqFileView struct {
	PullReqID   int64 `gorm:"column:pullreq_file_view_pullreq_id"`
	PrincipalID int64 `gorm:"column:pullreq_file_view_principal_id"`

	Path     string `gorm:"column:pullreq_file_view_path"`
	SHA      string `gorm:"column:pullreq_file_view_sha"`
	Obsolete bool   `gorm:"column:pullreq_file_view_obsolete"`

	Created int64 `gorm:"column:pullreq_file_view_created"`
	Updated int64 `gorm:"column:pullreq_file_view_updated"`
}

const (
	tableFileViews = "pullreq_file_views"
)

// FindByFileForPrincipal get the entry for the specified PR, principal, and file.
func (s *FileViewOrmStore) FindByFileForPrincipal(
	ctx context.Context,
	prID int64,
	principalID int64,
	filePath string,
) (*types.PullReqFileView, error) {
	q := pullReqFileView{PullReqID: prID, PrincipalID: principalID, Path: filePath}

	dst := new(pullReqFileView)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableFileViews).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "delete query failed")
	}

	return mapToPullreqFileView(dst), nil
}

// Upsert inserts or updates the latest viewed sha for a file in a PR.
func (s *FileViewOrmStore) Upsert(ctx context.Context, view *types.PullReqFileView) error {
	upsertFields := []string{"pullreq_file_view_updated", "pullreq_file_view_sha", "pullreq_file_view_obsolete"}
	dbObj := mapToInternalPullreqFileView(view)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableFileViews).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "pullreq_file_view_pullreq_id"},
			{Name: "pullreq_file_view_principal_id"},
			{Name: "pullreq_file_view_path"}},
		DoUpdates: clause.AssignmentColumns(upsertFields),
	}).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Upsert query failed")
	}

	view.Created = dbObj.Created
	return nil
}

// DeleteByFileForPrincipal deletes the entry for the specified PR, principal, and file.
func (s *FileViewOrmStore) DeleteByFileForPrincipal(
	ctx context.Context,
	prID int64,
	principalID int64,
	filePath string,
) error {
	q := pullReqFileView{PullReqID: prID, PrincipalID: principalID, Path: filePath}

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableFileViews).Where(&q).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "delete query failed")
	}

	return nil
}

// MarkObsolete updates all entries of the files as obsolete for the PR.
func (s *FileViewOrmStore) MarkObsolete(ctx context.Context, prID int64, filePaths []string) error {
	err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableFileViews).Where("pullreq_file_view_pullreq_id = ?", prID).
		Where("pullreq_file_view_path IN ?", filePaths).
		Where("pullreq_file_view_obsolete = ?", false).
		Updates(map[string]interface{}{
			"pullreq_file_view_obsolete": true,
			"pullreq_file_view_updated":  time.Now().UnixMilli(),
		}).Error

	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "failed to execute update query")
	}

	return nil
}

// List lists all files marked as viewed by the user for the specified PR.
func (s *FileViewOrmStore) List(
	ctx context.Context,
	prID int64,
	principalID int64,
) ([]*types.PullReqFileView, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableFileViews).
		Where(&pullReqFileView{PullReqID: prID, PrincipalID: principalID})

	var dst []*pullReqFileView
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to execute list query")
	}

	return mapToPullreqFileViews(dst), nil
}

func mapToInternalPullreqFileView(view *types.PullReqFileView) *pullReqFileView {
	return &pullReqFileView{
		PullReqID:   view.PullReqID,
		PrincipalID: view.PrincipalID,
		Path:        view.Path,
		SHA:         view.SHA,
		Obsolete:    view.Obsolete,
		Created:     view.Created,
		Updated:     view.Updated,
	}
}

func mapToPullreqFileView(view *pullReqFileView) *types.PullReqFileView {
	return &types.PullReqFileView{
		PullReqID:   view.PullReqID,
		PrincipalID: view.PrincipalID,
		Path:        view.Path,
		SHA:         view.SHA,
		Obsolete:    view.Obsolete,
		Created:     view.Created,
		Updated:     view.Updated,
	}
}

func mapToPullreqFileViews(views []*pullReqFileView) []*types.PullReqFileView {
	m := make([]*types.PullReqFileView, len(views))
	for i, view := range views {
		m[i] = mapToPullreqFileView(view)
	}
	return m
}
