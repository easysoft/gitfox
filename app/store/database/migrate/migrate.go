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

package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/easysoft/gitfox/pkg/migrate"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

//go:embed postgres/*.sql
var postgres embed.FS

//go:embed sqlite/*.sql
var sqlite embed.FS

//go:embed mysql/*.sql
var mysql embed.FS

const (
	tableName = "migrations"

	postgresDriverName = "postgres"
	postgresSourceDir  = "postgres"

	sqliteDriverName = "sqlite3"
	sqliteSourceDir  = "sqlite"

	mysqlDriverName = "mysql"
	mysqlSourceDir  = "mysql"
)

// Migrate performs the database migration.
func Migrate(ctx context.Context, db *sqlx.DB) error {
	opts, err := getMigrator(db)
	if err != nil {
		return fmt.Errorf("failed to get migrator: %w", err)
	}
	return migrate.New(opts).MigrateUp(ctx)
}

// To performs the database migration to the specific version.
func To(ctx context.Context, db *sqlx.DB, version string) error {
	opts, err := getMigrator(db)
	if err != nil {
		return fmt.Errorf("failed to get migrator: %w", err)
	}
	return migrate.New(opts).MigrateTo(ctx, version)
}

// Current returns the current version ID (the latest migration applied) of the database.
func Current(ctx context.Context, db *sqlx.DB) (string, error) {
	var (
		query               string
		migrationTableCount int
	)

	switch db.DriverName() {
	case sqliteDriverName:
		query = `
			SELECT count(*)
			FROM sqlite_master
			WHERE name = ? and type = 'table'`
	case postgresDriverName:
		query = `
			SELECT count(*)
			FROM information_schema.tables
			WHERE table_name = ? and table_schema = 'public'`
	case mysqlDriverName:
		query = `
			SELECT count(*)
			FROM information_schema.TABLES
			WHERE TABLE_NAME= ?`
	default:
		return "", fmt.Errorf("unsupported driver '%s'", db.DriverName())
	}

	if err := db.QueryRowContext(ctx, query, tableName).Scan(&migrationTableCount); err != nil {
		return "", fmt.Errorf("failed to check migration table existence: %w", err)
	}

	if migrationTableCount == 0 {
		return "", nil
	}

	var version string

	query = "select version from " + tableName + " limit 1"
	if err := db.QueryRowContext(ctx, query).Scan(&version); err != nil {
		return "", fmt.Errorf("failed to read current DB version from migration table: %w", err)
	}

	return version, nil
}

func getMigrator(db *sqlx.DB) (migrate.Options, error) {
	before := func(ctx context.Context, _ *sql.Tx, version string) error {
		ctx = log.Ctx(ctx).With().
			Str("migrate.version", version).
			Str("migrate.phase", "before").
			Logger().WithContext(ctx)
		log := log.Ctx(ctx)

		log.Info().Msg("[START]")
		defer log.Info().Msg("[DONE]")

		return nil
	}

	after := func(ctx context.Context, dbtx *sql.Tx, version string) error {
		ctx = log.Ctx(ctx).With().
			Str("migrate.version", version).
			Str("migrate.phase", "after").
			Logger().WithContext(ctx)
		log := log.Ctx(ctx)

		log.Info().Msg("[START]")
		defer log.Info().Msg("[DONE]")

		switch version {
		case "0039_alter_table_webhooks_uid":
			return migrateAfter_0039_alter_table_webhooks_uid(ctx, dbtx)
		case "0042_alter_table_rules":
			return migrateAfter_0042_alter_table_rules(ctx, dbtx)
		default:
			return nil
		}
	}

	opts := migrate.Options{
		After:      after,
		Before:     before,
		DriverName: db.DriverName(),
		DB:         db.DB,
		FS:         sqlite,
		Table:      tableName,
	}

	switch db.DriverName() {
	case sqliteDriverName:
		folder, _ := fs.Sub(sqlite, sqliteSourceDir)
		opts.FS = folder
	case postgresDriverName:
		folder, _ := fs.Sub(postgres, postgresSourceDir)
		opts.FS = folder
	case mysqlDriverName:
		folder, _ := fs.Sub(mysql, mysqlSourceDir)
		opts.FS = folder

	default:
		return migrate.Options{}, fmt.Errorf("unsupported driver '%s'", db.DriverName())
	}

	return opts, nil
}
