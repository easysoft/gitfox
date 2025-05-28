// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package principal

import (
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var _ store.PrincipalStore = (*PrincipalOrmStore)(nil)

type principalType string

const (
	principalUser           principalType = "user"
	principalService        principalType = "service"
	principalServiceAccount principalType = "serviceaccount"
)

// NewPrincipalOrmStore returns a new PrincipalStoreOrm.
func NewPrincipalOrmStore(db *gorm.DB, uidTransformation store.PrincipalUIDTransformation) *PrincipalOrmStore {
	return &PrincipalOrmStore{
		db:                db,
		uidTransformation: uidTransformation,
	}
}

// PrincipalOrmStore implements a PrincipalStore backed by a relational database.
type PrincipalOrmStore struct {
	db                *gorm.DB
	uidTransformation store.PrincipalUIDTransformation
}

const principalTable = "principals"

// principal is a DB representation of a principal.
// It is required to allow storing transformed UIDs used for uniquness constraints and searching.
type principal struct {
	types.Principal
	UIDUnique string `gorm:"column:principal_uid_unique"`
}

// Find finds the principal by id.
func (s *PrincipalOrmStore) Find(ctx context.Context, id int64) (*types.Principal, error) {
	dst := new(principal)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).First(&dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by id query failed")
	}

	return s.mapDBPrincipal(dst), nil
}

// FindByUID finds the principal by uid.
func (s *PrincipalOrmStore) FindByUID(ctx context.Context, uid string) (*types.Principal, error) {
	// map the UID to unique UID before searching!
	uidUnique, err := s.uidTransformation(uid)
	if err != nil {
		// in case we fail to transform, return a not found (as it can't exist in the first place)
		log.Ctx(ctx).Debug().Msgf("failed to transform uid '%s': %s", uid, err.Error())
		return nil, gitfox_store.ErrResourceNotFound
	}

	dst := new(principal)
	if err = dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).First(&dst, &principal{UIDUnique: uidUnique}).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by uid query failed")
	}

	return s.mapDBPrincipal(dst), nil
}

// FindManyByUID returns all principals found for the provided UIDs.
// If a UID isn't found, it's not returned in the list.
func (s *PrincipalOrmStore) FindManyByUID(ctx context.Context, uids []string) ([]*types.Principal, error) {
	// map the UIDs to unique UIDs before searching!
	uniqueUIDs := make([]string, len(uids))
	for i := range uids {
		var err error
		uniqueUIDs[i], err = s.uidTransformation(uids[i])
		if err != nil {
			// in case we fail to transform, skip the entry (as it can't exist in the first place)
			log.Ctx(ctx).Warn().Msgf("failed to transform uid '%s': %s", uids[i], err.Error())
		}
	}

	var dst []*principal
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where("principal_uid_unique IN ?", uids).Find(&dst)
	if res.Error != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, res.Error, "find many by uid for principals query failed")
	}

	if res.RowsAffected == 0 {
		return []*types.Principal{}, nil
	}

	return s.mapDBPrincipals(dst), nil
}

// FindByEmail finds the principal by email.
func (s *PrincipalOrmStore) FindByEmail(ctx context.Context, email string) (*types.Principal, error) {
	dst := new(principal)
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Where("LOWER(principal_email) = ?", email)
	if err := db.First(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select by email query failed")
	}

	return s.mapDBPrincipal(dst), nil
}

// List lists the principals matching the provided filter.
func (s *PrincipalOrmStore) List(ctx context.Context, opts *types.PrincipalFilter) ([]*types.Principal, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable)

	if len(opts.Types) == 1 {
		db = db.Where("principal_type = ?", opts.Types[0])
	} else if len(opts.Types) > 1 {
		db = db.Where("principal_type IN ?", opts.Types)
	}

	if opts.Query != "" {
		// TODO: optimize performance
		// https://harness.atlassian.net/browse/CODE-522
		searchTerm := fmt.Sprintf("%%%s%%", strings.ToLower(opts.Query))
		db = db.Where(
			"(LOWER(principal_uid) LIKE ? OR LOWER(principal_email) LIKE ? OR LOWER(principal_display_name) LIKE ?)",
			searchTerm,
			searchTerm,
			searchTerm,
		)
	}

	db = db.Limit(int(database.Limit(opts.Size)))
	db = db.Offset(int(database.Offset(opts.Page, opts.Size)))

	dst := []*principal{}
	if err := db.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Search by display_name and email query failed")
	}

	return s.mapDBPrincipals(dst), nil
}

func (s *PrincipalOrmStore) mapDBPrincipal(dbPrincipal *principal) *types.Principal {
	return &dbPrincipal.Principal
}

func (s *PrincipalOrmStore) mapDBPrincipals(dbPrincipals []*principal) []*types.Principal {
	res := make([]*types.Principal, len(dbPrincipals))
	for i := range dbPrincipals {
		res[i] = s.mapDBPrincipal(dbPrincipals[i])
	}
	return res
}
