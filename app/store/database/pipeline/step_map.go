// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline

import (
	"encoding/json"
	"fmt"

	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
)

func mapInternalToStep(in *step) (*types.Step, error) {
	var dependsOn []string
	err := json.Unmarshal([]byte(in.DependsOn), &dependsOn)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal step.DependsOn: %w", err)
	}
	return &types.Step{
		ID:        in.ID,
		StageID:   in.StageID,
		Number:    in.Number,
		Name:      in.Name,
		Display:   in.Display,
		Status:    in.Status,
		Error:     in.Error,
		ErrIgnore: in.ErrIgnore,
		ExitCode:  in.ExitCode,
		Started:   in.Started,
		Stopped:   in.Stopped,
		Version:   in.Version,
		DependsOn: dependsOn,
		Image:     in.Image,
		Detached:  in.Detached,
		Schema:    in.Schema,
	}, nil
}

func mapStepToInternal(in *types.Step) *step {
	return &step{
		ID:        in.ID,
		StageID:   in.StageID,
		Number:    in.Number,
		Name:      in.Name,
		Display:   in.Display,
		Status:    in.Status,
		Error:     in.Error,
		ErrIgnore: in.ErrIgnore,
		ExitCode:  in.ExitCode,
		Started:   in.Started,
		Stopped:   in.Stopped,
		Version:   in.Version,
		DependsOn: database.EncodeToJSONString(in.DependsOn),
		Image:     in.Image,
		Detached:  in.Detached,
		Schema:    in.Schema,
	}
}
