// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.StepStore = (*StepStore)(nil)

const (
	tableStep = "steps"
)

type step struct {
	ID            int64         `gorm:"column:step_id;primaryKey"`
	StageID       int64         `gorm:"column:step_stage_id"`
	Number        int64         `gorm:"column:step_number"`
	ParentGroupID int64         `gorm:"column:step_parent_group_id"`
	Name          string        `gorm:"column:step_name"`
	Display       string        `gorm:"column:step_display"`
	Status        enum.CIStatus `gorm:"column:step_status"`
	Error         string        `gorm:"column:step_error"`
	ErrIgnore     bool          `gorm:"column:step_errignore"`
	ExitCode      int           `gorm:"column:step_exit_code"`
	Started       int64         `gorm:"column:step_started"`
	Stopped       int64         `gorm:"column:step_stopped"`
	Version       int64         `gorm:"column:step_version"`
	DependsOn     string        `gorm:"column:step_depends_on"`
	Image         string        `gorm:"column:step_image"`
	Detached      bool          `gorm:"column:step_detached"`
	Schema        string        `gorm:"column:step_schema"`
}

// NewStepOrmStore returns a new StepStore.
func NewStepOrmStore(db *gorm.DB) *StepStore {
	return &StepStore{
		db: db,
	}
}

type StepStore struct {
	db *gorm.DB
}

// FindByNumber returns a step given a stage ID and a step number.
func (s *StepStore) FindByNumber(ctx context.Context, stageID int64, stepNum int) (*types.Step, error) {
	q := step{StageID: stageID, Number: int64(stepNum)}
	dst := new(step)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStep).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find step")
	}
	return mapInternalToStep(dst)
}

// Create creates a step.
func (s *StepStore) Create(ctx context.Context, step *types.Step) error {
	dbObj := mapStepToInternal(step)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStep).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Step query failed")
	}

	step.ID = dbObj.ID
	return nil
}

// Update tries to update a step in the datastore and returns a locking error
// if it was unable to do so.
func (s *StepStore) Update(ctx context.Context, e *types.Step) error {
	dbStep := mapStepToInternal(e)

	dbStep.Version++

	updateFields := []string{"Name", "Status", "Error", "ErrIgnore", "ExitCode",
		"Started", "Stopped", "DependsOn", "Image", "Detached", "Schema", "Version",
	}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableStep).
		Where(&step{ID: e.ID, Version: dbStep.Version - 1}).
		Select(updateFields).Updates(dbStep)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update step")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrVersionConflict
	}

	m, err := mapInternalToStep(dbStep)
	if err != nil {
		return fmt.Errorf("could not map step object: %w", err)
	}
	*e = *m
	e.Version = dbStep.Version
	return nil
}
