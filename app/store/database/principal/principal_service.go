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

	"github.com/rs/zerolog/log"
)

// service is a DB representation of a service principal.
// It is required to allow storing transformed UIDs used for uniquness constraints and searching.
type service struct {
	types.Service
	Type      principalType `gorm:"column:principal_type"`
	UIDUnique string        `gorm:"column:principal_uid_unique"`
}

// FindService finds the service by id.
func (s *PrincipalOrmStore) FindService(ctx context.Context, id int64) (*types.Service, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where(&service{Type: principalService})

	dst := new(service)
	if err := db.First(&dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by id query failed")
	}

	return s.mapDBService(dst), nil
}

// FindServiceByUID finds the service by uid.
func (s *PrincipalOrmStore) FindServiceByUID(ctx context.Context, uid string) (*types.Service, error) {
	// map the UID to unique UID before searching!
	uidUnique, err := s.uidTransformation(uid)
	if err != nil {
		// in case we fail to transform, return a not found (as it can't exist in the first place)
		log.Ctx(ctx).Debug().Msgf("failed to transform uid '%s': %s", uid, err.Error())
		return nil, gitfox_store.ErrResourceNotFound
	}

	dst := new(service)
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where(&service{Type: principalService, UIDUnique: uidUnique})
	if err = db.First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by uid query failed")
	}

	return s.mapDBService(dst), nil
}

// CreateService saves the service.
func (s *PrincipalOrmStore) CreateService(ctx context.Context, svc *types.Service) error {
	dbSVC, err := s.mapToDBservice(svc)
	if err != nil {
		return fmt.Errorf("failed to map db service: %w", err)
	}

	if err = dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Create(&dbSVC).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	svc.ID = dbSVC.ID
	return nil
}

// UpdateService updates the service.
func (s *PrincipalOrmStore) UpdateService(ctx context.Context, svc *types.Service) error {
	dbSVC, err := s.mapToDBservice(svc)
	if err != nil {
		return fmt.Errorf("failed to map db service: %w", err)
	}

	res := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Where(&service{Type: principalService, Service: types.Service{ID: svc.ID}}).
		Select("Email", "DisplayName", "Admin", "Blocked", "Updated").Updates(dbSVC)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Update query failed")
	}

	return err
}

// DeleteService deletes the service.
func (s *PrincipalOrmStore) DeleteService(ctx context.Context, id int64) error {
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).
		Delete(&service{Type: principalService, Service: types.Service{ID: id}})

	// delete the service
	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "The delete query failed")
	}

	return nil
}

// ListServices returns a list of service for a specific parent.
func (s *PrincipalOrmStore) ListServices(ctx context.Context) ([]*types.Service, error) {
	dst := []*service{}

	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where(&service{Type: principalService}).
		Order("principal_uid ASC")

	err := db.Scan(&dst).Error
	if err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing default list query")
	}

	return s.mapDBServices(dst), nil
}

// CountServices returns a count of service for a specific parent.
func (s *PrincipalOrmStore) CountServices(ctx context.Context) (int64, error) {
	var count int64
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where(&service{Type: principalService})

	if err := stmt.Count(&count).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}

	return count, nil
}

func (s *PrincipalOrmStore) mapDBService(dbSvc *service) *types.Service {
	return &dbSvc.Service
}

func (s *PrincipalOrmStore) mapDBServices(dbSVCs []*service) []*types.Service {
	res := make([]*types.Service, len(dbSVCs))
	for i := range dbSVCs {
		res[i] = s.mapDBService(dbSVCs[i])
	}
	return res
}

func (s *PrincipalOrmStore) mapToDBservice(svc *types.Service) (*service, error) {
	// service comes from outside.
	if svc == nil {
		return nil, fmt.Errorf("service is nil")
	}

	uidUnique, err := s.uidTransformation(svc.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to transform service UID: %w", err)
	}
	dbService := &service{
		Service:   *svc,
		Type:      principalService,
		UIDUnique: uidUnique,
	}

	return dbService, nil
}
