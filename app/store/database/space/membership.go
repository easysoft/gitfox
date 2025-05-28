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
	"github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.MembershipStore = (*MembershipStore)(nil)

// NewMembershipOrmStore returns a new MembershipStore.
func NewMembershipOrmStore(
	db *gorm.DB,
	pCache store.PrincipalInfoCache,
	spacePathStore store.SpacePathStore,
	spaceStore store.SpaceStore,
) *MembershipStore {
	return &MembershipStore{
		db:             db,
		pCache:         pCache,
		spacePathStore: spacePathStore,
		spaceStore:     spaceStore,
	}
}

// MembershipStore implements store.MembershipStore backed by a relational database.
type MembershipStore struct {
	db             *gorm.DB
	pCache         store.PrincipalInfoCache
	spacePathStore store.SpacePathStore
	spaceStore     store.SpaceStore
}

type membership struct {
	SpaceID     int64 `db:"membership_space_id"     gorm:"column:membership_space_id"`
	PrincipalID int64 `db:"membership_principal_id" gorm:"column:membership_principal_id"`

	CreatedBy int64 `db:"membership_created_by" gorm:"column:membership_created_by"`
	Created   int64 `db:"membership_created"    gorm:"column:membership_created"`
	Updated   int64 `db:"membership_updated"    gorm:"column:membership_updated"`

	Role enum.MembershipRole `db:"membership_role" gorm:"column:membership_role"`
}

type membershipPrincipal struct {
	MemberShip membership     `gorm:"embedded"`
	Info       principal.Info `gorm:"embedded"`
}

type membershipSpace struct {
	MemberShip membership `gorm:"embedded"`
	Space      space      `gorm:"embedded"`
}

const (
	tableMember = "memberships"
)

// Find finds the membership by space id and principal id.
func (s *MembershipStore) Find(ctx context.Context, key types.MembershipKey) (*types.Membership, error) {
	q := membership{SpaceID: key.SpaceID, PrincipalID: key.PrincipalID}
	dst := &membership{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableMember).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find membership")
	}
	result := mapToMembership(dst)

	return &result, nil
}

func (s *MembershipStore) FindUser(ctx context.Context, key types.MembershipKey) (*types.MembershipUser, error) {
	m, err := s.Find(ctx, key)
	if err != nil {
		return nil, err
	}

	result, err := s.addPrincipalInfos(ctx, m)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Create creates a new membership.
func (s *MembershipStore) Create(ctx context.Context, membership *types.Membership) error {
	dbObj := mapToInternalMembership(membership)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableMember).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to insert membership")
	}

	return nil
}

// Update updates the role of a member of a space.
func (s *MembershipStore) Update(ctx context.Context, m *types.Membership) error {
	dbMembership := mapToInternalMembership(m)
	dbMembership.Updated = time.Now().UnixMilli()

	updateFields := []string{"Updated", "Role"}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableMember).
		Where(&membership{SpaceID: m.SpaceID, PrincipalID: m.PrincipalID}).
		Select(updateFields).Updates(&dbMembership)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update secret")
	}

	m.Updated = dbMembership.Updated

	return nil
}

// Delete deletes the membership.
func (s *MembershipStore) Delete(ctx context.Context, key types.MembershipKey) error {
	q := membership{SpaceID: key.SpaceID, PrincipalID: key.PrincipalID}

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableMember).Where(&q).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "delete membership query failed")
	}
	return nil
}

// CountUsers returns a number of users memberships that matches the provided filter.
func (s *MembershipStore) CountUsers(ctx context.Context,
	spaceID int64,
	filter types.MembershipUserFilter,
) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableMember).
		Joins("join principals ON membership_principal_id = principal_id").
		Where("membership_space_id = ?", spaceID)

	stmt = applyMembershipUserFilter(stmt, filter)

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing membership users count query")
	}

	return count, nil
}

// ListUsers returns a list of memberships for a space or a user.
func (s *MembershipStore) ListUsers(ctx context.Context,
	spaceID int64,
	filter types.MembershipUserFilter,
) ([]types.MembershipUser, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableMember).
		Joins("join principals ON membership_principal_id = principal_id").
		Where("membership_space_id = ?", spaceID)

	stmt = applyMembershipUserFilter(stmt, filter)
	stmt = stmt.Limit(int(database.Limit(filter.Size)))
	stmt = stmt.Offset(int(database.Offset(filter.Page, filter.Size)))

	order := filter.Order
	if order == enum.OrderDefault {
		order = enum.OrderAsc
	}

	switch filter.Sort {
	case enum.MembershipUserSortName:
		stmt = stmt.Order("principal_display_name " + order.String())
	case enum.MembershipUserSortCreated:
		stmt = stmt.Order("membership_created " + order.String())
	}

	dst := make([]*membershipPrincipal, 0)

	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing membership users list query")
	}

	result, err := s.mapToMembershipUsers(ctx, dst)
	if err != nil {
		return nil, fmt.Errorf("failed to map memberships users to external type: %w", err)
	}

	return result, nil
}

func applyMembershipUserFilter(
	stmt *gorm.DB,
	opts types.MembershipUserFilter,
) *gorm.DB {
	if opts.Query != "" {
		searchTerm := "%%" + strings.ToLower(opts.Query) + "%%"
		stmt = stmt.Where("LOWER(principal_display_name) LIKE ?", searchTerm)
	}

	return stmt
}

func (s *MembershipStore) CountSpaces(ctx context.Context,
	userID int64,
	filter types.MembershipSpaceFilter,
) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableMember).
		Joins("join spaces ON spaces.space_id = membership_space_id").
		Where("membership_principal_id = ? AND spaces.space_deleted IS NULL", userID)

	stmt = applyMembershipSpaceFilter(stmt, filter)

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing membership spaces count query")
	}

	return count, nil
}

// ListSpaces returns a list of spaces in which the provided user is a member.
func (s *MembershipStore) ListSpaces(ctx context.Context,
	userID int64,
	filter types.MembershipSpaceFilter,
) ([]types.MembershipSpace, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableMember).
		Joins("join spaces ON spaces.space_id = membership_space_id").
		Where("membership_principal_id = ? AND spaces.space_deleted IS NULL", userID)

	stmt = applyMembershipSpaceFilter(stmt, filter)
	stmt = stmt.Limit(int(database.Limit(filter.Size)))
	stmt = stmt.Offset(int(database.Offset(filter.Page, filter.Size)))

	order := filter.Order
	if order == enum.OrderDefault {
		order = enum.OrderAsc
	}

	switch filter.Sort {
	// TODO [CODE-1363]: remove after identifier migration.
	case enum.MembershipSpaceSortUID, enum.MembershipSpaceSortIdentifier:
		stmt = stmt.Order("space_uid " + order.String())
	case enum.MembershipSpaceSortCreated:
		stmt = stmt.Order("membership_created " + order.String())
	}

	dst := make([]*membershipSpace, 0)
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	result, err := s.mapToMembershipSpaces(ctx, dst)
	if err != nil {
		return nil, fmt.Errorf("failed to map memberships spaces to external type: %w", err)
	}

	return result, nil
}

func applyMembershipSpaceFilter(
	stmt *gorm.DB,
	opts types.MembershipSpaceFilter,
) *gorm.DB {
	if opts.Query != "" {
		searchTerm := "%%" + strings.ToLower(opts.Query) + "%%"
		stmt = stmt.Where("LOWER(space_uid) LIKE ?", searchTerm)
	}

	return stmt
}

func mapToMembership(m *membership) types.Membership {
	return types.Membership{
		MembershipKey: types.MembershipKey{
			SpaceID:     m.SpaceID,
			PrincipalID: m.PrincipalID,
		},
		CreatedBy: m.CreatedBy,
		Created:   m.Created,
		Updated:   m.Updated,
		Role:      m.Role,
	}
}

func mapToInternalMembership(m *types.Membership) membership {
	return membership{
		SpaceID:     m.SpaceID,
		PrincipalID: m.PrincipalID,
		CreatedBy:   m.CreatedBy,
		Created:     m.Created,
		Updated:     m.Updated,
		Role:        m.Role,
	}
}

func (s *MembershipStore) addPrincipalInfos(ctx context.Context, m *types.Membership) (types.MembershipUser, error) {
	var result types.MembershipUser

	// pull principal infos from cache
	infoMap, err := s.pCache.Map(ctx, []int64{m.CreatedBy, m.PrincipalID})
	if err != nil {
		return result, fmt.Errorf("failed to load membership principal infos: %w", err)
	}

	if user, ok := infoMap[m.PrincipalID]; ok {
		result.Principal = *user
	} else {
		return result, fmt.Errorf("failed to find membership principal info: %w", err)
	}

	if addedBy, ok := infoMap[m.CreatedBy]; ok {
		result.AddedBy = *addedBy
	}

	result.Membership = *m

	return result, nil
}

func (s *MembershipStore) mapToMembershipUsers(ctx context.Context,
	ms []*membershipPrincipal,
) ([]types.MembershipUser, error) {
	// collect all principal IDs
	ids := make([]int64, 0, len(ms))
	for _, m := range ms {
		ids = append(ids, m.MemberShip.CreatedBy)
	}

	// pull principal infos from cache
	infoMap, err := s.pCache.Map(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load membership principal infos: %w", err)
	}

	// attach the principal infos back to the slice items
	res := make([]types.MembershipUser, len(ms))
	for i := range ms {
		m := ms[i]
		res[i].Membership = mapToMembership(&m.MemberShip)
		res[i].Principal = principal.MapToInfo(&m.Info)
		if addedBy, ok := infoMap[m.MemberShip.CreatedBy]; ok {
			res[i].AddedBy = *addedBy
		}
	}

	return res, nil
}

func (s *MembershipStore) mapToMembershipSpaces(ctx context.Context,
	ms []*membershipSpace,
) ([]types.MembershipSpace, error) {
	// collect all principal IDs
	ids := make([]int64, 0, len(ms))
	for _, m := range ms {
		ids = append(ids, m.MemberShip.CreatedBy)
	}

	// pull principal infos from cache
	infoMap, err := s.pCache.Map(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load membership principal infos: %w", err)
	}

	// attach the principal infos back to the slice items
	res := make([]types.MembershipSpace, len(ms))
	for i := range ms {
		m := ms[i]
		res[i].Membership = mapToMembership(&m.MemberShip)
		space, err := mapToSpace(ctx, s.db, s.spacePathStore, &m.Space)
		if err != nil {
			return nil, fmt.Errorf("faild to map space %d: %w", m.Space.ID, err)
		}
		res[i].Space = *space
		if addedBy, ok := infoMap[m.MemberShip.CreatedBy]; ok {
			res[i].AddedBy = *addedBy
		}
	}

	return res, nil
}

func (s *MembershipStore) ListUserSpaces(ctx context.Context,
	userID int64,
) ([]int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableMember).
		Select("membership_space_id").
		Where("membership_principal_id = ?", userID)
	dst := make([]int64, 0)
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}
	return dst, nil
}
