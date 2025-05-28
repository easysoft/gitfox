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
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PrincipalInfoSuite struct {
	testsuite.BaseSuite

	ormStore  *principal.InfoView
	sqlxStore *database.PrincipalInfoView
}

func TestPrincipalInfoSuite(t *testing.T) {
	ctx := context.Background()

	st := &PrincipalInfoSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "principals_info",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = principal.NewPrincipalOrmInfoView(st.Gdb)
		st.sqlxStore = database.NewPrincipalInfoView(st.Sdb)

		writer := principal.NewPrincipalOrmStore(st.Gdb, store.ProvidePrincipalUIDTransformation())
		for id, item := range svcItems {
			err := writer.CreateService(st.Ctx, &item)
			require.NoError(st.T(), err, testsuite.InvalidLoopMsgF, id)
		}

		for id, item := range svcAccountItems {
			err := writer.CreateServiceAccount(st.Ctx, &item)
			require.NoError(st.T(), err, testsuite.InvalidLoopMsgF, id)
		}

		for id, item := range userItems {
			err := writer.CreateUser(st.Ctx, &item)
			require.NoError(st.T(), err, testsuite.InvalidLoopMsgF, id)
		}
	}

	suite.Run(t, st)
}

func (suite *PrincipalInfoSuite) TestFind() {
	tests := []testListCase{
		{1, enum.PrincipalTypeService, "svc_1", "svc_1" + emailSuffix},
		{3, enum.PrincipalTypeService, "svc_3", "svc_3" + emailSuffix},
		{11, enum.PrincipalTypeServiceAccount, "sa_1", "sa_1" + emailSuffix},
		{13, enum.PrincipalTypeServiceAccount, "sa_3", "sa_3" + emailSuffix},
		{21, enum.PrincipalTypeUser, "admin", "admin" + emailSuffix},
		{22, enum.PrincipalTypeUser, "user1", "user1" + emailSuffix},
	}

	for _, test := range tests {
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

func (suite *PrincipalInfoSuite) TestFindMany() {
	tests := []struct {
		idList     []int64
		wantLength int
	}{
		{[]int64{11}, 1},
		{[]int64{11, 12, 13}, 3},
		{[]int64{11, 12, 18}, 2},
		{[]int64{11, 2, 21}, 3},
		{[]int64{1, 13, 11, 24, 2, 25}, 5},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.FindMany(suite.Ctx, test.idList)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.FindMany(suite.Ctx, test.idList)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.wantLength, len(objsB), testsuite.InvalidLoopMsgF, id)

		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}
