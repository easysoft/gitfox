// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package labels

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

const (
	labelValueColumns = `
		 label_value_label_id
		,label_value_value
		,label_value_color
		,label_value_created
		,label_value_updated
		,label_value_created_by
		,label_value_updated_by`

	labelValueSelectBase = `SELECT label_value_id, ` + labelValueColumns + ` FROM label_values`
)

type labelValue struct {
	ID        int64           `gorm:"column:label_value_id"`
	LabelID   int64           `gorm:"column:label_value_label_id"`
	Value     string          `gorm:"column:label_value_value"`
	Color     enum.LabelColor `gorm:"column:label_value_color"`
	Created   int64           `gorm:"column:label_value_created"`
	Updated   int64           `gorm:"column:label_value_updated"`
	CreatedBy int64           `gorm:"column:label_value_created_by"`
	UpdatedBy int64           `gorm:"column:label_value_updated_by"`
}

type labelValueInfo struct {
	ValueID    null.Int    `gorm:"column:label_value_id"`
	LabelID    null.Int    `gorm:"column:label_value_label_id"`
	Value      null.String `gorm:"column:label_value_value"`
	ValueColor null.String `gorm:"column:label_value_color"`
}

type labelValueStore struct {
	db *gorm.DB
}

func NewLabelValueStore(
	db *gorm.DB,
) store.LabelValueStore {
	return &labelValueStore{
		db: db,
	}
}

var _ store.LabelValueStore = (*labelValueStore)(nil)

func (s *labelValueStore) Define(ctx context.Context, lblVal *types.LabelValue) error {
	result := s.db.Model(&labelValue{}).
		Create(lblVal)
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to create label value")
	}
	return nil
}

func (s *labelValueStore) Update(ctx context.Context, lblVal *types.LabelValue) error {
	result := s.db.Model(&labelValue{}).
		Where("label_value_id = ?", lblVal.ID).
		Updates(map[string]interface{}{
			"label_value_value":      lblVal.Value,
			"label_value_color":      lblVal.Color,
			"label_value_updated":    lblVal.Updated,
			"label_value_updated_by": lblVal.UpdatedBy,
		})
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to update label value")
	}
	return nil
}

func (s *labelValueStore) Delete(
	ctx context.Context,
	labelID int64,
	value string,
) error {
	if err := s.db.Where("label_value_label_id = ? AND LOWER(label_value_value) = LOWER(?)", labelID, value).Delete(&labelValue{}).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to delete label")
	}
	return nil
}

func (s *labelValueStore) DeleteMany(
	ctx context.Context,
	labelID int64,
	values []string,
) error {
	if err := s.db.
		Where("label_value_label_id = ?", labelID).
		Where("label_value_value IN ?", values).
		Delete(&labelValue{}).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to delete label")
	}

	return nil
}

// List returns a list of label values for a specified label.
func (s *labelValueStore) List(
	ctx context.Context,
	labelID int64,
	opts *types.ListQueryFilter,
) ([]*types.LabelValue, error) {
	var dst []*labelValue
	if err := s.db.
		Select("label_value_id, "+labelValueColumns).
		Table("label_values").
		Where("label_value_label_id = ?", labelID).
		Limit(database.GormLimit(opts.Size)).
		Offset(database.GormOffset(opts.Page, opts.Size)).
		Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list labels")
	}

	return mapSliceLabelValue(dst), nil
}

func (s *labelValueStore) ListInfosByLabelIDs(
	ctx context.Context,
	labelIDs []int64,
) (map[int64][]*types.LabelValueInfo, error) {
	var dst []*labelValueInfo
	if err := s.db.
		Select(`
			label_value_id,
			label_value_label_id,
			label_value_value,
			label_value_color
		`).
		Table("label_values").
		Where("label_value_label_id IN ?", labelIDs).
		Order("label_value_value").
		Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list labels")
	}

	valueInfos := mapLabelValuInfos(dst)
	labelValueMap := make(map[int64][]*types.LabelValueInfo)
	for _, info := range valueInfos {
		labelValueMap[*info.LabelID] = append(labelValueMap[*info.LabelID], info)
	}

	return labelValueMap, nil
}

func (s *labelValueStore) FindByLabelID(
	ctx context.Context,
	labelID int64,
	value string,
) (*types.LabelValue, error) {
	var dst labelValue
	if err := s.db.Where("label_value_label_id = ? AND LOWER(label_value_value) = LOWER(?)", labelID, value).First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find label")
	}

	return mapLabelValue(&dst), nil
}

func (s *labelValueStore) FindByID(ctx context.Context, id int64) (*types.LabelValue, error) {
	var dst labelValue
	if err := s.db.Where("label_value_id = ?", id).First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find label")
	}

	return mapLabelValue(&dst), nil
}

func mapLabelValue(lbl *labelValue) *types.LabelValue {
	return &types.LabelValue{
		ID:        lbl.ID,
		LabelID:   lbl.LabelID,
		Value:     lbl.Value,
		Color:     lbl.Color,
		Created:   lbl.Created,
		Updated:   lbl.Updated,
		CreatedBy: lbl.CreatedBy,
		UpdatedBy: lbl.UpdatedBy,
	}
}

func mapSliceLabelValue(dbLabelValues []*labelValue) []*types.LabelValue {
	result := make([]*types.LabelValue, len(dbLabelValues))

	for i, lbl := range dbLabelValues {
		result[i] = mapLabelValue(lbl)
	}

	return result
}

func mapInternalLabelValue(lblVal *types.LabelValue) *labelValue {
	return &labelValue{
		ID:        lblVal.ID,
		LabelID:   lblVal.LabelID,
		Value:     lblVal.Value,
		Color:     lblVal.Color,
		Created:   lblVal.Created,
		Updated:   lblVal.Updated,
		CreatedBy: lblVal.CreatedBy,
		UpdatedBy: lblVal.UpdatedBy,
	}
}

func mapLabeValuelInfo(internal *labelValueInfo) *types.LabelValueInfo {
	if !internal.ValueID.Valid {
		return nil
	}
	return &types.LabelValueInfo{
		ID:      internal.ValueID.Ptr(),
		LabelID: internal.LabelID.Ptr(),
		Value:   internal.Value.Ptr(),
		Color:   internal.ValueColor.Ptr(),
	}
}

func mapLabelValuInfos(
	dbLabels []*labelValueInfo,
) []*types.LabelValueInfo {
	result := make([]*types.LabelValueInfo, len(dbLabels))

	for i, lbl := range dbLabels {
		result[i] = mapLabeValuelInfo(lbl)
	}

	return result
}
