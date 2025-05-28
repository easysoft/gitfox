// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package logs

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"

	"gorm.io/gorm"
)

const tableLog = "logs"

// NewDatabaseLogOrmStore returns a new LogStore.
func NewDatabaseLogOrmStore(db *gorm.DB) store.LogStore {
	return &logOrmStore{
		db: db,
	}
}

type logOrmStore struct {
	db *gorm.DB
}

// Find returns a log given a log ID.
func (s *logOrmStore) Find(ctx context.Context, stepID int64) (io.ReadCloser, error) {
	var err error
	dst := new(logs)
	if err = dbtx.GetOrmAccessor(ctx, s.db).Table(tableLog).First(dst, stepID).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find log")
	}
	return io.NopCloser(
		bytes.NewBuffer(dst.Data),
	), err
}

// Create creates a log.
func (s *logOrmStore) Create(ctx context.Context, stepID int64, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("could not read log data: %w", err)
	}
	params := &logs{
		ID:   stepID,
		Data: data,
	}

	if err = dbtx.GetOrmAccessor(ctx, s.db).Table(tableLog).Create(params).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "log query failed")
	}

	return nil
}

// Update overrides existing logs data.
func (s *logOrmStore) Update(ctx context.Context, stepID int64, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("could not read log data: %w", err)
	}

	err = dbtx.GetOrmAccessor(ctx, s.db).Table(tableLog).
		UpdateColumn("log_data", data).Error

	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to update log")
	}

	return nil
}

// Delete deletes a log given a log ID.
func (s *logOrmStore) Delete(ctx context.Context, stepID int64) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableLog).Delete(&logs{}, stepID).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err, "Could not delete log")
	}

	return nil
}
