// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ store.PipelineStore = (*OrmStore)(nil)

const (
	tablePipeline = "pipelines"
)

// NewPipelineOrmStore returns a new PipelineStore.
func NewPipelineOrmStore(db *gorm.DB) *OrmStore {
	return &OrmStore{
		db: db,
	}
}

type OrmStore struct {
	db *gorm.DB
}

type pipeline struct {
	ID            int64  `gorm:"column:pipeline_id;primaryKey"`
	Description   string `gorm:"column:pipeline_description"`
	Identifier    string `gorm:"column:pipeline_uid"`
	Disabled      bool   `gorm:"column:pipeline_disabled"`
	CreatedBy     int64  `gorm:"column:pipeline_created_by"`
	Seq           int64  `gorm:"column:pipeline_seq"`
	RepoID        int64  `gorm:"column:pipeline_repo_id"`
	DefaultBranch string `gorm:"column:pipeline_default_branch"`
	ConfigPath    string `gorm:"column:pipeline_config_path"`
	Created       int64  `gorm:"column:pipeline_created"`
	Updated       int64  `gorm:"column:pipeline_updated"`
	Version       int64  `gorm:"column:pipeline_version"`
}

// Find returns a pipeline given a pipeline ID.
func (s *OrmStore) Find(ctx context.Context, id int64) (*types.Pipeline, error) {
	dst := new(pipeline)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline).First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pipeline")
	}
	return mapInternalToPipeline(dst), nil
}

// FindByIdentifier returns a pipeline for a given repo with a given Identifier.
func (s *OrmStore) FindByIdentifier(
	ctx context.Context,
	repoID int64,
	identifier string,
) (*types.Pipeline, error) {
	q := pipeline{RepoID: repoID, Identifier: identifier}
	dst := new(pipeline)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pipeline")
	}
	return mapInternalToPipeline(dst), nil
}

// Create creates a pipeline.
func (s *OrmStore) Create(ctx context.Context, pipeline *types.Pipeline) error {
	dbObj := mapPipelineToInternal(pipeline)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Pipeline query failed")
	}

	pipeline.ID = dbObj.ID
	return nil
}

// Update updates a pipeline.
func (s *OrmStore) Update(ctx context.Context, p *types.Pipeline) error {
	updatedAt := time.Now()
	dbPipeline := mapPipelineToInternal(p)

	dbPipeline.Version++
	dbPipeline.Updated = updatedAt.UnixMilli()

	updateFields := []string{"Description", "Identifier", "Seq",
		"Disabled", "DefaultBranch", "ConfigPath",
		"Updated", "Version",
	}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline).
		Where(&pipeline{ID: p.ID, Version: dbPipeline.Version - 1}).
		Select(updateFields).Updates(dbPipeline)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update pipeline")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrVersionConflict
	}

	p.Updated = dbPipeline.Updated
	p.Version = dbPipeline.Version
	return nil
}

// List lists all the pipelines for a repository.
func (s *OrmStore) List(
	ctx context.Context,
	repoID int64,
	filter *types.ListPipelinesFilter,
) ([]*types.Pipeline, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline).
		Where("pipeline_repo_id = ?", fmt.Sprint(repoID))

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(pipeline_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	stmt = stmt.Limit(int(database.Limit(filter.Size)))
	stmt = stmt.Offset(int(database.Offset(filter.Page, filter.Size)))

	dst := make([]*pipeline, 0)
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return mapInternalToPipelineList(dst), nil
}

// ListInSpace lists all the pipelines for a space.
func (s *OrmStore) ListInSpace(
	ctx context.Context,
	spaceID int64,
	filter types.ListPipelinesFilter,
) ([]*types.Pipeline, error) {
	// TODO: 与上游不一致
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline).
		Select("pipelines.*, repositories.repo_id, repositories.repo_uid").
		Joins("INNER JOIN repositories ON pipelines.pipeline_repo_id = repositories.repo_id").
		Where("repositories.repo_parent_id = ?", spaceID)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(pipeline_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	stmt = stmt.Limit(database.GormLimit(filter.Size))
	stmt = stmt.Offset(database.GormOffset(filter.Page, filter.Size))

	dst := []*pipelineRepoJoin{}
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return convertPipelineRepoJoins(dst), nil
}

// CountInSpace counts the number of pipelines in a space.
func (s *OrmStore) CountInSpace(
	ctx context.Context,
	spaceID int64,
	filter types.ListPipelinesFilter,
) (int64, error) {
	stmt := s.db.Table("pipelines").
		Select("count(*)").
		Joins("INNER JOIN repositories ON pipelines.pipeline_repo_id = repositories.repo_id").
		Where("repositories.repo_parent_id = ?", fmt.Sprint(spaceID))

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(pipeline_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}

// ListLatest lists all the pipelines under a repository with information
// about the latest build if available.
func (s *OrmStore) ListLatest(
	ctx context.Context,
	repoID int64,
	filter *types.ListPipelinesFilter,
) ([]*types.Pipeline, error) {
	subQuerySQL := s.db.Table(tableExecution).
		Select("execution_pipeline_id", "MAX(execution_id) AS execution_id").
		Where("execution_repo_id = ?", repoID).Group("execution_pipeline_id")

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline).
		Joins("left join (?) AS max_executions ON pipelines.pipeline_id = max_executions.execution_pipeline_id", subQuerySQL).
		Joins("left join executions ON executions.execution_id = max_executions.execution_id").
		Where("pipeline_repo_id = ?", repoID)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(pipeline_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}
	stmt = stmt.Limit(int(database.Limit(filter.Size)))
	stmt = stmt.Offset(int(database.Offset(filter.Page, filter.Size)))

	dst := make([]*pipelineExecutionJoin, 0)
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return convert(dst), nil
}

// UpdateOptLock updates the pipeline using the optimistic locking mechanism.
func (s *OrmStore) UpdateOptLock(ctx context.Context,
	pipeline *types.Pipeline,
	mutateFn func(pipeline *types.Pipeline) error) (*types.Pipeline, error) {
	for {
		dup := *pipeline

		err := mutateFn(&dup)
		if err != nil {
			return nil, err
		}

		err = s.Update(ctx, &dup)
		if err == nil {
			return &dup, nil
		}
		if !errors.Is(err, gitfox_store.ErrVersionConflict) {
			return nil, err
		}

		pipeline, err = s.Find(ctx, pipeline.ID)
		if err != nil {
			return nil, err
		}
	}
}

// Count of pipelines under a repo, if repoID is zero it will count all pipelines in the system.
func (s *OrmStore) Count(ctx context.Context, repoID int64, filter *types.ListPipelinesFilter,
) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline)

	if repoID > 0 {
		stmt = stmt.Where("pipeline_repo_id = ?", repoID)
	}

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(pipeline_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}

// Delete deletes a pipeline given a pipeline ID.
func (s *OrmStore) Delete(ctx context.Context, id int64) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline).Delete(&pipeline{}, id).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete pipeline")
	}

	return nil
}

// DeleteByIdentifier deletes a pipeline with a given Identifier under a given repo.
func (s *OrmStore) DeleteByIdentifier(ctx context.Context, repoID int64, identifier string) error {
	q := pipeline{RepoID: repoID, Identifier: identifier}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePipeline).Where(&q).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete pipeline")
	}

	return nil
}

// Increment increments the pipeline sequence number. It will keep retrying in case
// of optimistic lock errors.
func (s *OrmStore) IncrementSeqNum(ctx context.Context, pipeline *types.Pipeline) (*types.Pipeline, error) {
	for {
		var err error
		pipeline.Seq++
		err = s.Update(ctx, pipeline)
		if err == nil {
			return pipeline, nil
		} else if !errors.Is(err, gitfox_store.ErrVersionConflict) {
			return pipeline, errors.Wrap(err, "could not increment pipeline sequence number")
		}
		pipeline, err = s.Find(ctx, pipeline.ID)
		if err != nil {
			return nil, errors.Wrap(err, "could not increment pipeline sequence number")
		}
	}
}

func mapInternalToPipeline(in *pipeline) *types.Pipeline {
	return &types.Pipeline{
		ID:            in.ID,
		Description:   in.Description,
		Identifier:    in.Identifier,
		Disabled:      in.Disabled,
		CreatedBy:     in.CreatedBy,
		Seq:           in.Seq,
		RepoID:        in.RepoID,
		DefaultBranch: in.DefaultBranch,
		ConfigPath:    in.ConfigPath,
		Created:       in.Created,
		Updated:       in.Updated,
		Version:       in.Version,
	}
}

func mapPipelineToInternal(in *types.Pipeline) *pipeline {
	return &pipeline{
		ID:            in.ID,
		Description:   in.Description,
		Identifier:    in.Identifier,
		Disabled:      in.Disabled,
		CreatedBy:     in.CreatedBy,
		Seq:           in.Seq,
		RepoID:        in.RepoID,
		DefaultBranch: in.DefaultBranch,
		ConfigPath:    in.ConfigPath,
		Created:       in.Created,
		Updated:       in.Updated,
		Version:       in.Version,
	}
}

func mapInternalToPipelineList(in []*pipeline) []*types.Pipeline {
	pipelines := make([]*types.Pipeline, len(in))
	for i, k := range in {
		s := mapInternalToPipeline(k)
		pipelines[i] = s
	}
	return pipelines
}
