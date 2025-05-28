// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitspace

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.GitspaceEventStore = (*gitspaceEventStore)(nil)

const (
	gitspaceEventIDColumn = `geven_id`
	gitspaceEventsColumns = `
		geven_event,
		geven_created,
		geven_entity_type,
		geven_query_key,
		geven_entity_id,
		geven_timestamp
	`
	gitspaceEventsColumnsWithID = gitspaceEventIDColumn + `,
		` + gitspaceEventsColumns
	gitspaceEventsTable = `gitspace_events`
)

type gitspaceEventStore struct {
	db *gorm.DB
}

type gitspaceEvent struct {
	ID         int64                   `gorm:"column:geven_id"`
	Event      enum.GitspaceEventType  `gorm:"column:geven_event"`
	Created    int64                   `gorm:"column:geven_created"`
	EntityType enum.GitspaceEntityType `gorm:"column:geven_entity_type"`
	QueryKey   string                  `gorm:"column:geven_query_key"`
	EntityID   int64                   `gorm:"column:geven_entity_id"`
	Timestamp  int64                   `gorm:"column:geven_timestamp"`
}

func NewGitspaceEventStore(db *gorm.DB) store.GitspaceEventStore {
	return &gitspaceEventStore{
		db: db,
	}
}

func (g gitspaceEventStore) FindLatestByTypeAndGitspaceConfigID(
	ctx context.Context,
	eventType enum.GitspaceEventType,
	gitspaceConfigID int64,
) (*types.GitspaceEvent, error) {
	gitspaceEventEntity := new(gitspaceEvent)
	if err := g.db.
		Where("geven_event = ?", eventType).
		Where("geven_entity_id = ?", gitspaceConfigID).
		Order("geven_timestamp DESC ").
		First(gitspaceEventEntity).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err, "Failed to find gitspace event for %d", gitspaceConfigID)
	}
	return g.mapGitspaceEvent(gitspaceEventEntity), nil
}

func (g gitspaceEventStore) Create(ctx context.Context, gitspaceEvent *types.GitspaceEvent) error {
	if err := g.db.
		Create(gitspaceEvent).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err, "failed to create gitspace event for %s", gitspaceEvent.QueryKey)
	}
	return nil
}

func (g gitspaceEventStore) List(
	ctx context.Context,
	filter *types.GitspaceEventFilter,
) ([]*types.GitspaceEvent, int, error) {
	gitspaceEventEntities := make([]*gitspaceEvent, 0)
	dbctx := g.db
	dbctx = g.setQueryFilter(dbctx, filter)
	dbctx = g.setPaginationFilter(dbctx, filter)
	if err := dbctx.Find(&gitspaceEventEntities).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err, "Failed to list gitspaceEvents")
	}
	var count int64
	if err := g.db.Table(gitspaceEventsTable).Count(&count).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err, "Failed to count gitspaceEvents")
	}
	gitspaceEvents := g.mapGitspaceEvents(gitspaceEventEntities)
	return gitspaceEvents, int(count), nil
}

func (g gitspaceEventStore) setQueryFilter(
	stmt *gorm.DB,
	filter *types.GitspaceEventFilter,
) *gorm.DB {
	if filter.QueryKey != "" {
		stmt = stmt.Where("geven_entity_uid = ?", filter.QueryKey)
	}
	if filter.EntityType != "" {
		stmt = stmt.Where("geven_entity_type = ?", filter.EntityType)
	}
	if filter.EntityID != 0 {
		stmt = stmt.Where("geven_entity_id = ?", filter.EntityID)
	}
	return stmt
}

func (g gitspaceEventStore) setPaginationFilter(
	stmt *gorm.DB,
	filter *types.GitspaceEventFilter,
) *gorm.DB {
	offset := (filter.Page - 1) * filter.Size
	stmt = stmt.Offset(offset).Limit(filter.Size)
	return stmt
}

func (g gitspaceEventStore) mapGitspaceEvents(gitspaceEventEntities []*gitspaceEvent) []*types.GitspaceEvent {
	gitspaceEvents := make([]*types.GitspaceEvent, len(gitspaceEventEntities))
	for index, gitspaceEventEntity := range gitspaceEventEntities {
		currentEntity := gitspaceEventEntity
		gitspaceEvents[index] = g.mapGitspaceEvent(currentEntity)
	}
	return gitspaceEvents
}

func (g gitspaceEventStore) mapGitspaceEvent(event *gitspaceEvent) *types.GitspaceEvent {
	return &types.GitspaceEvent{
		Event:      event.Event,
		Created:    event.Created,
		EntityType: event.EntityType,
		QueryKey:   event.QueryKey,
		EntityID:   event.EntityID,
		Timestamp:  event.Timestamp,
	}
}
