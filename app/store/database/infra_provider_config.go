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

package database

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	infraProviderConfigIDColumn      = `ipconf_id`
	infraProviderConfigInsertColumns = `
		ipconf_uid,
		ipconf_display_name,
		ipconf_type,
		ipconf_space_id,
		ipconf_created,
		ipconf_updated
	`
	infraProviderConfigSelectColumns = "ipconf_id," + infraProviderConfigInsertColumns
	infraProviderConfigTable         = `infra_provider_configs`
)

type infraProviderConfig struct {
	ID         int64                  `db:"ipconf_id"`
	Identifier string                 `db:"ipconf_uid"`
	Name       string                 `db:"ipconf_display_name"`
	Type       enum.InfraProviderType `db:"ipconf_type"`
	SpaceID    int64                  `db:"ipconf_space_id"`
	Created    int64                  `db:"ipconf_created"`
	Updated    int64                  `db:"ipconf_updated"`
}

var _ store.InfraProviderConfigStore = (*infraProviderConfigStore)(nil)

// NewGitspaceConfigStore returns a new GitspaceConfigStore.
func NewInfraProviderConfigStore(db *sqlx.DB) store.InfraProviderConfigStore {
	return &infraProviderConfigStore{
		db: db,
	}
}

type infraProviderConfigStore struct {
	db *sqlx.DB
}

func (i infraProviderConfigStore) Update(ctx context.Context, infraProviderConfig *types.InfraProviderConfig) error {
	dbinfraProviderConfig := i.mapToInternalInfraProviderConfig(ctx, infraProviderConfig)
	stmt := database.Builder.
		Update(infraProviderConfigTable).
		Set("ipconf_display_name", dbinfraProviderConfig.Name).
		Set("ipconf_updated", dbinfraProviderConfig.Updated).
		Where("ipconf_id = ?", infraProviderConfig.ID)
	sql, args, err := stmt.ToSql()
	if err != nil {
		return errors.Wrap(err, "Failed to convert squirrel builder to sql")
	}
	db := dbtx.GetAccessor(ctx, i.db)
	if _, err := db.ExecContext(ctx, sql, args...); err != nil {
		return database.ProcessSQLErrorf(
			ctx, err, "Failed to update infra provider config %s", infraProviderConfig.Identifier)
	}
	return nil
}

func (i infraProviderConfigStore) Find(ctx context.Context, id int64) (*types.InfraProviderConfig, error) {
	stmt := database.Builder.
		Select(infraProviderConfigSelectColumns).
		From(infraProviderConfigTable).
		Where(infraProviderConfigIDColumn+" = $1", id) //nolint:goconst
	sql, args, err := stmt.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to convert squirrel builder to sql")
	}
	dst := new(infraProviderConfig)
	db := dbtx.GetAccessor(ctx, i.db)
	if err := db.GetContext(ctx, dst, sql, args...); err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err, "Failed to find infraprovider config %d", id)
	}
	return i.mapToInfraProviderConfig(ctx, dst), nil
}

func (i infraProviderConfigStore) FindByIdentifier(
	ctx context.Context,
	spaceID int64,
	identifier string,
) (*types.InfraProviderConfig, error) {
	stmt := database.Builder.
		Select(infraProviderConfigSelectColumns).
		From(infraProviderConfigTable).
		Where("ipconf_uid = $1", identifier). //nolint:goconst
		Where("ipconf_space_id = $2", spaceID)
	sql, args, err := stmt.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to convert squirrel builder to sql")
	}
	dst := new(infraProviderConfig)
	db := dbtx.GetAccessor(ctx, i.db)
	if err := db.GetContext(ctx, dst, sql, args...); err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err, "Failed to find infraprovider config %s", identifier)
	}
	return i.mapToInfraProviderConfig(ctx, dst), nil
}

func (i infraProviderConfigStore) Create(ctx context.Context, infraProviderConfig *types.InfraProviderConfig) error {
	stmt := database.Builder.
		Insert(infraProviderConfigTable).
		Columns(infraProviderConfigInsertColumns).
		Values(
			infraProviderConfig.Identifier,
			infraProviderConfig.Name,
			infraProviderConfig.Type,
			infraProviderConfig.SpaceID,
			infraProviderConfig.Created,
			infraProviderConfig.Updated,
		).
		Suffix(ReturningClause + infraProviderConfigIDColumn)
	sql, args, err := stmt.ToSql()
	if err != nil {
		return errors.Wrap(err, "Failed to convert squirrel builder to sql")
	}
	db := dbtx.GetAccessor(ctx, i.db)
	if err = db.QueryRowContext(ctx, sql, args...).Scan(&infraProviderConfig.ID); err != nil {
		return database.ProcessSQLErrorf(
			ctx, err, "infraprovider config create query failed for %s", infraProviderConfig.Identifier)
	}
	return nil
}

func (i infraProviderConfigStore) mapToInfraProviderConfig(
	_ context.Context,
	in *infraProviderConfig) *types.InfraProviderConfig {
	infraProviderConfigEntity := &types.InfraProviderConfig{
		ID:         in.ID,
		Identifier: in.Identifier,
		Name:       in.Name,
		Type:       in.Type,
		SpaceID:    in.SpaceID,
		Created:    in.Created,
		Updated:    in.Updated,
	}
	return infraProviderConfigEntity
}

func (i infraProviderConfigStore) mapToInternalInfraProviderConfig(
	_ context.Context,
	in *types.InfraProviderConfig) *infraProviderConfig {
	infraProviderConfigEntity := &infraProviderConfig{
		Identifier: in.Identifier,
		Name:       in.Name,
		Type:       in.Type,
		SpaceID:    in.SpaceID,
		Created:    in.Created,
		Updated:    in.Updated,
	}
	return infraProviderConfigEntity
}
