// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package system

import (
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

var _ store.PluginStore = (*PluginStore)(nil)

const (
	tablePlugin = "plugins"
)

// NewPluginOrmStore returns a new PluginStore.
func NewPluginOrmStore(db *gorm.DB) *PluginStore {
	return &PluginStore{
		db: db,
	}
}

type PluginStore struct {
	db *gorm.DB
}

// Create creates a new entry in the plugin datastore.
func (s *PluginStore) Create(ctx context.Context, plugin *types.Plugin) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePlugin).Create(plugin).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "plugin query failed")
	}

	return nil
}

// Find finds a version of a plugin.
func (s *PluginStore) Find(ctx context.Context, name, version string) (*types.Plugin, error) {
	q := types.Plugin{Identifier: name, Version: version}

	dst := new(types.Plugin)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePlugin).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find pipeline")
	}

	return dst, nil
}

// List returns back the list of plugins along with their associated schemas.
func (s *PluginStore) List(
	ctx context.Context,
	filter types.ListQueryFilter,
) ([]*types.Plugin, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePlugin)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(plugin_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	stmt = stmt.Limit(int(database.Limit(filter.Size)))
	stmt = stmt.Offset(int(database.Offset(filter.Page, filter.Size)))

	dst := []*types.Plugin{}
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return dst, nil
}

// ListAll returns back the full list of plugins in the database.
func (s *PluginStore) ListAll(
	ctx context.Context,
) ([]*types.Plugin, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePlugin)

	dst := []*types.Plugin{}
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return dst, nil
}

// Count of plugins matching the filter criteria.
func (s *PluginStore) Count(ctx context.Context, filter types.ListQueryFilter) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePlugin)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(plugin_uid) LIKE ?", fmt.Sprintf("%%%s%%", filter.Query))
	}

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}

// Update updates a plugin row.
func (s *PluginStore) Update(ctx context.Context, p *types.Plugin) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tablePlugin).Save(p).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to update plugin")
	}

	return nil
}
