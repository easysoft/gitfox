// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package space_test

import (
	"context"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/cache"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/app/store/database/space"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	cache2 "github.com/easysoft/gitfox/cache"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableMembership = "memberships"
)

type MemberShipSuite struct {
	testsuite.BaseSuite

	ormStore  *space.MembershipStore
	sqlxStore *database.MembershipStore
}

func TestMemberShipSuite(t *testing.T) {
	ctx := context.Background()

	st := &MemberShipSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "memberships",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		spacePathTransformation := store.ToLowerSpacePathTransformation

		ormPathStore := space.NewSpacePathOrmStore(st.Gdb, store.ToLowerSpacePathTransformation)
		sqlxPathStore := database.NewSpacePathStore(st.Sdb, store.ToLowerSpacePathTransformation)

		ormPathCache := cache.New(ormPathStore, spacePathTransformation)
		sqlxPathCache := cache.New(sqlxPathStore, spacePathTransformation)

		ormSpaceStore := space.NewSpaceOrmStore(st.Gdb, ormPathCache, ormPathStore)
		sqlxSpaceStore := database.NewSpaceStore(st.Sdb, sqlxPathCache, sqlxPathStore)

		ormPrincipalView := principal.NewPrincipalOrmInfoView(st.Gdb)
		ormPCache := cache2.NewExtended[int64, *types.PrincipalInfo](ormPrincipalView, 30*time.Second)

		sqlxPrincipalView := database.NewPrincipalInfoView(st.Sdb)
		sqlxPCache := cache2.NewExtended[int64, *types.PrincipalInfo](sqlxPrincipalView, 30*time.Second)

		st.ormStore = space.NewMembershipOrmStore(st.Gdb, ormPCache, ormPathStore, ormSpaceStore)
		st.sqlxStore = database.NewMembershipStore(st.Sdb, sqlxPCache, sqlxPathStore, sqlxSpaceStore)

		// add init data
		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 3, false)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 4, false)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 5, true)

		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 2, 1)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 3, 1)
	}

	suite.Run(t, st)
}

func (suite *MemberShipSuite) SetupTest() {
	suite.addData()
}

func (suite *MemberShipSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableMembership).Where("1 = 1").Delete(nil)
}

var testAddItems = []struct {
	spaceId     int64
	principalId int64
	createdBy   int64
	role        enum.MembershipRole
}{
	{spaceId: 1, principalId: 1, createdBy: 1, role: enum.MembershipRoleSpaceOwner},
	{spaceId: 1, principalId: 2, createdBy: 1, role: enum.MembershipRoleContributor},
	{spaceId: 1, principalId: 3, createdBy: 1, role: enum.MembershipRoleExecutor},
	{spaceId: 2, principalId: 2, createdBy: 1, role: enum.MembershipRoleContributor},
	{spaceId: 2, principalId: 3, createdBy: 2, role: enum.MembershipRoleReader},
	{spaceId: 2, principalId: 5, createdBy: 2, role: enum.MembershipRoleReader},
	{spaceId: 3, principalId: 3, createdBy: 1, role: enum.MembershipRoleSpaceOwner},
	{spaceId: 3, principalId: 1, createdBy: 3, role: enum.MembershipRoleContributor},
	{spaceId: 3, principalId: 2, createdBy: 1, role: enum.MembershipRoleExecutor},
	{spaceId: 3, principalId: 4, createdBy: 3, role: enum.MembershipRoleReader},
}

func (suite *MemberShipSuite) addData() {
	now := time.Now().UnixMilli()
	for id, item := range testAddItems {
		obj := types.Membership{
			MembershipKey: types.MembershipKey{
				SpaceID:     item.spaceId,
				PrincipalID: item.principalId,
			},
			CreatedBy: item.createdBy,
			Created:   now,
			Updated:   now,
			Role:      item.role,
		}

		err := suite.ormStore.Create(suite.Ctx, &obj)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *MemberShipSuite) TestFind() {
	for id, test := range testAddItems {
		key := types.MembershipKey{SpaceID: test.spaceId, PrincipalID: test.principalId}
		obj, err := suite.ormStore.Find(suite.Ctx, key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.role, obj.Role)

		objB, err := suite.sqlxStore.Find(suite.Ctx, key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB)
	}
}

func (suite *MemberShipSuite) TestFindUser() {
	for id, test := range testAddItems {
		key := types.MembershipKey{SpaceID: test.spaceId, PrincipalID: test.principalId}
		obj, err := suite.ormStore.FindUser(suite.Ctx, key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.createdBy, obj.AddedBy.ID)
		require.Equal(suite.T(), test.principalId, obj.Principal.ID)

		objB, err := suite.sqlxStore.FindUser(suite.Ctx, key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB)
	}
}

func (suite *MemberShipSuite) TestListAndCountUser() {
	tests := []struct {
		spaceID    int64
		filter     types.MembershipUserFilter
		wantLength int64
	}{
		{spaceID: 1, filter: types.MembershipUserFilter{}, wantLength: 3},
		{spaceID: 2, filter: types.MembershipUserFilter{}, wantLength: 3},
		{spaceID: 3, filter: types.MembershipUserFilter{}, wantLength: 4},
		{spaceID: 3, filter: types.MembershipUserFilter{Sort: enum.MembershipUserSortName}, wantLength: 4},
		{spaceID: 3, filter: types.MembershipUserFilter{Sort: enum.MembershipUserSortName, Order: enum.OrderDesc}, wantLength: 4},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.ListUsers(suite.Ctx, test.spaceID, test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.CountUsers(suite.Ctx, test.spaceID, test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.ListUsers(suite.Ctx, test.spaceID, test.filter)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *MemberShipSuite) TestListAndCountSpace() {
	tests := []struct {
		userID     int64
		filter     types.MembershipSpaceFilter
		wantLength int64
	}{
		{userID: 1, filter: types.MembershipSpaceFilter{}, wantLength: 2},
		{userID: 2, filter: types.MembershipSpaceFilter{}, wantLength: 3},
		{userID: 3, filter: types.MembershipSpaceFilter{}, wantLength: 3},
		{userID: 4, filter: types.MembershipSpaceFilter{}, wantLength: 1},
		{userID: 5, filter: types.MembershipSpaceFilter{}, wantLength: 1},
		{userID: 3, filter: types.MembershipSpaceFilter{Sort: enum.MembershipSpaceSortIdentifier}, wantLength: 3},
		{userID: 3, filter: types.MembershipSpaceFilter{Sort: enum.MembershipSpaceSortIdentifier, Order: enum.OrderDesc}, wantLength: 3},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.ListSpaces(suite.Ctx, test.userID, test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.CountSpaces(suite.Ctx, test.userID, test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.ListSpaces(suite.Ctx, test.userID, test.filter)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *MemberShipSuite) TestDelete() {
	e := suite.ormStore.Delete(suite.Ctx, types.MembershipKey{SpaceID: 1, PrincipalID: 3})
	require.NoError(suite.T(), e)
	e = suite.sqlxStore.Delete(suite.Ctx, types.MembershipKey{SpaceID: 2, PrincipalID: 3})
	require.NoError(suite.T(), e)

	tests := []struct {
		spaceID    int64
		filter     types.MembershipUserFilter
		wantLength int64
	}{
		{spaceID: 1, filter: types.MembershipUserFilter{}, wantLength: 2},
		{spaceID: 2, filter: types.MembershipUserFilter{}, wantLength: 2},
	}

	for id, test := range tests {
		count, err := suite.ormStore.CountUsers(suite.Ctx, test.spaceID, test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *MemberShipSuite) TestUpdate() {
	now := time.Now().UnixMilli() - 1

	tests := []struct {
		spaceID     int64
		principalID int64
		role        enum.MembershipRole
		createdBy   int64
	}{
		{spaceID: 1, principalID: 2, role: enum.MembershipRoleExecutor, createdBy: 1},
		{spaceID: 2, principalID: 3, role: enum.MembershipRoleContributor, createdBy: 2},
	}

	ups := make([]types.Membership, 0)
	for _, test := range tests {
		up := types.Membership{
			MembershipKey: types.MembershipKey{SpaceID: test.spaceID, PrincipalID: test.principalID},
			Updated:       now,
			Role:          test.role,
		}
		ups = append(ups, up)
	}

	e := suite.ormStore.Update(suite.Ctx, &ups[0])
	require.NoError(suite.T(), e)
	e = suite.sqlxStore.Update(suite.Ctx, &ups[1])
	require.NoError(suite.T(), e)

	for id, test := range tests {
		key := types.MembershipKey{SpaceID: test.spaceID, PrincipalID: test.principalID}
		obj, err := suite.ormStore.Find(suite.Ctx, key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.role, obj.Role)
		require.Equal(suite.T(), test.createdBy, obj.CreatedBy)
		require.Greater(suite.T(), obj.Updated, now)

		objB, err := suite.sqlxStore.Find(suite.Ctx, key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB)
	}
}
