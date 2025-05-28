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

	"github.com/gotidy/ptr"
	"github.com/guregu/null"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ store.PullReqLabelAssignmentStore = (*pullReqLabelStore)(nil)

func NewPullReqLabelStore(db *gorm.DB) store.PullReqLabelAssignmentStore {
	return &pullReqLabelStore{
		db: db,
	}
}

type pullReqLabelStore struct {
	db *gorm.DB
}

type pullReqLabel struct {
	PullReqID    int64    `gorm:"column:pullreq_label_pullreq_id"`
	LabelID      int64    `gorm:"column:pullreq_label_label_id"`
	LabelValueID null.Int `gorm:"column:pullreq_label_label_value_id"`
	Created      int64    `gorm:"column:pullreq_label_created"`
	Updated      int64    `gorm:"column:pullreq_label_updated"`
	CreatedBy    int64    `gorm:"column:pullreq_label_created_by"`
	UpdatedBy    int64    `gorm:"column:pullreq_label_updated_by"`
}

type pullReqAssignmentInfo struct {
	PullReqID  int64           `gorm:"column:pullreq_label_pullreq_id"`
	LabelID    int64           `gorm:"column:label_id"`
	LabelKey   string          `gorm:"column:label_key"`
	LabelColor enum.LabelColor `gorm:"column:label_color"`
	LabelScope int64           `gorm:"column:label_scope"`
	ValueCount int64           `gorm:"column:label_value_count"`
	ValueID    null.Int        `gorm:"column:label_value_id"`
	Value      null.String     `gorm:"column:label_value_value"`
	ValueColor null.String     `gorm:"column:label_value_color"` // get's converted to *enum.LabelColor
}

const (
	pullReqLabelColumns = `
		 pullreq_label_pullreq_id
		,pullreq_label_label_id
		,pullreq_label_label_value_id
		,pullreq_label_created
		,pullreq_label_updated
		,pullreq_label_created_by
		,pullreq_label_updated_by`
)

func (s *pullReqLabelStore) Assign(ctx context.Context, label *types.PullReqLabel) error {
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(mapInternalPullReqLabel(label))
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "failed to create or update pull request label")
	}
	label.Created = result.Statement.ReflectValue.FieldByName("Created").Interface().(int64)
	label.CreatedBy = result.Statement.ReflectValue.FieldByName("CreatedBy").Interface().(int64)
	return nil
}

func (s *pullReqLabelStore) Unassign(ctx context.Context, pullreqID int64, labelID int64) error {
	if err := s.db.WithContext(ctx).Delete(&pullReqLabel{}, "pullreq_label_pullreq_id = ? AND pullreq_label_label_id = ?", pullreqID, labelID).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "failed to delete pullreq label")
	}
	return nil
}

func (s *pullReqLabelStore) FindByLabelID(
	ctx context.Context,
	pullreqID int64,
	labelID int64,
) (*types.PullReqLabel, error) {
	var dst pullReqLabel
	if err := s.db.WithContext(ctx).Table("pullreq_labels").
		Select(pullReqLabelColumns).
		Where("pullreq_label_pullreq_id = ? AND pullreq_label_label_id = ?", pullreqID, labelID).
		First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to find pull request label")
	}

	return mapPullReqLabel(&dst), nil
}

func (s *pullReqLabelStore) ListAssigned(
	ctx context.Context,
	pullreqID int64,
) (map[int64]*types.LabelAssignment, error) {
	var dst []*struct {
		labelInfo
		labelValueInfo
	}
	if err := s.db.WithContext(ctx).Table("pullreq_labels").
		Select(`
			labels.label_id
			,labels.label_repo_id
			,labels.label_space_id
			,labels.label_key
			,label_values.label_value_id
			,label_values.label_value_label_id
			,label_values.label_value_value
			,labels.label_color
			,label_values.label_value_color
			,labels.label_scope
			,labels.label_type
		`).
		Joins("INNER JOIN labels ON pullreq_labels.pullreq_label_label_id = labels.label_id").
		Joins("LEFT JOIN label_values ON pullreq_labels.pullreq_label_label_value_id = label_values.label_value_id").
		Where("pullreq_labels.pullreq_label_pullreq_id = ?", pullreqID).
		Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to list assigned label")
	}

	ret := make(map[int64]*types.LabelAssignment, len(dst))
	for _, res := range dst {
		li := mapLabelInfo(&res.labelInfo)
		lvi := mapLabeValuelInfo(&res.labelValueInfo)
		ret[li.ID] = &types.LabelAssignment{
			LabelInfo:     *li,
			AssignedValue: lvi,
		}
	}

	return ret, nil
}

func (s *pullReqLabelStore) ListAssignedByPullreqIDs(
	ctx context.Context,
	pullreqIDs []int64,
) (map[int64][]*types.LabelPullReqAssignmentInfo, error) {
	var dst []*pullReqAssignmentInfo
	if err := s.db.WithContext(ctx).Table("pullreq_labels").
		Select(`
			pullreq_label_pullreq_id
			,label_id
			,label_key
			,label_color
			,label_scope
			,label_value_count
			,label_value_id
			,label_value_value
			,label_value_color
		`).
		Joins("INNER JOIN labels ON pullreq_label_label_id = label_id").
		Joins("LEFT JOIN label_values ON pullreq_label_label_value_id = label_value_id").
		Where("pullreq_label_pullreq_id IN ?", pullreqIDs).
		Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to list assigned label")
	}

	return mapPullReqAssignmentInfos(dst), nil
}

func (s *pullReqLabelStore) FindValueByLabelID(
	ctx context.Context,
	labelID int64,
) (*types.LabelValue, error) {
	var dst labelValue
	if err := s.db.WithContext(ctx).Table("pullreq_labels").
		Select("label_values.label_value_id, "+labelValueColumns).
		Joins("JOIN label_values ON pullreq_labels.pullreq_label_label_value_id = label_values.label_value_id").
		Where("pullreq_labels.pullreq_label_label_id = ?", labelID).
		First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find label")
	}

	return mapLabelValue(&dst), nil
}

func mapInternalPullReqLabel(lbl *types.PullReqLabel) *pullReqLabel {
	return &pullReqLabel{
		PullReqID:    lbl.PullReqID,
		LabelID:      lbl.LabelID,
		LabelValueID: null.IntFromPtr(lbl.ValueID),
		Created:      lbl.Created,
		Updated:      lbl.Updated,
		CreatedBy:    lbl.CreatedBy,
		UpdatedBy:    lbl.UpdatedBy,
	}
}

func mapPullReqLabel(lbl *pullReqLabel) *types.PullReqLabel {
	return &types.PullReqLabel{
		PullReqID: lbl.PullReqID,
		LabelID:   lbl.LabelID,
		ValueID:   lbl.LabelValueID.Ptr(),
		Created:   lbl.Created,
		Updated:   lbl.Updated,
		CreatedBy: lbl.CreatedBy,
		UpdatedBy: lbl.UpdatedBy,
	}
}

func mapPullReqAssignmentInfo(lbl *pullReqAssignmentInfo) *types.LabelPullReqAssignmentInfo {
	var valueColor *enum.LabelColor
	if lbl.ValueColor.Valid {
		valueColor = ptr.Of(enum.LabelColor(lbl.ValueColor.String))
	}
	return &types.LabelPullReqAssignmentInfo{
		PullReqID:  lbl.PullReqID,
		LabelID:    lbl.LabelID,
		LabelKey:   lbl.LabelKey,
		LabelColor: lbl.LabelColor,
		LabelScope: lbl.LabelScope,
		ValueCount: lbl.ValueCount,
		ValueID:    lbl.ValueID.Ptr(),
		Value:      lbl.Value.Ptr(),
		ValueColor: valueColor,
	}
}

func mapPullReqAssignmentInfos(
	dbLabels []*pullReqAssignmentInfo,
) map[int64][]*types.LabelPullReqAssignmentInfo {
	result := make(map[int64][]*types.LabelPullReqAssignmentInfo)

	for _, lbl := range dbLabels {
		result[lbl.PullReqID] = append(result[lbl.PullReqID], mapPullReqAssignmentInfo(lbl))
	}

	return result
}
