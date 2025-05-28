// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package labels

import (
	"context"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

const (
	labelColumns = `
		 label_space_id
		,label_repo_id
		,label_scope
		,label_key
		,label_description
		,label_type
		,label_color
		,label_created
		,label_updated
		,label_created_by
		,label_updated_by`

	labelSelectBase = `SELECT label_id, ` + labelColumns + ` FROM labels`
)

type label struct {
	ID          int64           `gorm:"column:label_id"`
	SpaceID     null.Int        `gorm:"column:label_space_id"`
	RepoID      null.Int        `gorm:"column:label_repo_id"`
	Scope       int64           `gorm:"column:label_scope"`
	Key         string          `gorm:"column:label_key"`
	Description string          `gorm:"column:label_description"`
	Type        enum.LabelType  `gorm:"column:label_type"`
	Color       enum.LabelColor `gorm:"column:label_color"`
	ValueCount  int64           `gorm:"column:label_value_count"`
	Created     int64           `gorm:"column:label_created"`
	Updated     int64           `gorm:"column:label_updated"`
	CreatedBy   int64           `gorm:"column:label_created_by"`
	UpdatedBy   int64           `gorm:"column:label_updated_by"`
}

type labelInfo struct {
	LabelID    int64           `gorm:"column:label_id"`
	SpaceID    null.Int        `gorm:"column:label_space_id"`
	RepoID     null.Int        `gorm:"column:label_repo_id"`
	Scope      int64           `gorm:"column:label_scope"`
	Key        string          `gorm:"column:label_key"`
	Type       enum.LabelType  `gorm:"column:label_type"`
	LabelColor enum.LabelColor `gorm:"column:label_color"`
}

type labelStore struct {
	db *gorm.DB
}

func NewLabelStore(
	db *gorm.DB,
) store.LabelStore {
	return &labelStore{
		db: db,
	}
}

var _ store.LabelStore = (*labelStore)(nil)

func (s *labelStore) Define(ctx context.Context, lbl *types.Label) error {
	result := s.db.Model(&label{}).
		Create(mapInternalLabel(lbl))
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to create label")
	}
	return nil
}

func (s *labelStore) Update(ctx context.Context, lbl *types.Label) error {
	result := s.db.Model(&label{}).
		Where("label_id = ?", lbl.ID).
		Updates(map[string]interface{}{
			"label_key":         lbl.Key,
			"label_description": lbl.Description,
			"label_type":        lbl.Type,
			"label_color":       lbl.Color,
			"label_updated":     lbl.Updated,
			"label_updated_by":  lbl.UpdatedBy,
		})
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to update label")
	}
	return nil
}

func (s *labelStore) IncrementValueCount(
	ctx context.Context,
	labelID int64,
	increment int,
) (int64, error) {
	var valueCount int64
	result := s.db.Model(&label{}).
		Where("label_id = ?", labelID).
		Update("label_value_count", gorm.Expr("label_value_count + ?", increment)).
		Scan(&valueCount)
	if result.Error != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to increment label_value_count")
	}
	return valueCount, nil
}

func (s *labelStore) Find(
	ctx context.Context,
	spaceID, repoID *int64,
	key string,
) (*types.Label, error) {
	var dst label
	if err := s.db.Model(&label{}).
		Where("(label_space_id = ? OR label_repo_id = ?) AND LOWER(label_key) = LOWER(?)", spaceID, repoID, key).
		First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find label")
	}
	return mapLabel(&dst), nil
}

func (s *labelStore) FindByID(ctx context.Context, id int64) (*types.Label, error) {
	var dst label
	if err := s.db.Model(&label{}).Where("label_id = ?", id).First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find label")
	}
	return mapLabel(&dst), nil
}

func (s *labelStore) Delete(ctx context.Context, spaceID, repoID *int64, name string) error {
	if err := s.db.Where("(label_space_id = ? OR label_repo_id = ?) AND LOWER(label_key) = LOWER(?)", spaceID, repoID, name).Delete(&label{}).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to delete label")
	}
	return nil
}

// List returns a list of pull requests for a repo or space.
func (s *labelStore) List(
	ctx context.Context,
	spaceID, repoID *int64,
	filter *types.LabelFilter,
) ([]*types.Label, error) {
	var dst []*label
	query := s.db.Model(&label{}).
		Select("label_id, " + labelColumns + ", label_value_count").
		Where(s.db.Where("label_space_id = ?", spaceID).
			Or("label_repo_id = ?", repoID)).
		Order("label_key").
		Limit(database.GormLimit(filter.Size)).
		Offset(database.GormOffset(filter.Page, filter.Size))
	if filter.Query != "" {
		query = query.Where(
			"gorm.Lower(label_key) LIKE '%' || gorm.Lower(?) || '%'",
			strings.ToLower(filter.Query),
		)
	}
	if err := query.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list labels")
	}
	return mapSliceLabel(dst), nil
}

func (s *labelStore) ListInScopes(
	ctx context.Context,
	repoID int64,
	spaceIDs []int64,
	filter *types.LabelFilter,
) ([]*types.Label, error) {
	var dst []*label
	query := s.db.Model(&label{}).
		Select("label_id, " + labelColumns + ", label_value_count").
		Where(s.db.Where("label_space_id IN ?", spaceIDs).
			Or("label_repo_id = ?", repoID)).
		Order("label_key").
		Order("label_scope").
		Limit(database.GormLimit(filter.Size)).
		Offset(database.GormOffset(filter.Page, filter.Size))
	if filter.Query != "" {
		query = query.Where(
			"gorm.Lower(label_key) LIKE '%' || gorm.Lower(?) || '%'",
			strings.ToLower(filter.Query),
		)
	}
	if err := query.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list labels in hierarchy")
	}
	return mapSliceLabel(dst), nil
}

func (s *labelStore) ListInfosInScopes(
	ctx context.Context,
	repoID int64,
	spaceIDs []int64,
	filter *types.AssignableLabelFilter,
) ([]*types.LabelInfo, error) {
	var dst []*labelInfo
	query := s.db.Model(&label{}).
		Select("label_id, label_space_id, label_repo_id, label_scope, label_key, label_type, label_color").
		Where("label_space_id IN (?) OR label_repo_id = ?", spaceIDs, repoID).
		Order("label_key").
		Order("label_scope").
		Limit(database.GormLimit(filter.Size)).
		Offset(database.GormOffset(filter.Page, filter.Size))

	if filter.Query != "" {
		query = query.Where("LOWER(label_key) LIKE '%' || gorm.Lower(?) || '%'", strings.ToLower(filter.Query))
	}

	if err := query.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list labels")
	}

	return mapLabelInfos(dst), nil
}

func (s *labelStore) CountInSpace(ctx context.Context, spaceID int64, filter *types.LabelFilter) (int64, error) {
	const sqlQuery = `SELECT COUNT(*) FROM labels WHERE label_space_id = ?`
	return s.count(ctx, sqlQuery, spaceID, filter)
}

func (s *labelStore) CountInRepo(ctx context.Context, repoID int64, filter *types.LabelFilter) (int64, error) {
	const sqlQuery = `SELECT COUNT(*) FROM labels WHERE label_repo_id = ?`
	return s.count(ctx, sqlQuery, repoID, filter)
}

func (s *labelStore) count(ctx context.Context, sqlQuery string, scopeID int64, filter *types.LabelFilter) (int64, error) {
	sqlQuery += ` AND LOWER(label_key) LIKE '%' || LOWER(?) || '%'`
	var count int64
	result := s.db.Raw(sqlQuery, scopeID, filter.Query).Scan(&count)
	if result.Error != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to count labels")
	}
	return count, nil
}

func (s *labelStore) CountInScopes(
	ctx context.Context,
	repoID int64,
	spaceIDs []int64,
	filter *types.LabelFilter,
) (int64, error) {
	var count int64
	result := s.db.Model(&label{}).
		Where(s.db.Where("label_space_id IN ?", spaceIDs).
			Or("label_repo_id = ?", repoID)).
		Where("LOWER(label_key) LIKE '%' || LOWER(?) || '%'", filter.Query).
		Count(&count)
	if result.Error != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to count labels in scopes")
	}
	return count, nil
}

func mapLabel(lbl *label) *types.Label {
	return &types.Label{
		ID:          lbl.ID,
		SpaceID:     lbl.SpaceID.Ptr(),
		RepoID:      lbl.RepoID.Ptr(),
		Scope:       lbl.Scope,
		Key:         lbl.Key,
		Type:        lbl.Type,
		Description: lbl.Description,
		ValueCount:  lbl.ValueCount,
		Color:       lbl.Color,
		Created:     lbl.Created,
		Updated:     lbl.Updated,
		CreatedBy:   lbl.CreatedBy,
		UpdatedBy:   lbl.UpdatedBy,
	}
}

func mapSliceLabel(dbLabels []*label) []*types.Label {
	result := make([]*types.Label, len(dbLabels))

	for i, lbl := range dbLabels {
		result[i] = mapLabel(lbl)
	}

	return result
}

func mapInternalLabel(lbl *types.Label) *label {
	return &label{
		ID:          lbl.ID,
		SpaceID:     null.IntFromPtr(lbl.SpaceID),
		RepoID:      null.IntFromPtr(lbl.RepoID),
		Scope:       lbl.Scope,
		Key:         lbl.Key,
		Description: lbl.Description,
		Type:        lbl.Type,
		Color:       lbl.Color,
		Created:     lbl.Created,
		Updated:     lbl.Updated,
		CreatedBy:   lbl.CreatedBy,
		UpdatedBy:   lbl.UpdatedBy,
	}
}

func mapLabelInfo(internal *labelInfo) *types.LabelInfo {
	return &types.LabelInfo{
		ID:      internal.LabelID,
		RepoID:  internal.RepoID.Ptr(),
		SpaceID: internal.SpaceID.Ptr(),
		Scope:   internal.Scope,
		Key:     internal.Key,
		Type:    internal.Type,
		Color:   internal.LabelColor,
	}
}

func mapLabelInfos(
	dbLabels []*labelInfo,
) []*types.LabelInfo {
	result := make([]*types.LabelInfo, len(dbLabels))

	for i, lbl := range dbLabels {
		result[i] = mapLabelInfo(lbl)
	}

	return result
}
