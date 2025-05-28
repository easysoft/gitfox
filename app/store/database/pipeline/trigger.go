// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ store.TriggerStore = (*TriggerStore)(nil)

type trigger struct {
	ID          int64  `gorm:"column:trigger_id;primaryKey"`
	Identifier  string `gorm:"column:trigger_uid"`
	Description string `gorm:"column:trigger_description"`
	Type        string `gorm:"column:trigger_type"`
	Secret      string `gorm:"column:trigger_secret"`
	PipelineID  int64  `gorm:"column:trigger_pipeline_id"`
	RepoID      int64  `gorm:"column:trigger_repo_id"`
	CreatedBy   int64  `gorm:"column:trigger_created_by"`
	Disabled    bool   `gorm:"column:trigger_disabled"`
	Actions     string `gorm:"column:trigger_actions"`
	Created     int64  `gorm:"column:trigger_created"`
	Updated     int64  `gorm:"column:trigger_updated"`
	Version     int64  `gorm:"column:trigger_version"`
}

func mapInternalToTrigger(trigger *trigger) (*types.Trigger, error) {
	var actions []enum.TriggerAction
	err := json.Unmarshal([]byte(trigger.Actions), &actions)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal trigger.actions")
	}

	return &types.Trigger{
		ID:          trigger.ID,
		Description: trigger.Description,
		Type:        trigger.Type,
		Secret:      trigger.Secret,
		PipelineID:  trigger.PipelineID,
		RepoID:      trigger.RepoID,
		CreatedBy:   trigger.CreatedBy,
		Disabled:    trigger.Disabled,
		Actions:     actions,
		Identifier:  trigger.Identifier,
		Created:     trigger.Created,
		Updated:     trigger.Updated,
		Version:     trigger.Version,
	}, nil
}

func mapInternalToTriggerList(triggers []*trigger) ([]*types.Trigger, error) {
	ret := make([]*types.Trigger, len(triggers))
	for i, t := range triggers {
		trigger, err := mapInternalToTrigger(t)
		if err != nil {
			return nil, err
		}
		ret[i] = trigger
	}
	return ret, nil
}

func mapTriggerToInternal(t *types.Trigger) *trigger {
	return &trigger{
		ID:          t.ID,
		Identifier:  t.Identifier,
		Description: t.Description,
		Type:        t.Type,
		PipelineID:  t.PipelineID,
		Secret:      t.Secret,
		RepoID:      t.RepoID,
		CreatedBy:   t.CreatedBy,
		Disabled:    t.Disabled,
		Actions:     database.EncodeToJSONString(t.Actions),
		Created:     t.Created,
		Updated:     t.Updated,
		Version:     t.Version,
	}
}

// NewTriggerOrmStore returns a new TriggerStore.
func NewTriggerOrmStore(db *gorm.DB) *TriggerStore {
	return &TriggerStore{
		db: db,
	}
}

type TriggerStore struct {
	db *gorm.DB
}

const (
	tableTrigger = "triggers"
)

// FindByIdentifier returns an trigger given a pipeline ID and a trigger identifier.
func (s *TriggerStore) FindByIdentifier(
	ctx context.Context,
	pipelineID int64,
	identifier string,
) (*types.Trigger, error) {
	q := trigger{PipelineID: pipelineID, Identifier: identifier}

	dst := new(trigger)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTrigger).Omit("RepoID", "CreatedBy").
		Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find trigger")
	}
	return mapInternalToTrigger(dst)
}

// Create creates a new trigger in the datastore.
func (s *TriggerStore) Create(ctx context.Context, t *types.Trigger) error {
	dbTrigger := mapTriggerToInternal(t)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTrigger).Create(dbTrigger).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Trigger query failed")
	}

	t.ID = dbTrigger.ID
	return nil
}

// Update tries to update an trigger in the datastore with optimistic locking.
func (s *TriggerStore) Update(ctx context.Context, t *types.Trigger) error {
	updatedAt := time.Now()
	dbTrigger := mapTriggerToInternal(t)

	dbTrigger.Version++
	dbTrigger.Updated = updatedAt.UnixMilli()

	updateFields := []string{"Identifier", "Description", "Disabled", "Updated", "Actions", "Version"}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTrigger).
		Where(&trigger{ID: t.ID, Version: dbTrigger.Version - 1}).
		Select(updateFields).Updates(dbTrigger)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update stage")
	}

	if res.RowsAffected == 0 {
		return gitfox_store.ErrVersionConflict
	}

	t.Version = dbTrigger.Version
	t.Updated = dbTrigger.Updated
	return nil
}

// UpdateOptLock updates the pipeline using the optimistic locking mechanism.
func (s *TriggerStore) UpdateOptLock(ctx context.Context,
	trigger *types.Trigger,
	mutateFn func(trigger *types.Trigger) error) (*types.Trigger, error) {
	for {
		dup := *trigger

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

		trigger, err = s.FindByIdentifier(ctx, trigger.PipelineID, trigger.Identifier)
		if err != nil {
			return nil, err
		}
	}
}

// List lists the triggers for a given pipeline ID.
func (s *TriggerStore) List(
	ctx context.Context,
	pipelineID int64,
	filter types.ListQueryFilter,
) ([]*types.Trigger, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTrigger).Omit("RepoID", "CreatedBy").
		Where("trigger_pipeline_id = ?", pipelineID)

	stmt = stmt.Limit(int(database.Limit(filter.Size)))
	stmt = stmt.Offset(int(database.Offset(filter.Page, filter.Size)))

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(trigger_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	dst := make([]*trigger, 0)
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return mapInternalToTriggerList(dst)
}

// ListAllEnabled lists all enabled triggers for a given repo without pagination.
func (s *TriggerStore) ListAllEnabled(
	ctx context.Context,
	repoID int64,
) ([]*types.Trigger, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTrigger).Omit("RepoID", "CreatedBy").
		Where("trigger_repo_id = ? AND trigger_disabled = false", repoID)

	dst := make([]*trigger, 0)
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return mapInternalToTriggerList(dst)
}

// Count of triggers under a given pipeline.
func (s *TriggerStore) Count(ctx context.Context, pipelineID int64, filter types.ListQueryFilter) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTrigger).
		Where("trigger_pipeline_id = ?", pipelineID)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(trigger_uid) LIKE ?", fmt.Sprintf("%%%s%%", filter.Query))
	}

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}

// Delete deletes an trigger given a pipeline ID and a trigger identifier.
func (s *TriggerStore) DeleteByIdentifier(ctx context.Context, pipelineID int64, identifier string) error {
	q := trigger{PipelineID: pipelineID, Identifier: identifier}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTrigger).Where(&q).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete trigger")
	}

	return nil
}
