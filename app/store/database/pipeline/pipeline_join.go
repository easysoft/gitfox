// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline

import (
	"database/sql"

	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

// pipelineExecutionjoin struct represents a joined row between pipelines and executions.
type pipelineExecutionJoin struct {
	Pipeline     pipeline       `gorm:"embedded"`
	ID           sql.NullInt64  `gorm:"column:execution_id"`
	PipelineID   sql.NullInt64  `gorm:"column:execution_pipeline_id"`
	Action       sql.NullString `gorm:"column:execution_action"`
	Message      sql.NullString `gorm:"column:execution_message"`
	After        sql.NullString `gorm:"column:execution_after"`
	RepoID       sql.NullInt64  `gorm:"column:execution_repo_id"`
	Trigger      sql.NullString `gorm:"column:execution_trigger"`
	Number       sql.NullInt64  `gorm:"column:execution_number"`
	Status       sql.NullString `gorm:"column:execution_status"`
	Error        sql.NullString `gorm:"column:execution_error"`
	Link         sql.NullString `gorm:"column:execution_link"`
	Timestamp    sql.NullInt64  `gorm:"column:execution_timestamp"`
	Title        sql.NullString `gorm:"column:execution_title"`
	Fork         sql.NullString `gorm:"column:execution_source_repo"`
	Source       sql.NullString `gorm:"column:execution_source"`
	Target       sql.NullString `gorm:"column:execution_target"`
	Author       sql.NullString `gorm:"column:execution_author"`
	AuthorName   sql.NullString `gorm:"column:execution_author_name"`
	AuthorEmail  sql.NullString `gorm:"column:execution_author_email"`
	AuthorAvatar sql.NullString `gorm:"column:execution_author_avatar"`
	Started      sql.NullInt64  `gorm:"column:execution_started"`
	Finished     sql.NullInt64  `gorm:"column:execution_finished"`
	Created      sql.NullInt64  `gorm:"column:execution_created"`
	Updated      sql.NullInt64  `gorm:"column:execution_updated"`
}

func convert(rows []*pipelineExecutionJoin) []*types.Pipeline {
	pipelines := []*types.Pipeline{}
	for _, k := range rows {
		pipeline := convertPipelineJoin(k)
		pipelines = append(pipelines, pipeline)
	}
	return pipelines
}

func convertPipelineJoin(join *pipelineExecutionJoin) *types.Pipeline {
	ret := mapInternalToPipeline(&join.Pipeline)
	if !join.ID.Valid {
		return ret
	}
	ret.Execution = &types.Execution{
		ID:           join.ID.Int64,
		PipelineID:   join.PipelineID.Int64,
		RepoID:       join.RepoID.Int64,
		Action:       enum.TriggerAction(join.Action.String),
		Trigger:      join.Trigger.String,
		Number:       join.Number.Int64,
		After:        join.After.String,
		Message:      join.Message.String,
		Status:       enum.ParseCIStatus(join.Status.String),
		Error:        join.Error.String,
		Link:         join.Link.String,
		Timestamp:    join.Timestamp.Int64,
		Title:        join.Title.String,
		Fork:         join.Fork.String,
		Source:       join.Source.String,
		Target:       join.Target.String,
		Author:       join.Author.String,
		AuthorName:   join.AuthorName.String,
		AuthorEmail:  join.AuthorEmail.String,
		AuthorAvatar: join.AuthorAvatar.String,
		Started:      join.Started.Int64,
		Finished:     join.Finished.Int64,
		Created:      join.Created.Int64,
		Updated:      join.Updated.Int64,
	}
	return ret
}

type pipelineRepoJoin struct {
	*types.Pipeline
	RepoID  sql.NullInt64  `db:"repo_id"`
	RepoUID sql.NullString `db:"repo_uid"`
}

func convertPipelineRepoJoins(rows []*pipelineRepoJoin) []*types.Pipeline {
	pipelines := []*types.Pipeline{}
	for _, k := range rows {
		pipeline := convertPipelineRepoJoin(k)
		pipelines = append(pipelines, pipeline)
	}
	return pipelines
}

func convertPipelineRepoJoin(join *pipelineRepoJoin) *types.Pipeline {
	ret := join.Pipeline
	if !join.RepoID.Valid {
		return ret
	}
	ret.RepoUID = join.RepoUID.String
	return ret
}
