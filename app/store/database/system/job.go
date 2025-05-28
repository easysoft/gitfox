// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package system

import (
	"context"
	"time"

	"github.com/easysoft/gitfox/job"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ job.Store = (*JobStore)(nil)

func NewJobOrmStore(db *gorm.DB) *JobStore {
	return &JobStore{
		db: db,
	}
}

type JobStore struct {
	db *gorm.DB
}

const (
	tableJob = "jobs"
)

// Find fetches a job by its unique identifier.
func (s *JobStore) Find(ctx context.Context, uid string) (*job.Job, error) {
	result := &job.Job{UID: uid}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).First(result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find job by uid")
	}

	return result, nil
}

// DeleteByGroupID deletes all jobs for a group id.
func (s *JobStore) DeleteByGroupID(ctx context.Context, groupID string) (int64, error) {
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).Where(&job.Job{GroupID: groupID}).Delete(nil)
	if res.Error != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, res.Error, "failed to execute delete jobs by group id query")
	}

	return res.RowsAffected, nil
}

// ListByGroupID fetches all jobs for a group id.
func (s *JobStore) ListByGroupID(ctx context.Context, groupID string) ([]*job.Job, error) {
	q := job.Job{GroupID: groupID}
	dst := make([]*job.Job, 0)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).Where(&q).Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find job by group id")
	}

	return dst, nil
}

// Create creates a new job.
func (s *JobStore) Create(ctx context.Context, job *job.Job) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).Create(job).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}
	return nil
}

// Upsert creates or updates a job. If the job didn't exist it will insert it in the database,
// otherwise it will update it but only if its definition has changed.
func (s *JobStore) Upsert(ctx context.Context, job *job.Job) error {
	upsertFields := []string{"job_updated", "job_type", "job_priority", "job_data", "job_result",
		"job_max_duration_seconds", "job_max_retries", "job_state", "job_scheduled",
		"job_is_recurring", "job_recurring_cron",
	}

	conflictWhere := []clause.Expression{
		clause.Expr{SQL: "jobs.job_type <> ?", Vars: []interface{}{job.Type}},
		clause.Or(clause.Expr{SQL: "jobs.job_priority <> ?", Vars: []interface{}{job.Priority}}),
		clause.Or(clause.Expr{SQL: "jobs.job_data <> ?", Vars: []interface{}{job.Data}}),
		clause.Or(clause.Expr{SQL: "jobs.job_max_duration_seconds <> ?", Vars: []interface{}{job.MaxDurationSeconds}}),
		clause.Or(clause.Expr{SQL: "jobs.job_max_retries <> ?", Vars: []interface{}{job.MaxRetries}}),
		clause.Or(clause.Expr{SQL: "jobs.job_is_recurring <> ?", Vars: []interface{}{job.IsRecurring}}),
		clause.Or(clause.Expr{SQL: "jobs.job_recurring_cron <> ?", Vars: []interface{}{job.RecurringCron}})}

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "job_uid"}},
		DoUpdates: clause.AssignmentColumns(upsertFields),
		Where:     clause.Where{Exprs: conflictWhere},
	}).Create(job).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Upsert query failed")
	}

	return nil
}

// UpdateDefinition is used to update a job definition.
func (s *JobStore) UpdateDefinition(ctx context.Context, j *job.Job) error {
	updateFields := []string{"Updated", "Type", "Priority",
		"Data", "Result", "MaxDurationSeconds", "MaxRetries", "State",
		"Scheduled", "IsRecurring", "RecurringCron", "GroupID",
	}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).
		Where(&job.Job{UID: j.UID}).
		Select(updateFields).Updates(j)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update job")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrResourceNotFound
	}

	return nil
}

// UpdateExecution is used to update a job before and after execution.
func (s *JobStore) UpdateExecution(ctx context.Context, j *job.Job) error {
	updateFields := []string{"Updated", "Result", "State", "Scheduled", "TotalExecutions",
		"RunBy", "RunDeadline", "LastExecuted",
		"ConsecutiveFailures", "LastFailureError",
	}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).
		Where(&job.Job{UID: j.UID}).
		Select(updateFields).Updates(j)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update job")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrResourceNotFound
	}

	return nil
}

func (s *JobStore) UpdateProgress(ctx context.Context, j *job.Job) error {
	updateFields := []string{"Updated", "Result", "RunProgress"}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).
		Where(&job.Job{UID: j.UID, State: "running"}).
		Select(updateFields).Updates(j)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update job")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrResourceNotFound
	}

	return nil
}

// CountRunning returns number of jobs that are currently being run.
func (s *JobStore) CountRunning(ctx context.Context) (int, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).Where(&job.Job{State: job.JobStateRunning})

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "failed executing count running jobs query")
	}

	return int(count), nil
}

// ListReady returns a list of jobs that are ready for execution:
// The jobs with state="scheduled" and scheduled time in the past.
func (s *JobStore) ListReady(ctx context.Context, now time.Time, limit int) ([]*job.Job, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).
		Where(&job.Job{State: job.JobStateScheduled}).
		Where("job_scheduled <= ?", now.UnixMilli()).
		Order("job_priority desc, job_scheduled asc, job_uid asc").
		Limit(limit)

	result := make([]*job.Job, 0)

	if err := stmt.Scan(&result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to execute list scheduled jobs query")
	}

	return result, nil
}

// ListDeadlineExceeded returns a list of jobs that have exceeded their execution deadline.
func (s *JobStore) ListDeadlineExceeded(ctx context.Context, now time.Time) ([]*job.Job, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).
		Where(&job.Job{State: job.JobStateRunning}).
		Where("job_run_deadline <= ?", now.UnixMilli()).
		Order("job_run_deadline asc")

	result := make([]*job.Job, 0)

	if err := stmt.Scan(&result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to execute list overdue jobs query")
	}

	return result, nil
}

// NextScheduledTime returns a scheduled time of the next ready job or zero time if no such job exists.
func (s *JobStore) NextScheduledTime(ctx context.Context, now time.Time) (time.Time, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).
		Select("job_scheduled").
		Where(&job.Job{State: job.JobStateScheduled}).
		Where("job_scheduled > ?", now.UnixMilli()).
		Order("job_scheduled asc").Limit(1)

	result := make([]*job.Job, 0)
	res := stmt.Scan(&result)

	if res.Error != nil {
		return time.Time{}, database.ProcessGormSQLErrorf(ctx, res.Error, "failed to execute next scheduled time query")
	}

	if res.RowsAffected == 0 {
		return time.Time{}, nil
	}

	return time.UnixMilli(result[0].Scheduled), nil
}

// DeleteOld removes non-recurring jobs that have finished execution or have failed.
func (s *JobStore) DeleteOld(ctx context.Context, olderThan time.Time) (int64, error) {
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).
		Where("(job_state = ? OR job_state = ? OR job_state = ?)",
			enum.JobStateFinished, enum.JobStateFailed, enum.JobStateCanceled).
		Where("job_is_recurring = false").
		Where("job_last_executed < ?", olderThan.UnixMilli()).
		Delete(nil)

	if res.Error != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, res.Error, "failed to execute delete done jobs query")
	}

	return res.RowsAffected, nil
}

// DeleteByUID deletes a job by its unique identifier.
func (s *JobStore) DeleteByUID(ctx context.Context, jobUID string) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableJob).Delete(&job.Job{}, jobUID).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "failed to execute delete job query")
	}
	return nil
}
