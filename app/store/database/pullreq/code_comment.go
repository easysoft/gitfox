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
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var _ store.CodeCommentView = (*CodeCommentView)(nil)

// NewCodeCommentOrmView returns a new CodeCommentView.
func NewCodeCommentOrmView(db *gorm.DB) *CodeCommentView {
	return &CodeCommentView{
		db: db,
	}
}

// CodeCommentView implements store.CodeCommentView backed by a relational database.
type CodeCommentView struct {
	db *gorm.DB
}

// ListNotAtSourceSHA lists all code comments not already at the provided source SHA.
func (s *CodeCommentView) ListNotAtSourceSHA(ctx context.Context,
	prID int64, sourceSHA string,
) ([]*types.CodeComment, error) {
	return s.list(ctx, prID, "", sourceSHA)
}

// ListNotAtMergeBaseSHA lists all code comments not already at the provided merge base SHA.
func (s *CodeCommentView) ListNotAtMergeBaseSHA(ctx context.Context,
	prID int64, mergeBaseSHA string,
) ([]*types.CodeComment, error) {
	return s.list(ctx, prID, mergeBaseSHA, "")
}

// list is used by internal service that updates line numbers of code comments after
// branch updates and requires either mergeBaseSHA or sourceSHA but not both.
// Resulting list is ordered by the file name and the relevant line number.
func (s *CodeCommentView) list(ctx context.Context,
	prID int64, mergeBaseSHA, sourceSHA string,
) ([]*types.CodeComment, error) {
	const codeCommentColumns = `
		 pullreq_activity_id
		,pullreq_activity_version
		,pullreq_activity_updated
		,coalesce(pullreq_activity_outdated, false) as "pullreq_activity_outdated"
		,coalesce(pullreq_activity_code_comment_merge_base_sha, '') as "pullreq_activity_code_comment_merge_base_sha"
		,coalesce(pullreq_activity_code_comment_source_sha, '') as "pullreq_activity_code_comment_source_sha"
		,coalesce(pullreq_activity_code_comment_path, '') as "pullreq_activity_code_comment_path"
		,coalesce(pullreq_activity_code_comment_line_new, 1) as "pullreq_activity_code_comment_line_new"
		,coalesce(pullreq_activity_code_comment_span_new, 0) as "pullreq_activity_code_comment_span_new"
		,coalesce(pullreq_activity_code_comment_line_old, 1) as "pullreq_activity_code_comment_line_old"
		,coalesce(pullreq_activity_code_comment_span_old, 0) as "pullreq_activity_code_comment_span_old"`

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableActivity).
		Select(codeCommentColumns).
		Where("pullreq_activity_pullreq_id = ?", prID).
		Where("not pullreq_activity_outdated").
		Where("pullreq_activity_type = ?", enum.PullReqActivityTypeCodeComment).
		Where("pullreq_activity_kind = ?", enum.PullReqActivityKindChangeComment).
		Where("pullreq_activity_deleted is null and pullreq_activity_parent_id is null")

	if mergeBaseSHA != "" {
		stmt = stmt.
			Where("pullreq_activity_code_comment_merge_base_sha <> ?", mergeBaseSHA)
	} else {
		stmt = stmt.
			Where("pullreq_activity_code_comment_source_sha <> ?", sourceSHA)
	}

	stmt = stmt.Order("pullreq_activity_code_comment_path asc, pullreq_activity_code_comment_line_new asc")

	result := make([]*types.CodeComment, 0)

	if err := stmt.Find(&result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing code comment list query")
	}

	return result, nil
}

// UpdateAll updates all code comments provided in the slice.
func (s *CodeCommentView) UpdateAll(ctx context.Context, codeComments []*types.CodeComment) error {
	if len(codeComments) == 0 {
		return nil
	}

	updatedAt := time.Now()

	updateFields := []string{"Version", "Updated", "Outdated", "MergeBaseSHA", "SourceSHA", "Path",
		"LineNew", "SpanNew", "LineOld", "SpanOld",
	}

	for _, codeComment := range codeComments {
		codeComment.Version++
		codeComment.Updated = updatedAt.UnixMilli()

		res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableActivity).
			Where(&types.CodeComment{ID: codeComment.ID, Version: codeComment.Version - 1}).
			Select(updateFields).Updates(codeComment)

		if res.Error != nil {
			return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update code comment=%d", codeComment.ID)
		}

		if res.RowsAffected == 0 {
			log.Ctx(ctx).Warn().Msgf("Version conflict when trying to update code comment=%d", codeComment.ID)
			continue
		}
	}

	return nil
}
