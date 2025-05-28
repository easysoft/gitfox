// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pullreq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var _ store.PullReqActivityStore = (*ActivityOrmStore)(nil)

// NewPullReqActivityOrmStore returns a new PullReqJournalStore.
func NewPullReqActivityOrmStore(
	db *gorm.DB,
	pCache store.PrincipalInfoCache,
) *ActivityOrmStore {
	return &ActivityOrmStore{
		db:     db,
		pCache: pCache,
	}
}

// ActivityOrmStore implements store.PullReqActivityStore backed by a relational database.
type ActivityOrmStore struct {
	db     *gorm.DB
	pCache store.PrincipalInfoCache
}

// journal is used to fetch pull request data from the database.
// The object should be later re-packed into a different struct to return it as an API response.
type pullReqActivity struct {
	ID      int64 `gorm:"column:pullreq_activity_id;primaryKey"`
	Version int64 `gorm:"column:pullreq_activity_version"`

	CreatedBy int64    `gorm:"column:pullreq_activity_created_by"`
	Created   int64    `gorm:"column:pullreq_activity_created"`
	Updated   int64    `gorm:"column:pullreq_activity_updated"`
	Edited    int64    `gorm:"column:pullreq_activity_edited"`
	Deleted   null.Int `gorm:"column:pullreq_activity_deleted"`

	ParentID  null.Int `gorm:"column:pullreq_activity_parent_id"`
	RepoID    int64    `gorm:"column:pullreq_activity_repo_id"`
	PullReqID int64    `gorm:"column:pullreq_activity_pullreq_id"`

	Order    int64 `gorm:"column:pullreq_activity_order"`
	SubOrder int64 `gorm:"column:pullreq_activity_sub_order"`
	ReplySeq int64 `gorm:"column:pullreq_activity_reply_seq"`

	Type enum.PullReqActivityType `gorm:"column:pullreq_activity_type"`
	Kind enum.PullReqActivityKind `gorm:"column:pullreq_activity_kind"`

	Text     string          `gorm:"column:pullreq_activity_text"`
	Payload  json.RawMessage `gorm:"column:pullreq_activity_payload"`
	Metadata json.RawMessage `gorm:"column:pullreq_activity_metadata"`

	ResolvedBy null.Int `gorm:"column:pullreq_activity_resolved_by"`
	Resolved   null.Int `gorm:"column:pullreq_activity_resolved"`

	Outdated                null.Bool   `gorm:"column:pullreq_activity_outdated"`
	CodeCommentMergeBaseSHA null.String `gorm:"column:pullreq_activity_code_comment_merge_base_sha"`
	CodeCommentSourceSHA    null.String `gorm:"column:pullreq_activity_code_comment_source_sha"`
	CodeCommentPath         null.String `gorm:"column:pullreq_activity_code_comment_path"`
	CodeCommentLineNew      null.Int    `gorm:"column:pullreq_activity_code_comment_line_new"`
	CodeCommentSpanNew      null.Int    `gorm:"column:pullreq_activity_code_comment_span_new"`
	CodeCommentLineOld      null.Int    `gorm:"column:pullreq_activity_code_comment_line_old"`
	CodeCommentSpanOld      null.Int    `gorm:"column:pullreq_activity_code_comment_span_old"`
}

const (
	tableActivity = "pullreq_activities"
)

// Find finds the pull request activity by id.
func (s *ActivityOrmStore) Find(ctx context.Context, id int64) (*types.PullReqActivity, error) {
	dst := &pullReqActivity{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableActivity).First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pull request activity")
	}
	act, err := s.mapPullReqActivity(ctx, dst)
	if err != nil {
		return nil, fmt.Errorf("failed to map pull request activity: %w", err)
	}

	return act, nil
}

// Create creates a new pull request.
func (s *ActivityOrmStore) Create(ctx context.Context, act *types.PullReqActivity) error {
	dbObj, err := mapInternalPullReqActivity(act)
	if err != nil {
		return fmt.Errorf("failed to map pull request activity: %w", err)
	}

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableActivity).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to insert pull request activity")
	}

	act.ID = dbObj.ID
	return nil
}

func (s *ActivityOrmStore) CreateWithPayload(ctx context.Context,
	pr *types.PullReq, principalID int64, payload types.PullReqActivityPayload, metadata *types.PullReqActivityMetadata,
) (*types.PullReqActivity, error) {
	now := time.Now().UnixMilli()
	act := &types.PullReqActivity{
		CreatedBy: principalID,
		Created:   now,
		Updated:   now,
		Edited:    now,
		RepoID:    pr.TargetRepoID,
		PullReqID: pr.ID,
		Order:     pr.ActivitySeq,
		SubOrder:  0,
		ReplySeq:  0,
		Type:      payload.ActivityType(),
		Kind:      enum.PullReqActivityKindSystem,
		Text:      "",
		Metadata:  metadata,
	}

	_ = act.SetPayload(payload)

	err := s.Create(ctx, act)
	if err != nil {
		err = fmt.Errorf("failed to write pull request system '%s' activity: %w", payload.ActivityType(), err)
		return nil, err
	}

	return act, nil
}

// Update updates the pull request.
func (s *ActivityOrmStore) Update(ctx context.Context, act *types.PullReqActivity) error {
	updatedAt := time.Now()

	dbAct, err := mapInternalPullReqActivity(act)
	if err != nil {
		return fmt.Errorf("failed to map pull request activity: %w", err)
	}

	dbAct.Version++
	dbAct.Updated = updatedAt.UnixMilli()

	updateFields := []string{"Version", "Updated", "Edited", "Deleted", "ReplySeq", "Text",
		"Payload", "Metadata", "ResolvedBy", "Resolved", "Outdated", "CodeCommentMergeBaseSHA",
		"CodeCommentSourceSHA", "CodeCommentPath", "CodeCommentLineNew", "CodeCommentSpanNew",
		"CodeCommentLineOld", "CodeCommentSpanOld",
	}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableActivity).
		Where(&pullReqActivity{ID: act.ID, Version: dbAct.Version - 1}).
		Select(updateFields).Updates(dbAct)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update pull request activity")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrVersionConflict
	}

	updatedAct, err := s.mapPullReqActivity(ctx, dbAct)
	if err != nil {
		return fmt.Errorf("failed to map db pull request activity: %w", err)
	}
	*act = *updatedAct

	return nil
}

// UpdateOptLock updates the pull request using the optimistic locking mechanism.
func (s *ActivityOrmStore) UpdateOptLock(ctx context.Context,
	act *types.PullReqActivity,
	mutateFn func(act *types.PullReqActivity) error,
) (*types.PullReqActivity, error) {
	for {
		dup := *act

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

		act, err = s.Find(ctx, act.ID)
		if err != nil {
			return nil, err
		}
	}
}

// Count of pull requests for a repo.
func (s *ActivityOrmStore) Count(ctx context.Context,
	prID int64,
	opts *types.PullReqActivityFilter,
) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableActivity).Where("pullreq_activity_pullreq_id = ?", prID)

	if len(opts.Types) == 1 {
		stmt = stmt.Where("pullreq_activity_type = ?", opts.Types[0])
	} else if len(opts.Types) > 1 {
		stmt = stmt.Where("pullreq_activity_type IN ?", opts.Types)
	}

	if len(opts.Kinds) == 1 {
		stmt = stmt.Where("pullreq_activity_kind = ?", opts.Kinds[0])
	} else if len(opts.Kinds) > 1 {
		stmt = stmt.Where("pullreq_activity_kind IN ?", opts.Kinds)
	}

	if opts.After != 0 {
		stmt = stmt.Where("pullreq_activity_created > ?", opts.After)
	}

	if opts.Before != 0 {
		stmt = stmt.Where("pullreq_activity_created < ?", opts.Before)
	}

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}

	return count, nil
}

// List returns a list of pull request activities for a PR.
func (s *ActivityOrmStore) List(ctx context.Context,
	prID int64,
	filter *types.PullReqActivityFilter,
) ([]*types.PullReqActivity, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableActivity).Where("pullreq_activity_pullreq_id = ?", prID)

	stmt = applyFilter(filter, stmt)

	stmt = stmt.Order("pullreq_activity_order asc, pullreq_activity_sub_order asc")

	dst := make([]*pullReqActivity, 0)

	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing pull request activity list query")
	}

	result, err := s.mapSlicePullReqActivity(ctx, dst)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListAuthorIDs returns a list of pull request activity author ids in a thread for a PR.
func (s *ActivityOrmStore) ListAuthorIDs(ctx context.Context, prID int64, order int64) ([]int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableActivity).Where(&pullReqActivity{PullReqID: prID, Order: order})

	var dst []int64

	if err := stmt.Distinct("pullreq_activity_created_by").Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing pull request activity list query")
	}

	return dst, nil
}

func (s *ActivityOrmStore) CountUnresolved(ctx context.Context, prID int64) (int, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableActivity).
		Where("pullreq_activity_pullreq_id = ?", prID).
		Where("pullreq_activity_sub_order = 0").
		Where("pullreq_activity_resolved IS NULL").
		Where("pullreq_activity_deleted IS NULL").
		Where("pullreq_activity_kind <> ?", enum.PullReqActivityKindSystem)

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count unresolved query")
	}

	return int(count), nil
}

func mapPullReqActivity(act *pullReqActivity) (*types.PullReqActivity, error) {
	metadata := &types.PullReqActivityMetadata{}
	err := json.Unmarshal(act.Metadata, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
	}

	m := &types.PullReqActivity{
		ID:         act.ID,
		Version:    act.Version,
		CreatedBy:  act.CreatedBy,
		Created:    act.Created,
		Updated:    act.Updated,
		Edited:     act.Edited,
		Deleted:    act.Deleted.Ptr(),
		ParentID:   act.ParentID.Ptr(),
		RepoID:     act.RepoID,
		PullReqID:  act.PullReqID,
		Order:      act.Order,
		SubOrder:   act.SubOrder,
		ReplySeq:   act.ReplySeq,
		Type:       act.Type,
		Kind:       act.Kind,
		Text:       act.Text,
		PayloadRaw: act.Payload,
		Metadata:   metadata,
		ResolvedBy: act.ResolvedBy.Ptr(),
		Resolved:   act.Resolved.Ptr(),
		Author:     types.PrincipalInfo{},
		Resolver:   nil,
	}
	if m.Type == enum.PullReqActivityTypeCodeComment && m.Kind == enum.PullReqActivityKindChangeComment {
		m.CodeComment = &types.CodeCommentFields{
			Outdated:     act.Outdated.Bool,
			MergeBaseSHA: act.CodeCommentMergeBaseSHA.String,
			SourceSHA:    act.CodeCommentSourceSHA.String,
			Path:         act.CodeCommentPath.String,
			LineNew:      int(act.CodeCommentLineNew.Int64),
			SpanNew:      int(act.CodeCommentSpanNew.Int64),
			LineOld:      int(act.CodeCommentLineOld.Int64),
			SpanOld:      int(act.CodeCommentSpanOld.Int64),
		}
	}

	return m, nil
}

func mapInternalPullReqActivity(act *types.PullReqActivity) (*pullReqActivity, error) {
	m := &pullReqActivity{
		ID:         act.ID,
		Version:    act.Version,
		CreatedBy:  act.CreatedBy,
		Created:    act.Created,
		Updated:    act.Updated,
		Edited:     act.Edited,
		Deleted:    null.IntFromPtr(act.Deleted),
		ParentID:   null.IntFromPtr(act.ParentID),
		RepoID:     act.RepoID,
		PullReqID:  act.PullReqID,
		Order:      act.Order,
		SubOrder:   act.SubOrder,
		ReplySeq:   act.ReplySeq,
		Type:       act.Type,
		Kind:       act.Kind,
		Text:       act.Text,
		Payload:    act.PayloadRaw,
		Metadata:   nil,
		ResolvedBy: null.IntFromPtr(act.ResolvedBy),
		Resolved:   null.IntFromPtr(act.Resolved),
	}
	if act.IsValidCodeComment() {
		m.Outdated = null.BoolFrom(act.CodeComment.Outdated)
		m.CodeCommentMergeBaseSHA = null.StringFrom(act.CodeComment.MergeBaseSHA)
		m.CodeCommentSourceSHA = null.StringFrom(act.CodeComment.SourceSHA)
		m.CodeCommentPath = null.StringFrom(act.CodeComment.Path)
		m.CodeCommentLineNew = null.IntFrom(int64(act.CodeComment.LineNew))
		m.CodeCommentSpanNew = null.IntFrom(int64(act.CodeComment.SpanNew))
		m.CodeCommentLineOld = null.IntFrom(int64(act.CodeComment.LineOld))
		m.CodeCommentSpanOld = null.IntFrom(int64(act.CodeComment.SpanOld))
	}

	var err error
	m.Metadata, err = json.Marshal(act.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metadata: %w", err)
	}

	return m, nil
}

func (s *ActivityOrmStore) mapPullReqActivity(ctx context.Context, act *pullReqActivity) (*types.PullReqActivity, error) {
	m, err := mapPullReqActivity(act)
	if err != nil {
		return nil, err
	}
	var author, resolver *types.PrincipalInfo

	author, err = s.pCache.Get(ctx, act.CreatedBy)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to load PR activity author")
	}
	if author != nil {
		m.Author = *author
	}

	if act.ResolvedBy.Valid {
		resolver, err = s.pCache.Get(ctx, act.ResolvedBy.Int64)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to load PR activity resolver")
		}
		m.Resolver = resolver
	}

	return m, nil
}

func (s *ActivityOrmStore) mapSlicePullReqActivity(
	ctx context.Context,
	activities []*pullReqActivity,
) ([]*types.PullReqActivity, error) {
	// collect all principal IDs
	ids := make([]int64, 0, 2*len(activities))
	for _, act := range activities {
		ids = append(ids, act.CreatedBy)
		if act.ResolvedBy.Valid {
			ids = append(ids, act.ResolvedBy.Int64)
		}
	}

	// pull principal infos from cache
	infoMap, err := s.pCache.Map(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load PR principal infos: %w", err)
	}

	// attach the principal infos back to the slice items
	m := make([]*types.PullReqActivity, len(activities))
	for i, act := range activities {
		m[i], err = mapPullReqActivity(act)
		if err != nil {
			return nil, fmt.Errorf("failed to map pull request activity %d: %w", act.ID, err)
		}
		if author, ok := infoMap[act.CreatedBy]; ok {
			m[i].Author = *author
		}
		if act.ResolvedBy.Valid {
			if merger, ok := infoMap[act.ResolvedBy.Int64]; ok {
				m[i].Resolver = merger
			}
		}
	}

	return m, nil
}

func applyFilter(
	filter *types.PullReqActivityFilter,
	stmt *gorm.DB,
) *gorm.DB {
	if len(filter.Types) == 1 {
		stmt = stmt.Where("pullreq_activity_type = ?", filter.Types[0])
	} else if len(filter.Types) > 1 {
		stmt = stmt.Where("pullreq_activity_type IN ?", filter.Types)
	}

	if len(filter.Kinds) == 1 {
		stmt = stmt.Where("pullreq_activity_kind = ?", filter.Kinds[0])
	} else if len(filter.Kinds) > 1 {
		stmt = stmt.Where("pullreq_activity_kind IN ?", filter.Kinds)
	}

	if filter.After != 0 {
		stmt = stmt.Where("pullreq_activity_created > ?", filter.After)
	}

	if filter.Before != 0 {
		stmt = stmt.Where("pullreq_activity_created < ?", filter.Before)
	}

	if filter.Limit > 0 {
		stmt = stmt.Limit(int(database.Limit(filter.Limit)))
	}

	return stmt
}
