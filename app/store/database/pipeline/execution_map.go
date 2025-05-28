// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline

import (
	"encoding/json"

	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
)

func mapInternalToExecution(in *execution) (*types.Execution, error) {
	var params map[string]string
	err := json.Unmarshal([]byte(in.Params), &params)
	if err != nil {
		return nil, err
	}
	return &types.Execution{
		ID:           in.ID,
		PipelineID:   in.PipelineID,
		CreatedBy:    in.CreatedBy,
		RepoID:       in.RepoID,
		Trigger:      in.Trigger,
		Number:       in.Number,
		Parent:       in.Parent,
		Status:       in.Status,
		Error:        in.Error,
		Event:        in.Event,
		Action:       in.Action,
		Link:         in.Link,
		Timestamp:    in.Timestamp,
		Title:        in.Title,
		Message:      in.Message,
		Before:       in.Before,
		After:        in.After,
		Ref:          in.Ref,
		Fork:         in.Fork,
		Source:       in.Source,
		Target:       in.Target,
		Author:       in.Author,
		AuthorName:   in.AuthorName,
		AuthorEmail:  in.AuthorEmail,
		AuthorAvatar: in.AuthorAvatar,
		Sender:       in.Sender,
		Params:       params,
		Cron:         in.Cron,
		Deploy:       in.Deploy,
		DeployID:     in.DeployID,
		Debug:        in.Debug,
		Started:      in.Started,
		Finished:     in.Finished,
		Created:      in.Created,
		Updated:      in.Updated,
		Version:      in.Version,
	}, nil
}

func mapExecutionToInternal(in *types.Execution) *execution {
	return &execution{
		ID:           in.ID,
		PipelineID:   in.PipelineID,
		CreatedBy:    in.CreatedBy,
		RepoID:       in.RepoID,
		Trigger:      in.Trigger,
		Number:       in.Number,
		Parent:       in.Parent,
		Status:       in.Status,
		Error:        in.Error,
		Event:        in.Event,
		Action:       in.Action,
		Link:         in.Link,
		Timestamp:    in.Timestamp,
		Title:        in.Title,
		Message:      in.Message,
		Before:       in.Before,
		After:        in.After,
		Ref:          in.Ref,
		Fork:         in.Fork,
		Source:       in.Source,
		Target:       in.Target,
		Author:       in.Author,
		AuthorName:   in.AuthorName,
		AuthorEmail:  in.AuthorEmail,
		AuthorAvatar: in.AuthorAvatar,
		Sender:       in.Sender,
		Params:       database.EncodeToJSONString(in.Params),
		Cron:         in.Cron,
		Deploy:       in.Deploy,
		DeployID:     in.DeployID,
		Debug:        in.Debug,
		Started:      in.Started,
		Finished:     in.Finished,
		Created:      in.Created,
		Updated:      in.Updated,
		Version:      in.Version,
	}
}

func mapInternalToExecutionList(in []*execution) ([]*types.Execution, error) {
	executions := make([]*types.Execution, len(in))
	for i, k := range in {
		e, err := mapInternalToExecution(k)
		if err != nil {
			return nil, err
		}
		executions[i] = e
	}
	return executions, nil
}
