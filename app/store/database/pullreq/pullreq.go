// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pullreq

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/Masterminds/squirrel"
	"github.com/guregu/null"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var _ store.PullReqStore = (*OrmStore)(nil)

// NewPullReqOrmStore returns a new PullReqStore.
func NewPullReqOrmStore(db *gorm.DB,
	pCache store.PrincipalInfoCache) *OrmStore {
	return &OrmStore{
		db:     db,
		pCache: pCache,
	}
}

// OrmStore implements store.PullReqStore backed by a relational database.
type OrmStore struct {
	db     *gorm.DB
	pCache store.PrincipalInfoCache
}

// pullReq is used to fetch pull request data from the database.
// The object should be later re-packed into a different struct to return it as an API response.
type pullReq struct {
	ID      int64 `gorm:"column:pullreq_id;primaryKey"`
	Version int64 `gorm:"column:pullreq_version"`
	Number  int64 `gorm:"column:pullreq_number"`

	CreatedBy int64    `gorm:"column:pullreq_created_by"`
	Created   int64    `gorm:"column:pullreq_created"`
	Updated   int64    `gorm:"column:pullreq_updated"`
	Edited    int64    `gorm:"column:pullreq_edited"`
	Closed    null.Int `gorm:"column:pullreq_closed"`

	State   enum.PullReqState `gorm:"column:pullreq_state"`
	IsDraft string            `gorm:"column:pullreq_is_draft"`

	CommentCount    int `gorm:"column:pullreq_comment_count"`
	UnresolvedCount int `gorm:"column:pullreq_unresolved_count"`

	Title       string `gorm:"column:pullreq_title"`
	Description string `gorm:"column:pullreq_description"`

	SourceRepoID int64  `gorm:"column:pullreq_source_repo_id"`
	SourceBranch string `gorm:"column:pullreq_source_branch"`
	SourceSHA    string `gorm:"column:pullreq_source_sha"`
	TargetRepoID int64  `gorm:"column:pullreq_target_repo_id"`
	TargetBranch string `gorm:"column:pullreq_target_branch"`

	ActivitySeq int64 `gorm:"column:pullreq_activity_seq"`

	MergedBy    null.Int    `gorm:"column:pullreq_merged_by"`
	Merged      null.Int    `gorm:"column:pullreq_merged"`
	MergeMethod null.String `gorm:"column:pullreq_merge_method"`

	MergeTargetSHA null.String `gorm:"column:pullreq_merge_target_sha"`
	MergeBaseSHA   string      `gorm:"column:pullreq_merge_base_sha"`
	MergeSHA       null.String `gorm:"column:pullreq_merge_sha"`

	MergeCheckStatus  enum.MergeCheckStatus `gorm:"column:pullreq_merge_check_status"`
	MergeConflicts    null.String           `gorm:"column:pullreq_merge_conflicts"`
	RebaseCheckStatus enum.MergeCheckStatus `gorm:"column:pullreq_rebase_check_status"`
	RebaseConflicts   null.String           `gorm:"column:pullreq_rebase_conflicts"`

	CommitCount null.Int             `gorm:"column:pullreq_commit_count"`
	FileCount   null.Int             `gorm:"column:pullreq_file_count"`
	Flow        enum.PullRequestFlow `gorm:"column:pullreq_flow;NOT NULL;DEFAULT 0"`
	Additions   null.Int             `gorm:"column:pullreq_additions"`
	Deletions   null.Int             `gorm:"column:pullreq_deletions"`
}

const (
	tablePullReq = "pullreqs"
)

// Find finds the pull request by id.
func (s *OrmStore) Find(ctx context.Context, id int64) (*types.PullReq, error) {
	dst := &pullReq{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pull request")
	}

	return s.mapPullReq(ctx, dst), nil
}

func (s *OrmStore) findByNumberInternal(
	ctx context.Context,
	repoID,
	number int64,
	lock bool,
) (*types.PullReq, error) {
	// todo table lock with orm
	// FindByNumberWithLock doesn't be used,
	//if lock && !strings.HasPrefix(s.db.Name(), "sqlite") {
	//	sqlQuery += "\n" + database.SQLForUpdate
	//}

	dst := &pullReq{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).Where(&pullReq{
		TargetRepoID: repoID, Number: number,
	}).First(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pull request by number")
	}

	return s.mapPullReq(ctx, dst), nil
}

// FindByNumberWithLock finds the pull request by repo ID and pull request number
// and locks the pull request for the duration of the transaction.
func (s *OrmStore) FindByNumberWithLock(
	ctx context.Context,
	repoID,
	number int64,
) (*types.PullReq, error) {
	return s.findByNumberInternal(ctx, repoID, number, true)
}

// FindByNumber finds the pull request by repo ID and pull request number.
func (s *OrmStore) FindByNumber(ctx context.Context, repoID, number int64) (*types.PullReq, error) {
	return s.findByNumberInternal(ctx, repoID, number, false)
}

func (s *OrmStore) findOpenedBySourceTargetBranch(ctx context.Context, sourceRepo, targetRepo int64,
	sourceBranch, targetBranch string) (*types.PullReq, bool, error) {
	q := pullReq{
		SourceRepoID: sourceRepo, SourceBranch: sourceBranch,
		TargetRepoID: targetRepo, TargetBranch: targetBranch,
		State: enum.PullReqStateOpen,
	}
	dst := &pullReq{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).Where(&q).First(dst).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pull request by source/target")
	}
	return s.mapPullReq(ctx, dst), true, nil
}

func (s *OrmStore) GetUnmergedPullRequest(ctx context.Context, sourceRepo, targetRepo int64,
	sourceBranch, targetBranch string, flow enum.PullRequestFlow) (*types.PullReq, bool, error) {
	q := pullReq{
		SourceRepoID: sourceRepo, SourceBranch: sourceBranch,
		TargetRepoID: targetRepo, TargetBranch: targetBranch,
		Flow: flow,
	}
	dst := &pullReq{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).Where(&q).First(dst).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pull request by source/target")
	}
	return s.mapPullReq(ctx, dst), true, nil
}

// Stream returns a list of pull requests for a repo.
func (s *OrmStore) Stream(ctx context.Context, opts *types.PullReqFilter) (<-chan *types.PullReq, <-chan error) {
	stmt := s.listQuery(opts)

	stmt = stmt.Order("pullreq_updated desc")

	chPRs := make(chan *types.PullReq)
	chErr := make(chan error, 1)

	go func() {
		defer close(chPRs)
		defer close(chErr)

		db := s.db.WithContext(ctx)

		rows, err := db.Raw(stmt.Statement.SQL.String(), stmt.Statement.Vars...).Rows()
		if err != nil {
			chErr <- database.ProcessGormSQLErrorf(ctx, err, "Failed to execute stream query")
			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var prData pullReq
			err = db.ScanRows(rows, &prData)
			if err != nil {
				chErr <- fmt.Errorf("failed to scan pull request: %w", err)
				return
			}

			chPRs <- s.mapPullReq(ctx, &prData)
		}

		if err := rows.Err(); err != nil {
			chErr <- fmt.Errorf("failed to scan pull request: %w", err)
		}
	}()

	return chPRs, chErr
}

func (s *OrmStore) ListOpenByBranchName(
	ctx context.Context,
	repoID int64,
	branchNames []string,
) (map[string][]*types.PullReq, error) {
	var dst []*pullReq

	err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).
		Where("pullreq_source_repo_id = ? AND pullreq_state = ? AND pullreq_source_branch IN ?", repoID, enum.PullReqStateOpen, branchNames).
		Order("pullreq_updated desc").
		Find(&dst).Error
	if err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to fetch list of PRs by branch")
	}

	prMap := make(map[string][]*types.PullReq)
	for _, prDB := range dst {
		pr := s.mapPullReq(ctx, prDB)
		prMap[prDB.SourceBranch] = append(prMap[prDB.SourceBranch], pr)
	}

	return prMap, nil
}

func (s *OrmStore) listQuery(opts *types.PullReqFilter) *gorm.DB {
	db := s.db

	// columns := pullReqColumnsNoDescription
	// if opts.IncludeDescription {
	// 	columns = pullReqColumns
	// }

	// db = db.Select(columns).Table(tablePullReq)

	db = db.Table(tablePullReq)

	s.applyFilter(db, opts)

	return db
}

func (s *OrmStore) applyFilter(db *gorm.DB, opts *types.PullReqFilter) *gorm.DB {
	if len(opts.States) == 1 {
		db = db.Where("pullreq_state = ?", opts.States[0])
	} else if len(opts.States) > 1 {
		db = db.Where("pullreq_state IN ?", opts.States)
	}

	if opts.SourceRepoID != 0 {
		db = db.Where("pullreq_source_repo_id = ?", opts.SourceRepoID)
	}

	if opts.SourceBranch != "" {
		db = db.Where("pullreq_source_branch = ?", opts.SourceBranch)
	}

	if opts.TargetRepoID != 0 {
		db = db.Where("pullreq_target_repo_id = ?", opts.TargetRepoID)
	}

	if opts.TargetBranch != "" {
		db = db.Where("pullreq_target_branch = ?", opts.TargetBranch)
	}

	if opts.Query != "" {
		db = db.Where("LOWER(pullreq_title) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(opts.Query)))
	}

	if len(opts.CreatedBy) > 0 {
		db = db.Where("pullreq_created_by IN ?", opts.CreatedBy)
	}

	if opts.CreatedLt > 0 {
		db = db.Where("pullreq_created < ?", opts.CreatedLt)
	}

	if opts.CreatedGt > 0 {
		db = db.Where("pullreq_created > ?", opts.CreatedGt)
	}

	if opts.UpdatedLt > 0 {
		db = db.Where("pullreq_updated < ?", opts.UpdatedLt)
	}

	if opts.UpdatedGt > 0 {
		db = db.Where("pullreq_updated > ?", opts.UpdatedGt)
	}

	if opts.EditedLt > 0 {
		db = db.Where("pullreq_edited < ?", opts.EditedLt)
	}

	if opts.EditedGt > 0 {
		db = db.Where("pullreq_edited > ?", opts.EditedGt)
	}

	if len(opts.SpaceIDs) == 1 {
		db = db.InnerJoins("repositories ON repo_id = pullreq_target_repo_id")
		db = db.Where("repo_parent_id = ?", opts.SpaceIDs[0])
	} else if len(opts.SpaceIDs) > 1 {
		db = db.InnerJoins("repositories ON repo_id = pullreq_target_repo_id")
		db = db.Where("repo_parent_id in ?", opts.SpaceIDs)
	}

	if len(opts.RepoIDBlacklist) == 1 {
		db = db.Where("pullreq_target_repo_id <> ?", opts.RepoIDBlacklist[0])
	} else if len(opts.RepoIDBlacklist) > 1 {
		db = db.Where("pullreq_target_repo_id in ?", opts.RepoIDBlacklist)
	}

	if opts.AuthorID > 0 {
		db = db.Where("pullreq_created_by = ?", opts.AuthorID)
	}

	if opts.CommenterID > 0 {
		db = db.InnerJoins("pullreq_activities act_com ON act_com.pullreq_activity_pullreq_id = pullreq_id")
		db = db.Where("act_com.pullreq_activity_deleted IS NULL")
		db = db.Where("(" +
			"act_com.pullreq_activity_kind = '" + string(enum.PullReqActivityKindComment) + "' OR " +
			"act_com.pullreq_activity_kind = '" + string(enum.PullReqActivityKindChangeComment) + "')")
		db = db.Where("act_com.pullreq_activity_created_by = ?", opts.CommenterID)
	}

	if opts.ReviewerID > 0 {
		db = db.InnerJoins(
			fmt.Sprintf("pullreq_reviewers ON "+
				"pullreq_reviewer_pullreq_id = pullreq_id AND pullreq_reviewer_principal_id = %d", opts.ReviewerID))
		if len(opts.ReviewDecisions) > 0 {
			db = db.Where(squirrel.Eq{"pullreq_reviewer_review_decision": opts.ReviewDecisions})
		}
	}

	// TODO: 与上游不一致
	// if opts.MentionedID > 0 {
	// 	db = db.InnerJoins("pullreq_activities act_ment ON act_ment.pullreq_activity_pullreq_id = pullreq_id")
	// 	db = db.Where("act_ment.pullreq_activity_deleted IS NULL")
	// 	db = db.Where("(" +
	// 		"act_ment.pullreq_activity_kind = '" + string(enum.PullReqActivityKindComment) + "' OR " +
	// 		"act_ment.pullreq_activity_kind = '" + string(enum.PullReqActivityKindChangeComment) + "')")

	// 	switch s.db.DriverName() {
	// 	case database.SqliteDriverName:
	// 		db = db.InnerJoins(
	// 			"json_each(json_extract(act_ment.pullreq_activity_metadata, '$.mentions.ids')) as mentions")
	// 		db = db.Where("mentions.value = ?", opts.MentionedID)
	// 	case database.PostgresDriverName:
	// 		db = db.Where(fmt.Sprintf(
	// 			"act_ment.pullreq_activity_metadata->'mentions'->'ids' @> ('[%d]')::jsonb",
	// 			opts.MentionedID))
	// 	}
	// }

	// labels

	if len(opts.LabelID) == 0 && len(opts.ValueID) == 0 {
		return db
	}

	db = db.Joins("INNER JOIN pullreq_labels ON pullreq_label_pullreq_id = pullreq_id").
		Group("pullreq_id")

	switch {
	case len(opts.LabelID) > 0 && len(opts.ValueID) == 0:
		db = db.Where("pullreq_label_label_id IN ?", opts.LabelID)

	case len(opts.LabelID) == 0 && len(opts.ValueID) > 0:
		db = db.Where("pullreq_label_label_value_id IN ?", opts.ValueID)

	default:
		db = db.Where("pullreq_label_label_id IN ? OR pullreq_label_label_value_id IN ?", opts.LabelID, opts.ValueID)
	}

	db = db.Having("COUNT(pullreq_label_pullreq_id) = ?", len(opts.LabelID)+len(opts.ValueID))

	return db
}

// Create creates a new pull request.
func (s *OrmStore) Create(ctx context.Context, pr *types.PullReq) error {
	dbObj := mapInternalPullReq(pr)

	_, found, err := s.findOpenedBySourceTargetBranch(ctx, dbObj.SourceRepoID, dbObj.TargetRepoID, dbObj.SourceBranch, dbObj.TargetBranch)
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "check confict failed")
	}

	if found {
		err = gorm.ErrDuplicatedKey
		return database.ProcessGormSQLErrorf(ctx, err, "confict opened pr for source/target repo and branch")
	}

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	pr.ID = dbObj.ID
	return nil
}

// Update updates the pull request.
func (s *OrmStore) Update(ctx context.Context, pr *types.PullReq) error {
	updatedAt := time.Now().UnixMilli()

	dbPR := mapInternalPullReq(pr)
	dbPR.Version++
	dbPR.Updated = updatedAt
	dbPR.Edited = updatedAt

	updateFields := []string{"Version", "Updated", "Edited", "State", "IsDraft", "CommentCount",
		"UnresolvedCount", "Title", "Description", "ActivitySeq", "SourceSHA", "MergedBy", "Merged", "MergeMethod",
		"MergeCheckStatus", "MergeTargetSHA", "MergeBaseSHA", "MergeSHA",
		"MergeConflicts", "CommitCount", "FileCount",
	}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).
		Where(&pullReq{ID: pr.ID, Version: dbPR.Version - 1}).
		Select(updateFields).Updates(dbPR)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update pull request activity")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrVersionConflict
	}

	*pr = *s.mapPullReq(ctx, dbPR)

	return nil
}

// UpdateOptLock the pull request details using the optimistic locking mechanism.
func (s *OrmStore) UpdateOptLock(ctx context.Context, pr *types.PullReq,
	mutateFn func(pr *types.PullReq) error,
) (*types.PullReq, error) {
	for {
		dup := *pr

		err := mutateFn(&dup)
		if err != nil {
			return nil, err
		}

		err = s.Update(ctx, &dup)
		if err == nil {
			return &dup, nil
		}
		if !errors.Is(err, gitfox_store.ErrVersionConflict) {
			return nil, err
		}

		pr, err = s.Find(ctx, pr.ID)
		if err != nil {
			return nil, err
		}
	}
}

// UpdateActivitySeq updates the pull request's activity sequence.
func (s *OrmStore) UpdateActivitySeq(ctx context.Context, pr *types.PullReq) (*types.PullReq, error) {
	return s.UpdateOptLock(ctx, pr, func(pr *types.PullReq) error {
		pr.ActivitySeq++
		return nil
	})
}

// ResetMergeCheckStatus resets the pull request's mergeability status to unchecked
// for all pr which target branch points to targetBranch.
func (s *OrmStore) ResetMergeCheckStatus(
	ctx context.Context,
	targetRepo int64,
	targetBranch string,
) error {
	// NOTE: keep pullreq_merge_base_sha on old value as it's a required field.
	const query = `
	UPDATE pullreqs
	SET
		 pullreq_updated = ?
		,pullreq_version = pullreq_version + 1
		,pullreq_merge_target_sha = NULL
		,pullreq_merge_sha = NULL
		,pullreq_merge_check_status = ?
		,pullreq_merge_conflicts = NULL
		,pullreq_rebase_check_status = ?
		,pullreq_rebase_conflicts = NULL
		,pullreq_commit_count = NULL
		,pullreq_file_count = NULL
	WHERE pullreq_target_repo_id = ? AND
		pullreq_target_branch = ? AND
		pullreq_state not in (?, ?)`

	now := time.Now().UnixMilli()

	err := dbtx.GetOrmAccessor(ctx, s.db).Exec(query, now, enum.MergeCheckStatusUnchecked, enum.MergeCheckStatusUnchecked, targetRepo, targetBranch,
		enum.PullReqStateClosed, enum.PullReqStateMerged).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to reset mergeable status check in pull requests")
	}

	return nil
}

// Delete the pull request.
func (s *OrmStore) Delete(ctx context.Context, id int64) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).Delete(&pullReq{ID: id}).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "the delete query failed")
	}

	return nil
}

// Count of pull requests for a repo.
func (s *OrmStore) Count(ctx context.Context, opts *types.PullReqFilter) (int64, error) {
	q := pullReq{
		SourceRepoID: opts.SourceRepoID, SourceBranch: opts.SourceBranch,
		TargetRepoID: opts.TargetRepoID, TargetBranch: opts.TargetBranch,
	}
	// TODO: 与上游不一致, 但是这里的逻辑是正确的
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).Where(&q)
	if len(opts.LabelID) > 0 || len(opts.ValueID) > 0 || opts.CommenterID > 0 || opts.MentionedID > 0 {
		stmt = stmt.Select("count(DISTINCT pullreq_id)")
	} else {
		stmt = stmt.Select("count(*)")
	}

	s.applyFilter(stmt, opts)
	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}

	return count, nil
}

// List returns a list of pull requests for a repo.
func (s *OrmStore) List(ctx context.Context, opts *types.PullReqFilter) ([]*types.PullReq, error) {
	stmt := s.listQuery(opts)
	stmt = stmt.Limit(database.GormLimit(opts.Size))
	stmt = stmt.Offset(database.GormOffset(opts.Page, opts.Size))

	// NOTE: string concatenation is safe because the
	// order attribute is an enum and is not user-defined,
	// and is therefore not subject to injection attacks.
	opts.Sort, _ = opts.Sort.Sanitize()
	stmt = stmt.Order("pullreq_" + string(opts.Sort) + " " + opts.Order.String())

	dst := make([]*pullReq, 0)
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	result, err := s.mapSlicePullReq(ctx, dst)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func mapPullReq(pr *pullReq) *types.PullReq {
	var mergeConflicts, rebaseConflicts []string
	if pr.MergeConflicts.Valid {
		mergeConflicts = strings.Split(pr.MergeConflicts.String, "\n")
	}
	if pr.RebaseConflicts.Valid {
		rebaseConflicts = strings.Split(pr.RebaseConflicts.String, "\n")
	}

	return &types.PullReq{
		ID:                pr.ID,
		Version:           pr.Version,
		Number:            pr.Number,
		CreatedBy:         pr.CreatedBy,
		Created:           pr.Created,
		Updated:           pr.Updated,
		Edited:            pr.Edited, // TODO: When we remove the DB column, make Edited equal to Updated
		Closed:            pr.Closed.Ptr(),
		State:             pr.State,
		IsDraft:           strings.EqualFold(strings.ToLower(pr.IsDraft), "true"),
		CommentCount:      pr.CommentCount,
		UnresolvedCount:   pr.UnresolvedCount,
		Title:             pr.Title,
		Description:       pr.Description,
		SourceRepoID:      pr.SourceRepoID,
		SourceBranch:      pr.SourceBranch,
		SourceSHA:         pr.SourceSHA,
		TargetRepoID:      pr.TargetRepoID,
		TargetBranch:      pr.TargetBranch,
		ActivitySeq:       pr.ActivitySeq,
		MergedBy:          pr.MergedBy.Ptr(),
		Merged:            pr.Merged.Ptr(),
		MergeMethod:       (*enum.MergeMethod)(pr.MergeMethod.Ptr()),
		MergeCheckStatus:  pr.MergeCheckStatus,
		MergeTargetSHA:    pr.MergeTargetSHA.Ptr(),
		MergeBaseSHA:      pr.MergeBaseSHA,
		MergeSHA:          pr.MergeSHA.Ptr(),
		MergeConflicts:    mergeConflicts,
		RebaseCheckStatus: pr.RebaseCheckStatus,
		RebaseConflicts:   rebaseConflicts,
		Author:            types.PrincipalInfo{},
		Merger:            nil,
		Stats: types.PullReqStats{
			Conversations:   pr.CommentCount,
			UnresolvedCount: pr.UnresolvedCount,
			DiffStats: types.DiffStats{
				Commits:      pr.CommitCount.Ptr(),
				FilesChanged: pr.FileCount.Ptr(),
				Additions:    pr.Additions.Ptr(),
				Deletions:    pr.Deletions.Ptr(),
			},
		},
		Flow: pr.Flow,
	}
}

func mapInternalPullReq(pr *types.PullReq) *pullReq {
	mergeConflicts := strings.Join(pr.MergeConflicts, "\n")
	rebaseConflicts := strings.Join(pr.RebaseConflicts, "\n")
	m := &pullReq{
		ID:                pr.ID,
		Version:           pr.Version,
		Number:            pr.Number,
		CreatedBy:         pr.CreatedBy,
		Created:           pr.Created,
		Updated:           pr.Updated,
		Edited:            pr.Edited, // TODO: When we remove the DB column, make Edited equal to Updated
		Closed:            null.IntFromPtr(pr.Closed),
		State:             pr.State,
		IsDraft:           strconv.FormatBool(pr.IsDraft),
		CommentCount:      pr.CommentCount,
		UnresolvedCount:   pr.UnresolvedCount,
		Title:             pr.Title,
		Description:       pr.Description,
		SourceRepoID:      pr.SourceRepoID,
		SourceBranch:      pr.SourceBranch,
		SourceSHA:         pr.SourceSHA,
		TargetRepoID:      pr.TargetRepoID,
		TargetBranch:      pr.TargetBranch,
		ActivitySeq:       pr.ActivitySeq,
		MergedBy:          null.IntFromPtr(pr.MergedBy),
		Merged:            null.IntFromPtr(pr.Merged),
		MergeMethod:       null.StringFromPtr((*string)(pr.MergeMethod)),
		MergeCheckStatus:  pr.MergeCheckStatus,
		MergeTargetSHA:    null.StringFromPtr(pr.MergeTargetSHA),
		MergeBaseSHA:      pr.MergeBaseSHA,
		MergeSHA:          null.StringFromPtr(pr.MergeSHA),
		MergeConflicts:    null.NewString(mergeConflicts, mergeConflicts != ""),
		RebaseCheckStatus: pr.RebaseCheckStatus,
		RebaseConflicts:   null.NewString(rebaseConflicts, rebaseConflicts != ""),
		CommitCount:       null.IntFromPtr(pr.Stats.Commits),
		FileCount:         null.IntFromPtr(pr.Stats.FilesChanged),
		Additions:         null.IntFromPtr(pr.Stats.Additions),
		Deletions:         null.IntFromPtr(pr.Stats.Deletions),
		Flow:              pr.Flow,
	}

	return m
}

func (s *OrmStore) mapPullReq(ctx context.Context, pr *pullReq) *types.PullReq {
	m := mapPullReq(pr)

	var author, merger *types.PrincipalInfo
	var err error

	author, err = s.pCache.Get(ctx, pr.CreatedBy)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to load PR author")
	}
	if author != nil {
		m.Author = *author
	}

	if pr.MergedBy.Valid {
		merger, err = s.pCache.Get(ctx, pr.MergedBy.Int64)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to load PR merger")
		}
		m.Merger = merger
	}

	return m
}

func (s *OrmStore) mapSlicePullReq(ctx context.Context, prs []*pullReq) ([]*types.PullReq, error) {
	// collect all principal IDs
	ids := make([]int64, 0, 2*len(prs))
	for _, pr := range prs {
		ids = append(ids, pr.CreatedBy)
		if pr.MergedBy.Valid {
			ids = append(ids, pr.MergedBy.Int64)
		}
	}

	// pull principal infos from cache
	infoMap, err := s.pCache.Map(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load PR principal infos: %w", err)
	}

	// attach the principal infos back to the slice items
	m := make([]*types.PullReq, len(prs))
	for i, pr := range prs {
		m[i] = mapPullReq(pr)
		if author, ok := infoMap[pr.CreatedBy]; ok {
			m[i].Author = *author
		}
		if pr.MergedBy.Valid {
			if merger, ok := infoMap[pr.MergedBy.Int64]; ok {
				m[i].Merger = merger
			}
		}
	}

	return m, nil
}

func (s *OrmStore) SummaryCount(ctx context.Context, opts *types.PullReqSummaryFilter) (*types.PullReqSummary, error) {
	type Result struct {
		Flow  enum.PullRequestFlow `gorm:"column:pullreq_flow"`
		Count int64                `gorm:"column:count"`
	}
	var results []Result
	err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePullReq).
		Where("pullreq_target_repo_id = ?", opts.RepoID).
		Where("pullreq_created >= ?", opts.Begin).
		Where("pullreq_created <= ?", opts.End).
		Select("pullreq_flow, COUNT(*) AS count").
		Group("pullreq_flow").
		Scan(&results).Error
	if err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	if len(results) == 0 {
		return &types.PullReqSummary{
			PullReqCount: 0,
			PushReqCount: 0,
			Total:        0,
		}, nil
	}
	var total, pullReqCount, pushReqCount int64
	for _, result := range results {
		count := result.Count
		total += count
		switch result.Flow {
		case enum.PullRequestFlowZentao:
			pushReqCount = count
		default:
			pullReqCount = count
		}
	}
	return &types.PullReqSummary{
		PullReqCount: int(pullReqCount),
		PushReqCount: int(pushReqCount),
		Total:        int(total),
	}, nil
}
