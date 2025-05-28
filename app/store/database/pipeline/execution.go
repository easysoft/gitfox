// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/pkg/util/common"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.ExecutionStore = (*ExecutionStore)(nil)

// NewExecutionOrmStore returns a new ExecutionStore.
func NewExecutionOrmStore(db *gorm.DB) *ExecutionStore {
	return &ExecutionStore{
		db: db,
	}
}

type ExecutionStore struct {
	db *gorm.DB
}

// execution represents an execution object stored in the database.
type execution struct {
	ID           int64              `gorm:"column:execution_id;primaryKey"`
	PipelineID   int64              `gorm:"column:execution_pipeline_id"`
	CreatedBy    int64              `gorm:"column:execution_created_by"`
	RepoID       int64              `gorm:"column:execution_repo_id"`
	Trigger      string             `gorm:"column:execution_trigger"`
	Number       int64              `gorm:"column:execution_number"`
	Parent       int64              `gorm:"column:execution_parent"`
	Status       enum.CIStatus      `gorm:"column:execution_status"`
	Error        string             `gorm:"column:execution_error"`
	Event        enum.TriggerEvent  `gorm:"column:execution_event"`
	Action       enum.TriggerAction `gorm:"column:execution_action"`
	Link         string             `gorm:"column:execution_link"`
	Timestamp    int64              `gorm:"column:execution_timestamp"`
	Title        string             `gorm:"column:execution_title"`
	Message      string             `gorm:"column:execution_message"`
	Before       string             `gorm:"column:execution_before"`
	After        string             `gorm:"column:execution_after"`
	Ref          string             `gorm:"column:execution_ref"`
	Fork         string             `gorm:"column:execution_source_repo"`
	Source       string             `gorm:"column:execution_source"`
	Target       string             `gorm:"column:execution_target"`
	Author       string             `gorm:"column:execution_author"`
	AuthorName   string             `gorm:"column:execution_author_name"`
	AuthorEmail  string             `gorm:"column:execution_author_email"`
	AuthorAvatar string             `gorm:"column:execution_author_avatar"`
	Sender       string             `gorm:"column:execution_sender"`
	Params       string             `gorm:"column:execution_params"`
	Cron         string             `gorm:"column:execution_cron"`
	Deploy       string             `gorm:"column:execution_deploy"`
	DeployID     int64              `gorm:"column:execution_deploy_id"`
	Debug        bool               `gorm:"column:execution_debug"`
	Started      int64              `gorm:"column:execution_started"`
	Finished     int64              `gorm:"column:execution_finished"`
	Created      int64              `gorm:"column:execution_created"`
	Updated      int64              `gorm:"column:execution_updated"`
	Version      int64              `gorm:"column:execution_version"`
}

type executionPipelineRepoJoin struct {
	execution
	PipelineUID sql.NullString `db:"pipeline_uid" gorm:"column:pipeline_uid"`
	RepoUID     sql.NullString `db:"repo_uid" gorm:"column:repo_uid"`
}

const (
	tableExecution = "executions"
)

// Find returns an execution given an execution ID.
func (s *ExecutionStore) Find(ctx context.Context, id int64) (*types.Execution, error) {
	dst := new(execution)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableExecution).First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find execution")
	}
	return mapInternalToExecution(dst)
}

// FindByNumber returns an execution given a pipeline ID and an execution number.
func (s *ExecutionStore) FindByNumber(
	ctx context.Context,
	pipelineID int64,
	executionNum int64,
) (*types.Execution, error) {
	dst := new(execution)
	q := execution{PipelineID: pipelineID, Number: executionNum}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableExecution).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find execution")
	}
	return mapInternalToExecution(dst)
}

// Create creates a new execution in the datastore.
func (s *ExecutionStore) Create(ctx context.Context, execution *types.Execution) error {
	dbObj := mapExecutionToInternal(execution)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableExecution).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Execution query failed")
	}

	execution.ID = dbObj.ID
	return nil
}

// Update tries to update an execution in the datastore with optimistic locking.
func (s *ExecutionStore) Update(ctx context.Context, e *types.Execution) error {
	updatedAt := time.Now()
	stages := e.Stages

	dbExec := mapExecutionToInternal(e)

	dbExec.Version++
	dbExec.Updated = updatedAt.UnixMilli()

	updateFields := []string{"Status", "Error", "Event", "Started", "Finished", "Updated", "Version"}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableExecution).
		Where(&execution{ID: e.ID, Version: dbExec.Version - 1}).
		Select(updateFields).Updates(dbExec)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update execution")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrVersionConflict
	}

	m, err := mapInternalToExecution(dbExec)
	if err != nil {
		return fmt.Errorf("could not map execution object: %w", err)
	}
	*e = *m
	e.Version = dbExec.Version
	e.Updated = dbExec.Updated
	e.Stages = stages // stages are not mapped in database.
	return nil
}

// List lists the executions for a given pipeline ID.
// It orders them in descending order of execution number.
func (s *ExecutionStore) List(
	ctx context.Context,
	pipelineID int64,
	pagination types.Pagination,
	createdAfter int64,
) ([]*types.Execution, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableExecution).Where(&execution{PipelineID: pipelineID}).
		Order("execution_number " + enum.OrderDesc.String())

	stmt = stmt.Limit(int(database.Limit(pagination.Size)))
	stmt = stmt.Offset(int(database.Offset(pagination.Page, pagination.Size)))

	if createdAfter >= common.DefaultPipelineExecutionCreatedUnix {
		stmt = stmt.Where("execution_created >= ?", createdAfter)
	}

	dst := []*execution{}
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return mapInternalToExecutionList(dst)
}

// ListInSpace lists the executions in a given space.
// It orders them in descending order of execution id.
func (s *ExecutionStore) ListInSpace(
	ctx context.Context,
	spaceID int64,
	filter types.ListExecutionsFilter,
) ([]*types.Execution, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableExecution).
		Select("executions.*, pipelines.pipeline_uid, repositories.repo_uid").
		Joins("INNER JOIN pipelines ON execution_pipeline_id = pipeline_id").
		Joins("INNER JOIN repositories ON execution_repo_id = repo_id").
		Where("repo_parent_id = ?", spaceID).
		Order("execution_" + string(filter.Sort) + " " + filter.Order.String())

	stmt = stmt.Limit(database.GormLimit(filter.Size))
	stmt = stmt.Offset(database.GormOffset(filter.Page, filter.Size))

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(pipeline_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	if filter.PipelineIdentifier != "" {
		stmt = stmt.Where("pipeline_uid = ?", filter.PipelineIdentifier)
	}

	dst := []*executionPipelineRepoJoin{}
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return convertExecutionPipelineRepoJoins(dst)
}

func (s *ExecutionStore) ListByPipelineIDs(
	ctx context.Context,
	pipelineIDs []int64,
	maxRows int64,
) (map[int64][]*types.ExecutionInfo, error) {
	stmt := s.db.Table(tableExecution).
		Select("execution_number, execution_pipeline_id, execution_status").
		Where("execution_pipeline_id IN ?", pipelineIDs).
		Order("execution_number DESC").
		Limit(int(maxRows))

	var dst []*types.ExecutionInfo
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list executions by pipeline IDs")
	}

	executionInfosMap := make(map[int64][]*types.ExecutionInfo)
	for _, info := range dst {
		executionInfosMap[info.PipelineID] = append(
			executionInfosMap[info.PipelineID],
			info,
		)
	}

	return executionInfosMap, nil
}

// Count of executions in a pipeline, if pipelineID is 0 then return total number of executions.
func (s *ExecutionStore) Count(ctx context.Context, pipelineID, createdAfter int64) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableExecution).Where(&execution{PipelineID: pipelineID})
	if createdAfter >= common.DefaultPipelineExecutionCreatedUnix {
		stmt = stmt.Where("execution_created >= ?", createdAfter)
	}
	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}

// CountInSpace counts the number of executions in a given space.
func (s *ExecutionStore) CountInSpace(
	ctx context.Context,
	spaceID int64,
	filter types.ListExecutionsFilter,
) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableExecution).
		Select("count(*)").
		Joins("INNER JOIN pipelines ON execution_pipeline_id = pipeline_id").
		Joins("INNER JOIN repositories ON execution_repo_id = repo_id").
		Where("repo_parent_id = ?", spaceID)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(pipeline_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	if filter.PipelineIdentifier != "" {
		stmt = stmt.Where("pipeline_uid = ?", filter.PipelineIdentifier)
	}

	var count int64
	if err := stmt.Count(&count).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}

// Delete deletes an execution given a pipeline ID and an execution number.
func (s *ExecutionStore) Delete(ctx context.Context, pipelineID int64, executionNum int64) error {
	q := execution{PipelineID: pipelineID, Number: executionNum}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableExecution).Where(&q).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete execution")
	}

	return nil
}

func convertExecutionPipelineRepoJoins(rows []*executionPipelineRepoJoin) ([]*types.Execution, error) {
	executions := make([]*types.Execution, len(rows))
	for i, k := range rows {
		e, err := convertExecutionPipelineRepoJoin(k)
		if err != nil {
			return nil, err
		}
		executions[i] = e
	}
	return executions, nil
}

func convertExecutionPipelineRepoJoin(join *executionPipelineRepoJoin) (*types.Execution, error) {
	e, err := mapInternalToExecution(&join.execution)
	if err != nil {
		return nil, err
	}
	e.RepoUID = join.RepoUID.String
	e.PipelineUID = join.PipelineUID.String
	return e, nil
}
