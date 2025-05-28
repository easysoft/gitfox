// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/git/sha"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ store.CheckStore = (*CheckStoreOrm)(nil)

// NewCheckStoreOrm returns a new CheckStore.
func NewCheckStoreOrm(
	db *gorm.DB,
	pCache store.PrincipalInfoCache,
) *CheckStoreOrm {
	return &CheckStoreOrm{
		db:     db,
		pCache: pCache,
	}
}

// CheckStoreOrm implements store.CheckStore backed by a relational database.
type CheckStoreOrm struct {
	db     *gorm.DB
	pCache store.PrincipalInfoCache
}

type check struct {
	ID             int64                 `gorm:"column:check_id"`
	CreatedBy      int64                 `gorm:"column:check_created_by"`
	Created        int64                 `gorm:"column:check_created"`
	Updated        int64                 `gorm:"column:check_updated"`
	RepoID         int64                 `gorm:"column:check_repo_id"`
	CommitSHA      string                `gorm:"column:check_commit_sha"`
	Identifier     string                `gorm:"column:check_uid"`
	Status         enum.CheckStatus      `gorm:"column:check_status"`
	Summary        string                `gorm:"column:check_summary"`
	Link           string                `gorm:"column:check_link"`
	Payload        json.RawMessage       `gorm:"column:check_payload"`
	Metadata       json.RawMessage       `gorm:"column:check_metadata"`
	PayloadKind    enum.CheckPayloadKind `gorm:"column:check_payload_kind"`
	PayloadVersion string                `gorm:"column:check_payload_version"`
	Started        int64                 `gorm:"column:check_started"`
	Ended          int64                 `gorm:"column:check_ended"`
}

const tableCheck = "checks"

// FindByIdentifier returns status check result for given unique key.
func (s *CheckStoreOrm) FindByIdentifier(
	ctx context.Context,
	repoID int64,
	commitSHA string,
	identifier string,
) (types.Check, error) {
	var obj check
	filter := &check{RepoID: repoID, Identifier: identifier, CommitSHA: commitSHA}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableCheck).Where(filter).First(&obj).Error; err != nil {
		return types.Check{}, database.ProcessGormSQLErrorf(ctx, err, "exec check query failed")
	}

	return s.mapCheck(&obj), nil
}

// Upsert creates new or updates an existing status check result.
func (s *CheckStoreOrm) Upsert(ctx context.Context, check *types.Check) error {
	upsertFields := []string{"check_updated", "check_status", "check_summary", "check_link",
		"check_payload", "check_metadata", "check_payload_kind", "check_payload_version", "check_started", "check_ended"}
	dbObj := s.mapInternalCheck(check)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableCheck).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "check_repo_id"}, {Name: "check_commit_sha"}, {Name: "check_uid"}},
		DoUpdates: clause.AssignmentColumns(upsertFields),
	}).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Upsert query failed")
	}

	check.ID = dbObj.ID
	check.Created = dbObj.Created
	check.CreatedBy = dbObj.CreatedBy
	return nil
}

// Count counts status check results for a specific commit in a repo.
func (s *CheckStoreOrm) Count(ctx context.Context,
	repoID int64,
	commitSHA string,
	opts types.CheckListOptions,
) (int, error) {
	var total int64

	filter := &check{RepoID: repoID, CommitSHA: commitSHA}
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(tableCheck).Where(filter)
	db = s.applyOpts(db, opts.Query)

	err := db.Count(&total).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed to execute count status checks query")
	}

	return int(total), nil
}

// List returns a list of status check results for a specific commit in a repo.
func (s *CheckStoreOrm) List(ctx context.Context,
	repoID int64,
	commitSHA string,
	opts types.CheckListOptions,
) ([]types.Check, error) {
	filter := &check{RepoID: repoID, CommitSHA: commitSHA}
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(tableCheck).Where(filter)
	db = s.applyOpts(db, opts.Query)

	dst := make([]*check, 0)

	if err := db.Limit(int(database.Limit(opts.Size))).
		Offset(int(database.Offset(opts.Page, opts.Size))).
		Order("check_updated desc").Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to execute list status checks query")
	}

	result, err := s.mapSliceCheck(ctx, dst)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListRecent returns a list of recently executed status checks in a repository.
func (s *CheckStoreOrm) ListRecent(ctx context.Context,
	repoID int64,
	opts types.CheckRecentOptions,
) ([]string, error) {
	filter := &check{RepoID: repoID}
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(tableCheck).Where(filter).Where("check_created > ?", opts.Since).Distinct("check_uid")
	db = s.applyOpts(db, opts.Query).Order("check_uid")

	dst := make([]string, 0)

	if err := db.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to execute list recent status checks query")
	}

	return dst, nil
}

// ListResults returns a list of status check results for a specific commit in a repo.
func (s *CheckStoreOrm) ListResults(ctx context.Context,
	repoID int64,
	commitSHA string,
) ([]types.CheckResult, error) {
	filter := &check{RepoID: repoID, CommitSHA: commitSHA}
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(tableCheck).Select("check_uid", "check_status").Where(filter)

	result := make([]types.CheckResult, 0)

	if err := db.Scan(&result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to execute list status checks results query")
	}

	return result, nil
}

// ResultSummary returns a list of status check result summaries for the provided list of commits in a repo.
func (s *CheckStoreOrm) ResultSummary(ctx context.Context,
	repoID int64,
	commitSHAs []string,
) (map[sha.SHA]types.CheckCountSummary, error) {
	type resultRow struct {
		CommitSHA    string `gorm:"column:check_commit_sha"`
		CountPending int    `gorm:"column:count_pending"`
		CountRunning int    `gorm:"column:count_running"`
		CountSuccess int    `gorm:"column:count_success"`
		CountFailure int    `gorm:"column:count_failure"`
		CountError   int    `gorm:"column:count_error"`
	}

	var results []resultRow

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableCheck).
		Select(`check_commit_sha,
			SUM(check_status = 'pending') as count_pending,
			SUM(check_status = 'running') as count_running,
			SUM(check_status = 'success') as count_success,
			SUM(check_status = 'failure') as count_failure,
			SUM(check_status = 'error') as count_error`).
		Where("check_repo_id = ?", repoID).
		Where("check_commit_sha IN ?", commitSHAs).
		Group("check_commit_sha").
		Scan(&results).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to execute status check summary query")
	}

	result := make(map[sha.SHA]types.CheckCountSummary)
	for _, row := range results {
		commitSHA, err := sha.New(row.CommitSHA)
		if err != nil {
			return nil, fmt.Errorf("invalid commit SHA read from DB: %s", row.CommitSHA)
		}

		result[commitSHA] = types.CheckCountSummary{
			Pending: row.CountPending,
			Running: row.CountRunning,
			Success: row.CountSuccess,
			Failure: row.CountFailure,
			Error:   row.CountError,
		}
	}

	return result, nil
}

func (*CheckStoreOrm) applyOpts(db *gorm.DB, query string) (optedDb *gorm.DB) {
	optedDb = db
	if query != "" {
		optedDb = db.Where("LOWER(check_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(query)))
	}
	return
}

func (*CheckStoreOrm) mapCheck(c *check) types.Check {
	return types.Check{
		ID:         c.ID,
		CreatedBy:  c.CreatedBy,
		Created:    c.Created,
		Updated:    c.Updated,
		RepoID:     c.RepoID,
		CommitSHA:  c.CommitSHA,
		Identifier: c.Identifier,
		Status:     c.Status,
		Summary:    c.Summary,
		Link:       c.Link,
		Metadata:   c.Metadata,
		Payload: types.CheckPayload{
			Version: c.PayloadVersion,
			Kind:    c.PayloadKind,
			Data:    c.Payload,
		},
		ReportedBy: nil,
		Started:    c.Started,
		Ended:      c.Ended,
	}
}

func (*CheckStoreOrm) mapInternalCheck(c *types.Check) *check {
	m := &check{
		ID:             c.ID,
		CreatedBy:      c.CreatedBy,
		Created:        c.Created,
		Updated:        c.Updated,
		RepoID:         c.RepoID,
		CommitSHA:      c.CommitSHA,
		Identifier:     c.Identifier,
		Status:         c.Status,
		Summary:        c.Summary,
		Link:           c.Link,
		Payload:        c.Payload.Data,
		Metadata:       c.Metadata,
		PayloadKind:    c.Payload.Kind,
		PayloadVersion: c.Payload.Version,
		Started:        c.Started,
		Ended:          c.Ended,
	}

	return m
}

func (s *CheckStoreOrm) mapSliceCheck(ctx context.Context, checks []*check) ([]types.Check, error) {
	// collect all principal IDs
	ids := make([]int64, len(checks))
	for i, req := range checks {
		ids[i] = req.CreatedBy
	}

	// pull principal infos from cache
	infoMap, err := s.pCache.Map(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load status check principal reporters: %w", err)
	}

	// attach the principal infos back to the slice items
	m := make([]types.Check, len(checks))
	for i, c := range checks {
		m[i] = s.mapCheck(c)
		if reportedBy, ok := infoMap[c.CreatedBy]; ok {
			m[i].ReportedBy = reportedBy
		}
	}

	return m, nil
}
