// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.StageStore = (*StageStore)(nil)

const (
	tableStage = "stages"
)

type stage struct {
	ID            int64         `gorm:"column:stage_id;primaryKey"`
	ExecutionID   int64         `gorm:"column:stage_execution_id"`
	RepoID        int64         `gorm:"column:stage_repo_id"`
	Number        int64         `gorm:"column:stage_number"`
	Name          string        `gorm:"column:stage_name"`
	Display       string        `gorm:"column:stage_display"`
	Kind          string        `gorm:"column:stage_kind"`
	Type          string        `gorm:"column:stage_type"`
	Status        enum.CIStatus `gorm:"column:stage_status"`
	Error         string        `gorm:"column:stage_error"`
	ParentGroupID int64         `gorm:"column:stage_parent_group_id"`
	ErrIgnore     bool          `gorm:"column:stage_errignore"`
	ExitCode      int           `gorm:"column:stage_exit_code"`
	Machine       string        `gorm:"column:stage_machine"`
	OS            string        `gorm:"column:stage_os"`
	Arch          string        `gorm:"column:stage_arch"`
	Variant       string        `gorm:"column:stage_variant"`
	Kernel        string        `gorm:"column:stage_kernel"`
	Limit         int           `gorm:"column:stage_limit"`
	LimitRepo     int           `gorm:"column:stage_limit_repo"`
	Started       int64         `gorm:"column:stage_started"`
	Stopped       int64         `gorm:"column:stage_stopped"`
	Created       int64         `gorm:"column:stage_created"`
	Updated       int64         `gorm:"column:stage_updated"`
	Version       int64         `gorm:"column:stage_version"`
	OnSuccess     bool          `gorm:"column:stage_on_success"`
	OnFailure     bool          `gorm:"column:stage_on_failure"`
	DependsOn     string        `gorm:"column:stage_depends_on"`
	Labels        string        `gorm:"column:stage_labels"`
}

// NewStageOrmStore returns a new StageStore.
func NewStageOrmStore(db *gorm.DB) *StageStore {
	return &StageStore{
		db: db,
	}
}

type StageStore struct {
	db *gorm.DB
}

// FindByNumber returns a stage given an execution ID and a stage number.
func (s *StageStore) FindByNumber(ctx context.Context, executionID int64, stageNum int) (*types.Stage, error) {
	q := stage{ExecutionID: executionID, Number: int64(stageNum)}

	dst := new(stage)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStage).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find stage")
	}
	return mapInternalToStage(dst)
}

// Create adds a stage in the database.
func (s *StageStore) Create(ctx context.Context, st *types.Stage) error {
	dbObj := mapStageToInternal(st)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStage).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Stage query failed")
	}
	st.ID = dbObj.ID
	return nil
}

// ListWithSteps returns a stage with information about all its containing steps.
func (s *StageStore) ListWithSteps(ctx context.Context, executionID int64) ([]*types.Stage, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStage).
		Joins(`LEFT JOIN steps on stages.stage_id=steps.step_stage_id`).
		Where(&stage{ExecutionID: executionID}).Order("stage_id ASC, step_id ASC")

	dst := make([]*stageStepJoin, 0)

	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to query stages and steps")
	}
	return scanRowsWithSteps(dst)
}

// Find returns a stage given the stage ID.
func (s *StageStore) Find(ctx context.Context, stageID int64) (*types.Stage, error) {
	dst := new(stage)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStage).First(dst, stageID).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find stage")
	}
	return mapInternalToStage(dst)
}

// ListIncomplete returns a list of stages with a pending status.
func (s *StageStore) ListIncomplete(ctx context.Context) ([]*types.Stage, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStage).
		Where("stage_status IN ('pending','running')").
		Order("stage_id ASC")

	dst := []*stage{}
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find incomplete stages")
	}
	// map stages list
	return mapInternalToStageList(dst)
}

// List returns a list of stages corresponding to an execution ID.
func (s *StageStore) List(ctx context.Context, executionID int64) ([]*types.Stage, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStage).
		Where(&stage{ExecutionID: executionID}).
		Order("stage_number ASC")

	dst := []*stage{}
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find stages")
	}
	// map stages list
	return mapInternalToStageList(dst)
}

// Update tries to update a stage in the datastore and returns a locking error
// if it was unable to do so.
func (s *StageStore) Update(ctx context.Context, st *types.Stage) error {
	updatedAt := time.Now()
	steps := st.Steps

	dbStage := mapStageToInternal(st)

	dbStage.Version++
	dbStage.Updated = updatedAt.UnixMilli()

	updateFields := []string{"Status", "Machine",
		"Started", "Stopped", "ExitCode", "Updated", "Version", "Error",
		"OnSuccess", "OnFailure", "ErrIgnore", "DependsOn", "Labels",
	}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStage).
		Where(&stage{ID: st.ID, Version: dbStage.Version - 1}).
		Select(updateFields).Updates(dbStage)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update stage")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrVersionConflict
	}

	m, err := mapInternalToStage(dbStage)
	if err != nil {
		return fmt.Errorf("could not map stage object: %w", err)
	}
	*st = *m
	st.Version = dbStage.Version
	st.Updated = dbStage.Updated
	st.Steps = steps // steps is not mapped in database.
	return nil
}
