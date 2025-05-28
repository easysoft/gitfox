// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package dbtest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store/database/migrate"
	"github.com/easysoft/gitfox/store/database"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Hooks struct {
	T *testing.T
}

func init() {

}

// Before hook will print the query with it's args and return the context with the timestamp
func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	h.T.Helper()
	//h.T.Logf("sql > %s %+v", query, args)
	return context.WithValue(ctx, "begin", time.Now()), nil
}

// After hook will get the timestamp registered on the Before hook and print the elapsed time
func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	//begin := ctx.Value("begin").(time.Time)
	//fmt.Printf(". took: %s\n", time.Since(begin))
	return ctx, nil
}

func New(ctx context.Context, t *testing.T, suite string, tables ...any) (*sqlx.DB, *gorm.DB) {
	var cleanup func(db *gorm.DB)

	dataSource := filepath.Join(os.TempDir(), fmt.Sprintf("gitfox-%s-%d.db", suite, time.Now().Unix()))
	gormDB, err := database.ConnectGorm(ctx, "sqlite3", dataSource, database.GormConfigLogger{Level: logger.Info})
	require.NoError(t, err)

	sqlxDB, err := database.ConnectAndMigrate(ctx, "sqlite3", dataSource, func(ctx context.Context, dbx *sqlx.DB) error {
		return migrate.Migrate(ctx, dbx)
	})
	require.NoError(t, err)

	//sql.Register("sqlite3WithHook", sqlhooks.Wrap(sqlxDB.Driver(), &Hooks{T: t}))
	//db, err := sql.Open("sqlite3WithHook", dataSource)
	//if err != nil {
	//	panic(err)
	//}
	//
	//sqlxDB = sqlx.NewDb(db, "sqlite3WithHook")
	//require.NoError(t, err)

	cleanup = func(db *gorm.DB) {
		sqlDB, e := db.DB()
		if e == nil {
			_ = sqlDB.Close()
		}
		_ = os.Remove(dataSource)
	}

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("Database %q left intact for inspection", dataSource)
			return
		}

		sqlxDB.DB.Close()
		cleanup(gormDB)
	})

	//err = gormDB.Migrator().AutoMigrate(tables...)
	//require.NoError(t, err)

	return sqlxDB, gormDB
}

func NewDB(ctx context.Context, t *testing.T, suite string, tables ...any) *gorm.DB {
	var cleanup func(db *gorm.DB)
	dbName := filepath.Join(os.TempDir(), fmt.Sprintf("gitfox-%s-%d.db", suite, time.Now().Unix()))
	db, err := database.ConnectGorm(ctx, "sqlite3", dbName, database.GormConfigLogger{Level: logger.Info})
	require.NoError(t, err)

	cleanup = func(db *gorm.DB) {
		sqlDB, e := db.DB()
		if e == nil {
			_ = sqlDB.Close()
		}
		_ = os.Remove(dbName)
	}

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("Database %q left intact for inspection", dbName)
			return
		}

		cleanup(db)
	})

	err = db.Migrator().AutoMigrate(tables...)
	require.NoError(t, err)

	return db
}

// ClearTables removes all rows from given tables.
func ClearTables(t *testing.T, db *gorm.DB, tables ...any) error {
	if t.Failed() {
		return nil
	}

	for _, t := range tables {
		err := db.Where("TRUE").Delete(t).Error
		if err != nil {
			return err
		}
	}
	return nil
}
