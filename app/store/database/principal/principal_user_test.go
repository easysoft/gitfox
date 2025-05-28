// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package principal_test

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const emailSuffix = "@local.dev"

type PrincipalUserSuite struct {
	testsuite.BaseSuite

	ormStore  *principal.PrincipalOrmStore
	sqlxStore *database.PrincipalStore
}

func TestPrincipalUserSuite(t *testing.T) {
	ctx := context.Background()

	st := &PrincipalUserSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "principals_user",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = principal.NewPrincipalOrmStore(st.Gdb, store.ProvidePrincipalUIDTransformation())
		st.sqlxStore = database.NewPrincipalStore(st.Sdb, store.ProvidePrincipalUIDTransformation())
	}

	suite.Run(t, st)
}

func (suite *PrincipalUserSuite) SetupTest() {
	suite.addData()
}

func (suite *PrincipalUserSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table("principals").Where("1 = 1").Delete(nil)
}

func (suite *PrincipalUserSuite) addData() {
	addItems := []types.User{
		{ID: 1, UID: "admin", Email: "admin" + emailSuffix, Admin: true},
		{ID: 2, UID: "user1", Email: "user1" + emailSuffix, Admin: false},
		{ID: 3, UID: "user2", Email: "user2" + emailSuffix, Admin: false},
		{ID: 4, UID: "admin2", Email: "admin2" + emailSuffix, Admin: true},
	}

	for id, item := range addItems {
		err := suite.ormStore.CreateUser(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PrincipalUserSuite) TestCreateUser() {
	err := suite.ormStore.CreateUser(suite.Ctx, &types.User{
		UID: "user_49", Email: "user_49" + emailSuffix,
	})
	require.NoError(suite.T(), err)
	testsuite.AddUser(suite.Ctx, suite.T(), suite.ormStore, 50, true)
}

func (suite *PrincipalUserSuite) TestFindUser() {
	tests := []struct {
		ID        int64
		UID       string
		expectErr bool
	}{
		{ID: 1, UID: "admin"}, {ID: 2, UID: "user1"},
		{ID: 3, UID: "user2"}, {ID: 4, UID: "admin2"},

		// invalid test
		{ID: 11, expectErr: true},
	}

	for id, test := range tests {
		user, err := suite.ormStore.FindUser(suite.Ctx, test.ID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.UID, user.UID, testsuite.InvalidLoopMsgF, id)
		}

		// check same
		userB, err := suite.sqlxStore.FindUser(suite.Ctx, test.ID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.EqualExportedValues(suite.T(), *user, *userB, testsuite.InvalidLoopMsgF, id)
		}
	}
}

func (suite *PrincipalUserSuite) TestFindUserByUID() {
	tests := []struct {
		ID        int64
		UID       string
		expectErr bool
	}{
		{ID: 1, UID: "admin"}, {ID: 2, UID: "user1"},
		{ID: 3, UID: "user2"}, {ID: 4, UID: "admin2"},

		// invalid test
		{ID: 11, UID: "user11", expectErr: true},
	}

	for id, test := range tests {
		user, err := suite.ormStore.FindUserByUID(suite.Ctx, test.UID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.ID, user.ID, testsuite.InvalidLoopMsgF, id)
		}

		// check same
		userB, err := suite.sqlxStore.FindUserByUID(suite.Ctx, test.UID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.EqualExportedValues(suite.T(), *user, *userB, testsuite.InvalidLoopMsgF, id)
		}
	}
}

func (suite *PrincipalUserSuite) TestFindUserByEmail() {
	tests := []struct {
		ID        int64
		Email     string
		expectErr bool
	}{
		{ID: 1, Email: "admin" + emailSuffix}, {ID: 2, Email: "user1" + emailSuffix},
		{ID: 3, Email: "user2" + emailSuffix}, {ID: 4, Email: "admin2" + emailSuffix},

		// invalid test
		{ID: 11, Email: "user11" + emailSuffix, expectErr: true},
		{ID: 11, Email: "user1" + emailSuffix + "v", expectErr: true},
	}

	for id, test := range tests {
		user, err := suite.ormStore.FindUserByEmail(suite.Ctx, test.Email)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.ID, user.ID, testsuite.InvalidLoopMsgF, id)
		}

		// check same
		userB, err := suite.sqlxStore.FindUserByEmail(suite.Ctx, test.Email)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.EqualExportedValues(suite.T(), *user, *userB, testsuite.InvalidLoopMsgF, id)
		}
	}
}

func (suite *PrincipalUserSuite) TestUpdateUser() {
	upEmailSuffix := "@local.test"
	updates := []types.User{
		{ID: 1, Email: "admin" + upEmailSuffix, Salt: "salt_admin", Password: "pwd_admin"},
		{ID: 2, DisplayName: "User Name 1"},                       // set email to blank
		{ID: 3, UID: "user2_add", Email: "user2" + upEmailSuffix}, // uid won't be changed
	}

	// sqlx updateUser method must provide a fulfilled User object
	updates2 := []types.User{
		{ID: 4, UID: "admin2_add", Email: "admin2" + upEmailSuffix}, // uid won't be changed
	}

	updatesInvalids := []types.User{
		{ID: 1, Email: "admin2" + upEmailSuffix}, // duplicate email
		{ID: 1, DisplayName: "Admin"},            // duplicate blank email
	}

	for id, up := range updates {
		err := suite.ormStore.UpdateUser(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, up := range updates2 {
		err := suite.sqlxStore.UpdateUser(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, up := range updatesInvalids {
		err := suite.ormStore.UpdateUser(suite.Ctx, &up)
		require.Error(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		err = suite.sqlxStore.UpdateUser(suite.Ctx, &up)
		require.Error(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	tests := []struct {
		ID        int64
		fieldName string
		expect    interface{}
	}{
		{ID: 1, expect: "admin" + upEmailSuffix, fieldName: "Email"},
		{ID: 2, expect: "User Name 1", fieldName: "DisplayName"},
		{ID: 3, expect: "user2", fieldName: "UID"},
		{ID: 4, expect: "admin2", fieldName: "UID"},
		{ID: 4, expect: "admin2" + upEmailSuffix, fieldName: "Email"},
	}

	for id, test := range tests {
		user, err := suite.ormStore.FindUser(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		testsuite.EqualFieldValue(suite.T(), test.expect, user, test.fieldName, testsuite.InvalidLoopMsgF, id)

		userB, err := suite.sqlxStore.FindUser(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		testsuite.EqualFieldValue(suite.T(), test.expect, userB, test.fieldName, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PrincipalUserSuite) TestDeleteUser() {
	err := suite.ormStore.DeleteUser(suite.Ctx, 1)
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.DeleteUser(suite.Ctx, 2)
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.FindUser(suite.Ctx, 2)
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.FindUser(suite.Ctx, 1)
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.DeleteUser(suite.Ctx, 1)
	require.NoError(suite.T(), err)

	err = suite.ormStore.DeleteUser(suite.Ctx, 2)
	require.NoError(suite.T(), err)
}

func (suite *PrincipalUserSuite) TestListUser() {
	appendItems := []struct {
		Id    int64
		Admin bool
	}{
		{101, true}, {102, false}, {103, false},
		{104, true}, {105, true},
	}

	for _, item := range appendItems {
		testsuite.AddUser(suite.Ctx, suite.T(), suite.ormStore, item.Id, item.Admin)
	}

	tests := []struct {
		filter types.UserFilter
		length int
	}{
		{filter: types.UserFilter{}, length: 9},
		{filter: types.UserFilter{Sort: enum.UserAttrName}, length: 9},
		{filter: types.UserFilter{Sort: enum.UserAttrNone}, length: 9},
		{filter: types.UserFilter{Size: 5, Sort: enum.UserAttrCreated}, length: 5},
		{filter: types.UserFilter{Page: 2, Size: 5, Sort: enum.UserAttrCreated}, length: 4},
		{filter: types.UserFilter{Page: 3, Size: 5, Sort: enum.UserAttrCreated}, length: 0},
		{filter: types.UserFilter{Page: 1, Size: 5, Sort: enum.UserAttrUpdated}, length: 5},
		{filter: types.UserFilter{Page: 1, Size: 5, Sort: enum.UserAttrUpdated, Order: enum.OrderAsc}, length: 5},
		{filter: types.UserFilter{Page: 1, Size: 5, Sort: enum.UserAttrUpdated, Order: enum.OrderDesc}, length: 5},
		{filter: types.UserFilter{Page: 2, Size: 5, Sort: enum.UserAttrEmail, Order: enum.OrderDesc}, length: 4},
		{filter: types.UserFilter{Page: 2, Size: 5, Sort: enum.UserAttrEmail, Order: enum.OrderAsc}, length: 4},
		{filter: types.UserFilter{Page: 1, Size: 4, Sort: enum.UserAttrUID, Order: enum.OrderDesc}, length: 4},
		{filter: types.UserFilter{Page: 2, Size: 4, Sort: enum.UserAttrUID, Order: enum.OrderDesc}, length: 4},
		{filter: types.UserFilter{Page: 3, Size: 4, Sort: enum.UserAttrUID, Order: enum.OrderDesc}, length: 1},
		{filter: types.UserFilter{Page: 1, Size: 4, Sort: enum.UserAttrAdmin, Order: enum.OrderAsc}, length: 4},
		{filter: types.UserFilter{Page: 3, Size: 4, Sort: enum.UserAttrUID, Order: enum.OrderAsc}, length: 1},
	}

	for id, test := range tests {
		users, err := suite.ormStore.ListUsers(suite.Ctx, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.length, len(users))

		usersB, err := suite.sqlxStore.ListUsers(suite.Ctx, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), users, usersB)
	}
}

func (suite *PrincipalUserSuite) TestCountUser() {
	appendItems := []struct {
		Id    int64
		Admin bool
	}{
		{101, true}, {102, false}, {103, false},
		{104, true}, {105, true},
	}

	for _, item := range appendItems {
		testsuite.AddUser(suite.Ctx, suite.T(), suite.ormStore, item.Id, item.Admin)
	}

	tests := []struct {
		filter types.UserFilter
		length int64
	}{
		// only Admin will be affected in CountUser
		{filter: types.UserFilter{}, length: 9},
		{filter: types.UserFilter{Sort: enum.UserAttrName}, length: 9},
		{filter: types.UserFilter{Page: 1, Size: 5, Sort: enum.UserAttrCreated}, length: 9},
		{filter: types.UserFilter{Page: 3, Size: 5, Sort: enum.UserAttrCreated}, length: 9},
		{filter: types.UserFilter{Admin: true}, length: 5},
	}

	for id, test := range tests {
		count, err := suite.ormStore.CountUsers(suite.Ctx, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.length, count)

		countB, err := suite.sqlxStore.CountUsers(suite.Ctx, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.length, countB)
	}
}
