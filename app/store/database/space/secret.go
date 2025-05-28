// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package space

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

var _ store.SecretStore = (*SecretStore)(nil)

const (
	tableSecret = "secrets"
)

// NewSecretOrmStore returns a new SecretStore.
func NewSecretOrmStore(db *gorm.DB) *SecretStore {
	return &SecretStore{
		db: db,
	}
}

type SecretStore struct {
	db *gorm.DB
}

// Find returns a secret given a secret ID.
func (s *SecretStore) Find(ctx context.Context, id int64) (*types.Secret, error) {
	dst := &types.Secret{ID: id}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSecret).First(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find secret")
	}
	return dst, nil
}

// FindByIdentifier returns a secret in a given space with a given identifier.
func (s *SecretStore) FindByIdentifier(ctx context.Context, spaceID int64, identifier string) (*types.Secret, error) {
	q := types.Secret{SpaceID: spaceID, Identifier: identifier}
	dst := new(types.Secret)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSecret).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find secret")
	}
	return dst, nil
}

// Create creates a secret.
func (s *SecretStore) Create(ctx context.Context, secret *types.Secret) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSecret).Create(secret).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "secret query failed")
	}

	return nil
}

func (s *SecretStore) Update(ctx context.Context, p *types.Secret) error {
	updatedAt := time.Now()
	secret := *p

	secret.Version++
	secret.Updated = updatedAt.UnixMilli()

	updateFields := []string{"Description", "Identifier", "Data", "Updated", "Version"}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSecret).
		Where(&types.Secret{ID: p.ID, Version: secret.Version - 1}).
		Select(updateFields).Updates(&secret)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update secret")
	}

	count := res.RowsAffected

	if count == 0 {
		return gitfox_store.ErrVersionConflict
	}

	p.Version = secret.Version
	p.Updated = secret.Updated
	return nil
}

// UpdateOptLock updates the pipeline using the optimistic locking mechanism.
func (s *SecretStore) UpdateOptLock(ctx context.Context,
	secret *types.Secret,
	mutateFn func(secret *types.Secret) error,
) (*types.Secret, error) {
	for {
		dup := *secret

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

		secret, err = s.Find(ctx, secret.ID)
		if err != nil {
			return nil, err
		}
	}
}

// List lists all the secrets present in a space.
func (s *SecretStore) List(ctx context.Context, parentID int64, filter types.ListQueryFilter) ([]*types.Secret, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSecret).
		Where("secret_space_id = ?", parentID)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(secret_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	stmt = stmt.Limit(int(database.Limit(filter.Size)))
	stmt = stmt.Offset(int(database.Offset(filter.Page, filter.Size)))

	dst := []*types.Secret{}
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return dst, nil
}

// ListAll lists all the secrets present in a space.
func (s *SecretStore) ListAll(ctx context.Context, parentID int64) ([]*types.Secret, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSecret).
		Where("secret_space_id = ?", parentID)

	dst := []*types.Secret{}
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return dst, nil
}

// Delete deletes a secret given a secret ID.
func (s *SecretStore) Delete(ctx context.Context, id int64) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSecret).Where(&types.Secret{ID: id}).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete secret")
	}

	return nil
}

// DeleteByIdentifier deletes a secret with a given identifier in a space.
func (s *SecretStore) DeleteByIdentifier(ctx context.Context, spaceID int64, identifier string) error {
	q := types.Secret{SpaceID: spaceID, Identifier: identifier}

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSecret).Where(&q).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete secret")
	}

	return nil
}

// Count of secrets in a space.
func (s *SecretStore) Count(ctx context.Context, parentID int64, filter types.ListQueryFilter) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableSecret).
		Where("secret_space_id = ?", parentID)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(secret_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}
