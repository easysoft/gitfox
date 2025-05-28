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

package publickey

import (
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/errors"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ store.PublicKeyStore = PublicKeyStore{}

// NewPublicKeyStore returns a new PublicKeyStore.
func NewPublicKeyStore(db *gorm.DB) PublicKeyStore {
	return PublicKeyStore{
		db: db,
	}
}

// PublicKeyStore implements a store.PublicKeyStore backed by a relational database.
type PublicKeyStore struct {
	db *gorm.DB
}

type publicKey struct {
	ID int64 `gorm:"column:public_key_id"`

	PrincipalID int64 `gorm:"column:public_key_principal_id"`

	Created  int64    `gorm:"column:public_key_created"`
	Verified null.Int `gorm:"column:public_key_verified"`

	Identifier string `gorm:"column:public_key_identifier"`
	Usage      string `gorm:"column:public_key_usage"`

	Fingerprint string `gorm:"column:public_key_fingerprint"`
	Content     string `gorm:"column:public_key_content"`
	Comment     string `gorm:"column:public_key_comment"`
	Type        string `gorm:"column:public_key_type"`
}

// Find fetches a job by its unique identifier.
func (s PublicKeyStore) Find(ctx context.Context, id int64) (*types.PublicKey, error) {
	result := &publicKey{}
	if err := s.db.Where("public_key_id = ?", id).First(result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find public key by id")
	}

	key := mapToPublicKey(result)

	return &key, nil
}

// FindByIdentifier returns a public key given a principal ID and an identifier.
func (s PublicKeyStore) FindByIdentifier(
	ctx context.Context,
	principalID int64,
	identifier string,
) (*types.PublicKey, error) {
	result := &publicKey{}
	if err := s.db.Where("public_key_principal_id = ? AND LOWER(public_key_identifier) = ?", principalID, strings.ToLower(identifier)).First(result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find public key by principal and identifier")
	}

	key := mapToPublicKey(result)

	return &key, nil
}

// Create creates a new public key.
func (s PublicKeyStore) Create(ctx context.Context, key *types.PublicKey) error {
	dbKey := mapToInternalPublicKey(key)

	if err := s.db.Create(&dbKey).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to create public key")
	}

	key.ID = dbKey.ID

	return nil
}

// DeleteByIdentifier deletes a public key.
func (s PublicKeyStore) DeleteByIdentifier(ctx context.Context, principalID int64, identifier string) error {
	result := s.db.Where("public_key_principal_id = ? AND LOWER(public_key_identifier) = ?", principalID, strings.ToLower(identifier)).Delete(&publicKey{})
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to delete public key")
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("Key not found")
	}

	return nil
}

// MarkAsVerified updates the public key to mark it as verified.
func (s PublicKeyStore) MarkAsVerified(ctx context.Context, id int64, verified int64) error {
	if err := s.db.Model(&publicKey{}).Where("public_key_id = ?", id).Update("public_key_verified", verified).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to mark public key as verified")
	}

	return nil
}

func (s PublicKeyStore) Count(
	ctx context.Context,
	principalID int64,
	filter *types.PublicKeyFilter,
) (int, error) {
	var count int64
	if err := s.db.Model(&publicKey{}).Where("public_key_principal_id = ?", principalID).Count(&count).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed to count public keys")
	}

	return int(count), nil
}

// List returns the public keys for the principal.
func (s PublicKeyStore) List(
	ctx context.Context,
	principalID int64,
	filter *types.PublicKeyFilter,
) ([]types.PublicKey, error) {
	stmt := s.db.Model(&publicKey{}).Where("public_key_principal_id = ?", principalID)

	stmt = s.applyQueryFilter(stmt, filter)
	stmt = s.applySortFilter(stmt, filter)

	keys := make([]publicKey, 0)
	if err := stmt.Find(&keys).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list public keys")
	}

	return mapToPublicKeys(keys), nil
}

// ListByFingerprint returns public keys given a fingerprint and key usage.
func (s PublicKeyStore) ListByFingerprint(
	ctx context.Context,
	fingerprint string,
) ([]types.PublicKey, error) {
	stmt := s.db.Model(&publicKey{}).Where("public_key_fingerprint = ?", fingerprint).Order("public_key_created ASC")

	keys := make([]publicKey, 0)
	if err := stmt.Find(&keys).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list public keys by fingerprint")
	}

	return mapToPublicKeys(keys), nil
}

func (PublicKeyStore) applyQueryFilter(
	stmt *gorm.DB,
	filter *types.PublicKeyFilter,
) *gorm.DB {
	if filter.Query != "" {
		stmt = stmt.Where("LOWER(public_key_identifier) LIKE ?",
			fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	return stmt
}

func (PublicKeyStore) applySortFilter(
	stmt *gorm.DB,
	filter *types.PublicKeyFilter,
) *gorm.DB {
	stmt = stmt.Limit(database.GormLimit(filter.Size))
	stmt = stmt.Offset(database.GormOffset(filter.Page, filter.Size))

	order := filter.Order
	if order == enum.OrderDefault {
		order = enum.OrderAsc
	}

	switch filter.Sort {
	case enum.PublicKeySortIdentifier:
		stmt = stmt.Order("public_key_identifier " + order.String())
	case enum.PublicKeySortCreated:
		stmt = stmt.Order("public_key_created " + order.String())
	}

	return stmt
}

func mapToInternalPublicKey(in *types.PublicKey) publicKey {
	return publicKey{
		ID:          in.ID,
		PrincipalID: in.PrincipalID,
		Created:     in.Created,
		Verified:    null.IntFromPtr(in.Verified),
		Identifier:  in.Identifier,
		Usage:       string(in.Usage),
		Fingerprint: in.Fingerprint,
		Content:     in.Content,
		Comment:     in.Comment,
		Type:        in.Type,
	}
}

func mapToPublicKey(in *publicKey) types.PublicKey {
	return types.PublicKey{
		ID:          in.ID,
		PrincipalID: in.PrincipalID,
		Created:     in.Created,
		Verified:    in.Verified.Ptr(),
		Identifier:  in.Identifier,
		Usage:       enum.PublicKeyUsage(in.Usage),
		Fingerprint: in.Fingerprint,
		Content:     in.Content,
		Comment:     in.Comment,
		Type:        in.Type,
	}
}

func mapToPublicKeys(
	keys []publicKey,
) []types.PublicKey {
	res := make([]types.PublicKey, len(keys))
	for i := 0; i < len(keys); i++ {
		res[i] = mapToPublicKey(&keys[i])
	}
	return res
}
