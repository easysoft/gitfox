// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package principal

import (
	"context"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.TokenStore = (*TokenStore)(nil)

// NewTokenOrmStore returns a new TokenStore.
func NewTokenOrmStore(db *gorm.DB) *TokenStore {
	return &TokenStore{db}
}

// TokenStore implements a TokenStore backed by a relational database.
type TokenStore struct {
	db *gorm.DB
}

const tableToken = "tokens"

// Find finds the token by id.
func (s *TokenStore) Find(ctx context.Context, id int64) (*types.Token, error) {
	dst := new(types.Token)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableToken).First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find token")
	}

	return dst, nil
}

// FindByIdentifier finds the token by principalId and token identifier.
func (s *TokenStore) FindByIdentifier(ctx context.Context, principalID int64, identifier string) (*types.Token, error) {
	dst := new(types.Token)
	q := types.Token{PrincipalID: principalID, Identifier: identifier}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableToken).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find token by identifier")
	}

	return dst, nil
}

// Create saves the token details.
func (s *TokenStore) Create(ctx context.Context, token *types.Token) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableToken).Create(token).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	return nil
}

// Delete deletes the token with the given id.
func (s *TokenStore) Delete(ctx context.Context, id int64) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableToken).Delete(types.Token{}, id).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "The delete query failed")
	}

	return nil
}

// DeleteExpiredBefore deletes all tokens that expired before the provided time.
// If tokenTypes are provided, then only tokens of that type are deleted.
func (s *TokenStore) DeleteExpiredBefore(
	ctx context.Context,
	before time.Time,
	tknTypes []enum.TokenType,
) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableToken).Where("token_expires_at < ?", before.UnixMilli())

	if len(tknTypes) > 0 {
		stmt = stmt.Where("token_type in ?", tknTypes)
	}

	res := stmt.Delete(nil)
	if res.Error != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, res.Error, "failed to execute delete token query")
	}

	return res.RowsAffected, nil
}

// Count returns a count of tokens of a specifc type for a specific principal.
func (s *TokenStore) Count(ctx context.Context, principalID int64, tokenType enum.TokenType) (int64, error) {
	q := types.Token{PrincipalID: principalID, Type: tokenType}
	var count int64
	err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableToken).Where(&q).Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}

	return count, nil
}

// List returns a list of tokens of a specific type for a specific principal.
func (s *TokenStore) List(ctx context.Context,
	principalID int64, tokenType enum.TokenType) ([]*types.Token, error) {
	q := types.Token{PrincipalID: principalID, Type: tokenType}
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableToken).Where(&q).Order("token_issued_at DESC")

	dst := []*types.Token{}

	// TODO: custom filters / sorting for tokens.

	err := stmt.Scan(&dst).Error
	if err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing token list query")
	}
	return dst, nil
}
