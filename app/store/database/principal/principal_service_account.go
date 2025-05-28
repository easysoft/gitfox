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

	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
)

// serviceAccount is a DB representation of a service account principal.
// It is required to allow storing transformed UIDs used for uniquness constraints and searching.
type serviceAccount struct {
	types.ServiceAccount
	Type      principalType `gorm:"column:principal_type"`
	UIDUnique string        `gorm:"column:principal_uid_unique"`
}

// FindServiceAccount finds the service account by id.
func (s *PrincipalOrmStore) FindServiceAccount(ctx context.Context, id int64) (*types.ServiceAccount, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Where(&serviceAccount{Type: principalServiceAccount})

	dst := new(serviceAccount)
	if err := db.First(&dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by id query failed")
	}
	return s.mapDBServiceAccount(dst), nil
}

// FindServiceAccountByUID finds the service account by uid.
func (s *PrincipalOrmStore) FindServiceAccountByUID(ctx context.Context, uid string) (*types.ServiceAccount, error) {
	// map the UID to unique UID before searching!
	uidUnique, err := s.uidTransformation(uid)
	if err != nil {
		// in case we fail to transform, return a not found (as it can't exist in the first place)
		log.Ctx(ctx).Debug().Msgf("failed to transform uid '%s': %s", uid, err.Error())
		return nil, gitfox_store.ErrResourceNotFound
	}

	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Where(&serviceAccount{Type: principalServiceAccount, UIDUnique: uidUnique})

	dst := new(serviceAccount)
	if err = db.First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by uid query failed")
	}

	return s.mapDBServiceAccount(dst), nil
}

// CreateServiceAccount saves the service account.
func (s *PrincipalOrmStore) CreateServiceAccount(ctx context.Context, sa *types.ServiceAccount) error {
	dbSA, err := s.mapToDBserviceAccount(sa)
	if err != nil {
		return fmt.Errorf("failed to map db service account: %w", err)
	}

	if err = dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Create(dbSA).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	sa.ID = dbSA.ID
	return nil
}

// UpdateServiceAccount updates the service account details.
func (s *PrincipalOrmStore) UpdateServiceAccount(ctx context.Context, sa *types.ServiceAccount) error {
	dbSA, err := s.mapToDBserviceAccount(sa)
	if err != nil {
		return fmt.Errorf("failed to map db service account: %w", err)
	}

	res := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Where(&serviceAccount{Type: principalServiceAccount}).
		Select("Email", "DisplayName", "Blocked", "Salt", "Updated").Updates(dbSA)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Update query failed")
	}

	return err
}

// DeleteServiceAccount deletes the service account.
func (s *PrincipalOrmStore) DeleteServiceAccount(ctx context.Context, id int64) error {
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Delete(&serviceAccount{Type: principalServiceAccount}, id)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "The delete query failed")
	}

	return nil
}

// ListServiceAccounts returns a list of service accounts for a specific parent.
func (s *PrincipalOrmStore) ListServiceAccounts(ctx context.Context, parentType enum.ParentResourceType,
	parentID int64) ([]*types.ServiceAccount, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Where(&serviceAccount{Type: principalServiceAccount, ServiceAccount: types.ServiceAccount{
			ParentType: parentType, ParentID: parentID,
		}}).Order("principal_uid ASC")

	dst := []*serviceAccount{}
	err := db.Scan(&dst).Error
	if err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing default list query")
	}

	return s.mapDBServiceAccounts(dst), nil
}

// CountServiceAccounts returns a count of service accounts for a specific parent.
func (s *PrincipalOrmStore) CountServiceAccounts(ctx context.Context,
	parentType enum.ParentResourceType, parentID int64) (int64, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Where(&serviceAccount{Type: principalServiceAccount, ServiceAccount: types.ServiceAccount{
			ParentType: parentType, ParentID: parentID,
		}})

	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}

	return count, nil
}

func (s *PrincipalOrmStore) mapDBServiceAccount(dbSA *serviceAccount) *types.ServiceAccount {
	return &dbSA.ServiceAccount
}

func (s *PrincipalOrmStore) mapDBServiceAccounts(dbSAs []*serviceAccount) []*types.ServiceAccount {
	res := make([]*types.ServiceAccount, len(dbSAs))
	for i := range dbSAs {
		res[i] = s.mapDBServiceAccount(dbSAs[i])
	}
	return res
}

func (s *PrincipalOrmStore) mapToDBserviceAccount(sa *types.ServiceAccount) (*serviceAccount, error) {
	// service account comes from outside.
	if sa == nil {
		return nil, fmt.Errorf("service account is nil")
	}

	uidUnique, err := s.uidTransformation(sa.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to transform service account UID: %w", err)
	}
	dbSA := &serviceAccount{
		ServiceAccount: *sa,
		Type:           principalServiceAccount,
		UIDUnique:      uidUnique,
	}

	return dbSA, nil
}
