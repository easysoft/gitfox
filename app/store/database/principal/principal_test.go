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

const testTableName = "principals"

type PrincipalSuite struct {
	testsuite.BaseSuite

	ormStore  *principal.PrincipalOrmStore
	sqlxStore *database.PrincipalStore
}

func TestPrincipalSuite(t *testing.T) {
	ctx := context.Background()

	st := &PrincipalSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "principals",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = principal.NewPrincipalOrmStore(st.Gdb, store.ProvidePrincipalUIDTransformation())
		st.sqlxStore = database.NewPrincipalStore(st.Sdb, store.ProvidePrincipalUIDTransformation())
	}

	suite.Run(t, st)
}

func (suite *PrincipalSuite) SetupTest() {
	suite.addData()
}

func (suite *PrincipalSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableName).Where("1 = 1").Delete(nil)
}

func (suite *PrincipalSuite) addData() {
	for id, item := range svcItems {
		err := suite.ormStore.CreateService(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, item := range svcAccountItems {
		err := suite.ormStore.CreateServiceAccount(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, item := range userItems {
		err := suite.ormStore.CreateUser(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	dupUser := userItems[0]
	dupUser.ID = 1
	dupUser.Email = "admin_1" + emailSuffix
	dupUser.UID = "admin_1"
	err := suite.ormStore.CreateUser(suite.Ctx, &dupUser)
	require.Error(suite.T(), err)
}

type testListCase struct {
	ID    int64
	Type  enum.PrincipalType
	UID   string
	Email string
}

func (suite *PrincipalSuite) testFindCases() []testListCase {
	return []testListCase{
		{1, enum.PrincipalTypeService, "svc_1", "svc_1" + emailSuffix},
		{3, enum.PrincipalTypeService, "svc_3", "svc_3" + emailSuffix},
		{11, enum.PrincipalTypeServiceAccount, "sa_1", "sa_1" + emailSuffix},
		{13, enum.PrincipalTypeServiceAccount, "sa_3", "sa_3" + emailSuffix},
		{21, enum.PrincipalTypeUser, "admin", "admin" + emailSuffix},
		{22, enum.PrincipalTypeUser, "user1", "user1" + emailSuffix},
	}
}

func (suite *PrincipalSuite) TestFind() {
	for _, test := range suite.testFindCases() {
		obj, err := suite.ormStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.Type, obj.Type)
		require.Equal(suite.T(), test.UID, obj.UID)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.Type, objB.Type)
		require.Equal(suite.T(), test.UID, objB.UID)
	}

	_, err := suite.ormStore.Find(suite.Ctx, 100)
	require.Error(suite.T(), err)
}

func (suite *PrincipalSuite) TestFindByUID() {
	for _, test := range suite.testFindCases() {
		obj, err := suite.ormStore.FindByUID(suite.Ctx, test.UID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.Type, obj.Type)
		require.Equal(suite.T(), test.ID, obj.ID)

		objB, err := suite.sqlxStore.FindByUID(suite.Ctx, test.UID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.Type, objB.Type)
		require.Equal(suite.T(), test.ID, objB.ID)
	}

	_, err := suite.ormStore.FindByUID(suite.Ctx, "u_100")
	require.Error(suite.T(), err)
}

func (suite *PrincipalSuite) TestFindManyByUID() {
	tests := []struct {
		uidList    []string
		wantLength int
	}{
		{[]string{"sa_1"}, 1},
		{[]string{"sa_1", "sa_2", "sa_3"}, 3},
		{[]string{"sa_1", "sa_2", "sa_8"}, 2},
		{[]string{"sa_1", "svc_2", "admin"}, 3},
		{[]string{"svc_1", "sa_3", "sa_1", "admin2", "svc_2", "admin3"}, 5},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.FindManyByUID(suite.Ctx, test.uidList)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.FindManyByUID(suite.Ctx, test.uidList)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.wantLength, len(objsB), testsuite.InvalidLoopMsgF, id)

		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PrincipalSuite) TestFindByEmail() {
	for _, test := range suite.testFindCases() {
		obj, err := suite.ormStore.FindByEmail(suite.Ctx, test.Email)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.Type, obj.Type)
		require.Equal(suite.T(), test.ID, obj.ID)

		objB, err := suite.sqlxStore.FindByEmail(suite.Ctx, test.Email)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.Type, objB.Type)
		require.Equal(suite.T(), test.ID, objB.ID)
	}

	_, err := suite.ormStore.FindByEmail(suite.Ctx, "u_100")
	require.Error(suite.T(), err)
}

func (suite *PrincipalSuite) TestList() {
	tests := []struct {
		filter     types.PrincipalFilter
		wantLength int
	}{
		{filter: types.PrincipalFilter{}, wantLength: 14},
		{filter: types.PrincipalFilter{
			Types: []enum.PrincipalType{enum.PrincipalTypeUser}}, wantLength: 4},
		{filter: types.PrincipalFilter{
			Types: []enum.PrincipalType{enum.PrincipalTypeService}}, wantLength: 4},
		{filter: types.PrincipalFilter{
			Types: []enum.PrincipalType{enum.PrincipalTypeServiceAccount}}, wantLength: 6},
		{filter: types.PrincipalFilter{
			Types: []enum.PrincipalType{
				enum.PrincipalTypeServiceAccount, enum.PrincipalTypeUser,
			}}, wantLength: 10},
		{filter: types.PrincipalFilter{
			Types: []enum.PrincipalType{
				enum.PrincipalTypeServiceAccount, enum.PrincipalTypeService,
			}}, wantLength: 10},
		{filter: types.PrincipalFilter{
			Types: []enum.PrincipalType{
				enum.PrincipalTypeService, enum.PrincipalTypeUser,
			}}, wantLength: 8},
		{filter: types.PrincipalFilter{
			Page: 1, Size: 8,
			Types: []enum.PrincipalType{
				enum.PrincipalTypeServiceAccount, enum.PrincipalTypeService,
			}}, wantLength: 8},
		{filter: types.PrincipalFilter{
			Page: 2, Size: 8,
			Types: []enum.PrincipalType{
				enum.PrincipalTypeServiceAccount, enum.PrincipalTypeService,
			}}, wantLength: 2},
		{filter: types.PrincipalFilter{
			Page: 3, Size: 8,
			Types: []enum.PrincipalType{
				enum.PrincipalTypeServiceAccount, enum.PrincipalTypeService,
			}}, wantLength: 0},
		{filter: types.PrincipalFilter{Query: "admin"}, wantLength: 2},
		{filter: types.PrincipalFilter{Query: "serviceaccount"}, wantLength: 6},
		{filter: types.PrincipalFilter{Query: "user", Types: []enum.PrincipalType{enum.PrincipalTypeService}}, wantLength: 0},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, &test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.List(suite.Ctx, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB)
	}
}
