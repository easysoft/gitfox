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

package principal

import (
	"context"
	"fmt"
	"strings"

	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
)

// user is a DB representation of a user principal.
// It is required to allow storing transformed UIDs used for uniquness constraints and searching.
type user struct {
	types.User
	Type      principalType `gorm:"column:principal_type"`
	UIDUnique string        `gorm:"column:principal_uid_unique"`
}

// FindUser finds the user by id.
func (s *PrincipalOrmStore) FindUser(ctx context.Context, id int64) (*types.User, error) {
	dst := new(user)
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where(&user{Type: principalUser})
	if err := db.First(&dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by id query failed")
	}

	return s.mapDBUser(dst), nil
}

// FindUserByUID finds the user by uid.
func (s *PrincipalOrmStore) FindUserByUID(ctx context.Context, uid string) (*types.User, error) {
	// map the UID to unique UID before searching!
	uidUnique, err := s.uidTransformation(uid)
	if err != nil {
		// in case we fail to transform, return a not found (as it can't exist in the first place)
		log.Ctx(ctx).Debug().Msgf("failed to transform uid '%s': %s", uid, err.Error())
		return nil, gitfox_store.ErrResourceNotFound
	}

	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where(&user{Type: principalUser, UIDUnique: uidUnique})

	dst := new(user)
	if err = db.First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by uid query failed")
	}

	return s.mapDBUser(dst), nil
}

// FindUserByEmail finds the user by email.
func (s *PrincipalOrmStore) FindUserByEmail(ctx context.Context, email string) (*types.User, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where(&user{Type: principalUser}).
		Where("LOWER(principal_email) = ?", strings.ToLower(email))

	dst := new(user)
	if err := db.First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by email query failed")
	}

	return s.mapDBUser(dst), nil
}

// CreateUser saves the user details.
func (s *PrincipalOrmStore) CreateUser(ctx context.Context, user *types.User) error {
	dbUser, err := s.mapToDBUser(user)
	if err != nil {
		return fmt.Errorf("failed to map db user: %w", err)
	}

	if err = dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Create(&dbUser).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	user.ID = dbUser.ID
	return nil
}

// UpdateUser updates an existing user.
func (s *PrincipalOrmStore) UpdateUser(ctx context.Context, usr *types.User) error {
	dbUser, err := s.mapToDBUser(usr)
	if err != nil {
		return fmt.Errorf("failed to map db user: %w", err)
	}

	updateFields := []string{"principal_email", "principal_display_name", "principal_admin", "principal_blocked",
		"principal_salt", "principal_updated", "principal_user_password", "principal_user_source"}

	res := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Where(&user{Type: principalUser}).
		Select(updateFields).Updates(dbUser)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Update query failed")
	}

	return nil
}

// DeleteUser deletes the user.
func (s *PrincipalOrmStore) DeleteUser(ctx context.Context, id int64) error {
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Delete(&user{Type: principalUser}, id)
	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "The delete query failed")
	}

	return nil
}

// ListUsers returns a list of users.
func (s *PrincipalOrmStore) ListUsers(ctx context.Context, opts *types.UserFilter) ([]*types.User, error) {
	dst := []*user{}

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where(&user{Type: principalUser})
	stmt = stmt.Limit(int(database.Limit(opts.Size)))
	stmt = stmt.Offset(int(database.Offset(opts.Page, opts.Size)))

	order := opts.Order
	if order == enum.OrderDefault {
		order = enum.OrderAsc
	}

	switch opts.Sort {
	case enum.UserAttrName, enum.UserAttrNone:
		// NOTE: string concatenation is safe because the
		// order attribute is an enum and is not user-defined,
		// and is therefore not subject to injection attacks.
		stmt = stmt.Order("principal_display_name " + order.String())
	case enum.UserAttrCreated:
		stmt = stmt.Order("principal_created " + order.String())
	case enum.UserAttrUpdated:
		stmt = stmt.Order("principal_updated " + order.String())
	case enum.UserAttrEmail:
		stmt = stmt.Order("LOWER(principal_email) " + order.String())
	case enum.UserAttrUID:
		stmt = stmt.Order("principal_uid " + order.String())
	case enum.UserAttrAdmin:
		stmt = stmt.Order("principal_admin " + order.String())
	}

	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return s.mapDBUsers(dst), nil
}

// CountUsers returns a count of users matching the given filter.
func (s *PrincipalOrmStore) CountUsers(ctx context.Context, opts *types.UserFilter) (int64, error) {
	var count int64
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where(&user{Type: principalUser})

	if opts.Admin {
		stmt = stmt.Where("principal_admin = ?", opts.Admin)
	}

	if err := stmt.Count(&count).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}

	return count, nil
}

func (s *PrincipalOrmStore) mapDBUser(dbUser *user) *types.User {
	return &dbUser.User
}

func (s *PrincipalOrmStore) mapDBUsers(dbUsers []*user) []*types.User {
	res := make([]*types.User, len(dbUsers))
	for i := range dbUsers {
		res[i] = s.mapDBUser(dbUsers[i])
	}
	return res
}

func (s *PrincipalOrmStore) mapToDBUser(usr *types.User) (*user, error) {
	// user comes from outside.
	if usr == nil {
		return nil, fmt.Errorf("user is nil")
	}

	uidUnique, err := s.uidTransformation(usr.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to transform user UID: %w", err)
	}
	dbUser := &user{
		User:      *usr,
		Type:      principalUser,
		UIDUnique: uidUnique,
	}

	return dbUser, nil
}
