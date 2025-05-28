// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	store2 "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ store.SettingsStore = (*SettingsStore)(nil)

// NewSettingsOrmStore returns a new SettingsStore.
func NewSettingsOrmStore(db *gorm.DB) *SettingsStore {
	return &SettingsStore{
		db: db,
	}
}

// SettingsStore implements store.SettingsStore backed by a relational database.
type SettingsStore struct {
	db *gorm.DB
}

// setting is an internal representation used to store setting data in the database.
type setting struct {
	ID      int64           `gorm:"column:setting_id"`
	SpaceID null.Int        `gorm:"column:setting_space_id"`
	RepoID  null.Int        `gorm:"column:setting_repo_id"`
	Key     string          `gorm:"column:setting_key"`
	Value   json.RawMessage `gorm:"column:setting_value"`
}

const (
	tableSettings = "settings"
)

func (s *SettingsStore) Find(
	ctx context.Context,
	scope enum.SettingsScope,
	scopeID int64,
	key string,
) (json.RawMessage, error) {
	dst, err := s.find(ctx, scope, scopeID, key)
	if err != nil {
		return nil, err
	}

	return dst.Value, nil
}

func (s *SettingsStore) find(ctx context.Context,
	scope enum.SettingsScope,
	scopeID int64,
	key string,
) (*setting, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSettings).
		Where("LOWER(setting_key) = ?", strings.ToLower(key))

	switch scope {
	case enum.SettingsScopeSpace:
		stmt = stmt.Where("setting_space_id = ?", scopeID)
	case enum.SettingsScopeRepo:
		stmt = stmt.Where("setting_repo_id = ?", scopeID)
	case enum.SettingsScopeSystem:
		stmt = stmt.Where("setting_repo_id IS NULL AND setting_space_id IS NULL")
	default:
		return nil, fmt.Errorf("setting scope %q is not supported", scope)
	}

	dst := &setting{}
	if err := stmt.Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
	}

	return dst, nil
}

func (s *SettingsStore) FindMany(
	ctx context.Context,
	scope enum.SettingsScope,
	scopeID int64,
	keys ...string,
) (map[string]json.RawMessage, error) {
	if len(keys) == 0 {
		return map[string]json.RawMessage{}, nil
	}

	keysLower := make([]string, len(keys))
	for i, k := range keys {
		keysLower[i] = strings.ToLower(k)
	}

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSettings).
		Where("LOWER(setting_key) IN ?", keysLower)

	switch scope {
	case enum.SettingsScopeSpace:
		stmt = stmt.Where("setting_space_id = ?", scopeID)
	case enum.SettingsScopeRepo:
		stmt = stmt.Where("setting_repo_id = ?", scopeID)
	case enum.SettingsScopeSystem:
		stmt = stmt.Where("setting_repo_id IS NULL AND setting_space_id IS NULL")
	default:
		return nil, fmt.Errorf("setting scope %q is not supported", scope)
	}

	dst := []*setting{}
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
	}

	out := map[string]json.RawMessage{}
	for _, d := range dst {
		out[d.Key] = d.Value
	}

	return out, nil
}

// Upsert will check conflict and insert an object,
// or update current object's value.
// rewrite without `on conflict` clause for mysql compatible
// unique index on NULL field `setting_space_id` and `setting_repo_id`
// not work well for mysql
func (s *SettingsStore) Upsert(ctx context.Context,
	scope enum.SettingsScope,
	scopeID int64,
	key string,
	value json.RawMessage,
) error {
	var dbObj *setting
	var err error

	dbObj, err = s.find(ctx, scope, scopeID, key)
	if err != nil {
		if !errors.Is(err, store2.ErrResourceNotFound) {
			return err
		}
	}

	if dbObj != nil {
		dbObj.Value = value
	} else {
		dbObj, err = buildNewSetting(scope, scopeID, key, value)
		if err != nil {
			return err
		}
	}

	if err = dbtx.GetOrmAccessor(ctx, s.db).Table(tableSettings).Save(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Upsert query failed")
	}
	return nil
}

func buildNewSetting(scope enum.SettingsScope,
	scopeID int64,
	key string,
	value json.RawMessage,
) (*setting, error) {
	newObj := &setting{Key: key, Value: value}
	switch scope {
	case enum.SettingsScopeSpace:
		newObj.SpaceID = null.IntFrom(scopeID)
		newObj.RepoID = null.Int{}
	case enum.SettingsScopeRepo:
		newObj.SpaceID = null.Int{}
		newObj.RepoID = null.IntFrom(scopeID)
	case enum.SettingsScopeSystem:
		newObj.SpaceID = null.Int{}
		newObj.RepoID = null.Int{}
	default:
		return nil, fmt.Errorf("setting scope %q is not supported", scope)
	}
	return newObj, nil
}
