// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/pkg/errors"
)

type stageStepJoin struct {
	Stage         stage          `gorm:"embedded"`
	ID            sql.NullInt64  `gorm:"column:step_id"`
	StageID       sql.NullInt64  `gorm:"column:step_stage_id"`
	Number        sql.NullInt64  `gorm:"column:step_number"`
	ParentGroupID sql.NullInt64  `gorm:"column:step_parent_group_id"`
	Name          sql.NullString `gorm:"column:step_name"`
	Display       sql.NullString `gorm:"column:step_display"`
	Status        sql.NullString `gorm:"column:step_status"`
	Error         sql.NullString `gorm:"column:step_error"`
	ErrIgnore     sql.NullBool   `gorm:"column:step_errignore"`
	ExitCode      sql.NullInt64  `gorm:"column:step_exit_code"`
	Started       sql.NullInt64  `gorm:"column:step_started"`
	Stopped       sql.NullInt64  `gorm:"column:step_stopped"`
	Version       sql.NullInt64  `gorm:"column:step_version"`
	DependsOn     string         `gorm:"column:step_depends_on"`
	Image         sql.NullString `gorm:"column:step_image"`
	Detached      sql.NullBool   `gorm:"column:step_detached"`
	Schema        sql.NullString `gorm:"column:step_schema"`
}

// used for join operations where fields may be null.
func convertFromNullStep(join *stageStepJoin) (*types.Step, error) {
	var dependsOn []string
	err := json.Unmarshal([]byte(join.DependsOn), &dependsOn)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal step.depends_on: %w", err)
	}
	return &types.Step{
		ID:        join.ID.Int64,
		StageID:   join.StageID.Int64,
		Number:    join.Number.Int64,
		Name:      join.Name.String,
		Display:   join.Display.String,
		Status:    enum.ParseCIStatus(join.Status.String),
		Error:     join.Error.String,
		ErrIgnore: join.ErrIgnore.Bool,
		ExitCode:  int(join.ExitCode.Int64),
		Started:   join.Started.Int64,
		Stopped:   join.Stopped.Int64,
		Version:   join.Version.Int64,
		DependsOn: dependsOn,
		Image:     join.Image.String,
		Detached:  join.Detached.Bool,
		Schema:    join.Schema.String,
	}, nil
}

func mapInternalToStage(in *stage) (*types.Stage, error) {
	var dependsOn []string
	err := json.Unmarshal([]byte(in.DependsOn), &dependsOn)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal stage.depends_on")
	}
	var labels map[string]string
	err = json.Unmarshal([]byte(in.Labels), &labels)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal stage.labels")
	}
	return &types.Stage{
		ID:          in.ID,
		ExecutionID: in.ExecutionID,
		RepoID:      in.RepoID,
		Number:      in.Number,
		Name:        in.Name,
		Display:     in.Display,
		Kind:        in.Kind,
		Type:        in.Type,
		Status:      in.Status,
		Error:       in.Error,
		ErrIgnore:   in.ErrIgnore,
		ExitCode:    in.ExitCode,
		Machine:     in.Machine,
		OS:          in.OS,
		Arch:        in.Arch,
		Variant:     in.Variant,
		Kernel:      in.Kernel,
		Limit:       in.Limit,
		LimitRepo:   in.LimitRepo,
		Started:     in.Started,
		Stopped:     in.Stopped,
		Created:     in.Created,
		Updated:     in.Updated,
		Version:     in.Version,
		OnSuccess:   in.OnSuccess,
		OnFailure:   in.OnFailure,
		DependsOn:   dependsOn,
		Labels:      labels,
	}, nil
}

func mapStageToInternal(in *types.Stage) *stage {
	return &stage{
		ID:          in.ID,
		ExecutionID: in.ExecutionID,
		RepoID:      in.RepoID,
		Number:      in.Number,
		Name:        in.Name,
		Display:     in.Display,
		Kind:        in.Kind,
		Type:        in.Type,
		Status:      in.Status,
		Error:       in.Error,
		ErrIgnore:   in.ErrIgnore,
		ExitCode:    in.ExitCode,
		Machine:     in.Machine,
		OS:          in.OS,
		Arch:        in.Arch,
		Variant:     in.Variant,
		Kernel:      in.Kernel,
		Limit:       in.Limit,
		LimitRepo:   in.LimitRepo,
		Started:     in.Started,
		Stopped:     in.Stopped,
		Created:     in.Created,
		Updated:     in.Updated,
		Version:     in.Version,
		OnSuccess:   in.OnSuccess,
		OnFailure:   in.OnFailure,
		DependsOn:   database.EncodeToJSONString(in.DependsOn),
		Labels:      database.EncodeToJSONString(in.Labels),
	}
}

func mapInternalToStageList(in []*stage) ([]*types.Stage, error) {
	stages := make([]*types.Stage, len(in))
	for i, k := range in {
		s, err := mapInternalToStage(k)
		if err != nil {
			return nil, err
		}
		stages[i] = s
	}
	return stages, nil
}

// helper function scans the sql.Row and copies the column
// values to the destination object.
func scanRowsWithSteps(rows []*stageStepJoin) ([]*types.Stage, error) {
	stages := make([]*types.Stage, 0)
	var curr *types.Stage
	for _, row := range rows {
		stageObj, err := mapInternalToStage(&row.Stage)

		if err != nil {
			return nil, err
		}
		if curr == nil || curr.ID != stageObj.ID {
			curr = stageObj
			stages = append(stages, curr)
		}
		if row.ID.Valid {
			convertedStep, err := convertFromNullStep(row)
			if err != nil {
				return nil, err
			}
			curr.Steps = append(curr.Steps, convertedStep)
		}
	}
	return stages, nil
}
