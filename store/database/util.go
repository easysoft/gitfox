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
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/easysoft/gitfox/store"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// default query range limit.
const defaultLimit = 100

// limit returns the page size to a sql limit.
func Limit(size int) uint64 {
	if size == 0 {
		size = defaultLimit
	}
	return uint64(size)
}

func GormLimit(size int) int {
	if size == 0 {
		size = defaultLimit
	}
	return size
}

// offset converts the page to a sql offset.
func Offset(page, size int) uint64 {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = defaultLimit
	}
	page--
	return uint64(page * size)
}

func GormOffset(page, size int) int {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = defaultLimit
	}
	page--
	return page * size
}

// Logs the error and message, returns either the provided message.
// Always logs the full message with error as warning.
//
//nolint:unparam // revisit error processing
func ProcessSQLErrorf(ctx context.Context, err error, format string, args ...interface{}) error {
	// If it's a known error, return converted error instead.
	translatedError := err
	switch {
	case errors.Is(err, sql.ErrNoRows):
		translatedError = store.ErrResourceNotFound
	case isSQLUniqueConstraintError(err):
		translatedError = store.ErrDuplicate
	case isSQLForeignKeyViolationError(err):
		translatedError = store.ErrForeignKeyViolation
	default:
	}

	//nolint:errorlint // we want to match exactly here.
	if translatedError != err {
		log.Ctx(ctx).Debug().Err(err).Msgf("translated sql error to: %s", translatedError)
	}

	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), translatedError)
}

func ProcessGormSQLErrorf(ctx context.Context, err error, format string, args ...interface{}) error {
	// If it's a known error, return converted error instead.
	translatedError := err
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		translatedError = store.ErrResourceNotFound
	case isSQLUniqueConstraintError(err):
		// todo: add mysql unique error detect
		translatedError = store.ErrDuplicate
	case errors.Is(err, gorm.ErrDuplicatedKey):
		translatedError = store.ErrDuplicate
	default:
	}

	//nolint:errorlint // we want to match exactly here.
	if translatedError != err {
		log.Ctx(ctx).Debug().Err(err).Msgf("translated sql error to: %s", translatedError)
	}

	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), translatedError)
}

func EncodeToJSONString(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}
