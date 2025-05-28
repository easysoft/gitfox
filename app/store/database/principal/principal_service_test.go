// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package principal_test

import (
	"context"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PrincipalServiceSuite struct {
	testsuite.BaseSuite

	ormStore  *principal.PrincipalOrmStore
	sqlxStore *database.PrincipalStore
}

func TestPrincipalServiceSuite(t *testing.T) {
	ctx := context.Background()

	st := &PrincipalServiceSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "principals_service",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = principal.NewPrincipalOrmStore(st.Gdb, store.ProvidePrincipalUIDTransformation())
		st.sqlxStore = database.NewPrincipalStore(st.Sdb, store.ProvidePrincipalUIDTransformation())
	}

	suite.Run(t, st)
}

func (suite *PrincipalServiceSuite) SetupTest() {
	suite.addData()
}

func (suite *PrincipalServiceSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table("principals").Where("1 = 1").Delete(nil)
}

func (suite *PrincipalServiceSuite) addData() {
	now := time.Now().UnixMilli()
	items := []types.Service{
		{ID: 1, UID: "svc_1", Email: "svc_1" + emailSuffix, DisplayName: "Service 1", Updated: now},
		{ID: 2, UID: "svc_2", Email: "svc_2" + emailSuffix, DisplayName: "Service 2", Updated: now},
		{ID: 3, UID: "svc_3", Email: "svc_3" + emailSuffix, DisplayName: "Service 3", Updated: now},
		{ID: 4, UID: "svc_4", Email: "svc_4" + emailSuffix, DisplayName: "Service 4", Updated: now},
	}

	for id, item := range items {
		err := suite.ormStore.CreateService(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PrincipalServiceSuite) TestCreateService() {
}

func (suite *PrincipalServiceSuite) TestFindService() {
	tests := []struct {
		ID        int64
		UID       string
		expectErr bool
	}{
		{ID: 1, UID: "svc_1"}, {ID: 2, UID: "svc_2"},
		{ID: 3, UID: "svc_3"}, {ID: 4, UID: "svc_4"},

		// invalid test
		{ID: 11, expectErr: true},
	}

	for id, test := range tests {
		svc, err := suite.ormStore.FindService(suite.Ctx, test.ID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.UID, svc.UID, testsuite.InvalidLoopMsgF, id)
		}

		// check same
		svcB, err := suite.sqlxStore.FindService(suite.Ctx, test.ID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.EqualExportedValues(suite.T(), *svc, *svcB, testsuite.InvalidLoopMsgF, id)
		}
	}
}

func (suite *PrincipalServiceSuite) TestFindServiceByUID() {
	tests := []struct {
		ID        int64
		UID       string
		expectErr bool
	}{
		{ID: 1, UID: "svc_1"}, {ID: 2, UID: "svc_2"},
		{ID: 3, UID: "svc_3"}, {ID: 4, UID: "svc_4"},

		// invalid test
		{ID: 11, UID: "svc_11", expectErr: true},
	}

	for id, test := range tests {
		svc, err := suite.ormStore.FindServiceByUID(suite.Ctx, test.UID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.ID, svc.ID, testsuite.InvalidLoopMsgF, id)
		}

		// check same
		svcB, err := suite.sqlxStore.FindServiceByUID(suite.Ctx, test.UID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.EqualExportedValues(suite.T(), *svc, *svcB, testsuite.InvalidLoopMsgF, id)
		}
	}
}

func (suite *PrincipalServiceSuite) TestUpdateService() {
	upEmailSuffix := "@local.test"
	now := time.Now().UnixMilli()

	updates := []types.Service{
		{ID: 1, UID: "svc_1", Email: "svc_1" + upEmailSuffix, DisplayName: "Service 1", Updated: now}, // change email
		{ID: 2, UID: "svc_2.1", Email: "svc_2" + emailSuffix, DisplayName: "Service 2", Updated: now}, // uid won't be changed

	}

	updates2 := []types.Service{
		{ID: 3, UID: "svc_3", Email: "svc_3" + upEmailSuffix, DisplayName: "Service 3", Updated: now}, // change email
		{ID: 4, UID: "svc_4.1", Email: "svc_4" + emailSuffix, DisplayName: "Service 4", Updated: now}, // uid won't be changed
	}

	updatesInvalids := []types.Service{
		{ID: 1, UID: "svc_1", Email: "svc_2" + emailSuffix, DisplayName: "Service 1", Updated: now},   // duplicate email
		{ID: 4, UID: "svc_4", Email: "svc_3" + upEmailSuffix, DisplayName: "Service 4", Updated: now}, // duplicate email
	}

	for id, up := range updates {
		err := suite.ormStore.UpdateService(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, up := range updates2 {
		err := suite.sqlxStore.UpdateService(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, up := range updatesInvalids {
		err := suite.ormStore.UpdateService(suite.Ctx, &up)
		require.Error(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		err = suite.sqlxStore.UpdateService(suite.Ctx, &up)
		require.Error(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	tests := []struct {
		ID        int64
		fieldName string
		expect    interface{}
	}{
		{ID: 1, expect: "svc_1" + upEmailSuffix, fieldName: "Email"},
		{ID: 2, expect: now, fieldName: "Updated"},
		{ID: 2, expect: "svc_2", fieldName: "UID"},
		{ID: 3, expect: "Service 3", fieldName: "DisplayName"},
		{ID: 3, expect: "svc_3" + upEmailSuffix, fieldName: "Email"},
		{ID: 4, expect: "svc_4", fieldName: "UID"},
		{ID: 4, expect: "svc_4" + emailSuffix, fieldName: "Email"},
	}

	for id, test := range tests {
		svc, err := suite.ormStore.FindService(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		testsuite.EqualFieldValue(suite.T(), test.expect, svc, test.fieldName, testsuite.InvalidLoopMsgF, id)

		svcB, err := suite.sqlxStore.FindService(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		testsuite.EqualFieldValue(suite.T(), test.expect, svcB, test.fieldName, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PrincipalServiceSuite) TestDeleteService() {
	err := suite.ormStore.DeleteService(suite.Ctx, 1)
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.DeleteService(suite.Ctx, 2)
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.FindService(suite.Ctx, 2)
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.FindService(suite.Ctx, 1)
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.DeleteService(suite.Ctx, 1)
	require.NoError(suite.T(), err)

	err = suite.ormStore.DeleteService(suite.Ctx, 2)
	require.NoError(suite.T(), err)
}

func (suite *PrincipalServiceSuite) TestListServiceAndCount() {
	var expectCount = 4

	svcList, err := suite.ormStore.ListServices(suite.Ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), expectCount, len(svcList))

	svcCount, err := suite.ormStore.CountServices(suite.Ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int64(expectCount), svcCount)

	svcListB, err := suite.sqlxStore.ListServices(suite.Ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), expectCount, len(svcListB))

	require.ElementsMatch(suite.T(), svcList, svcListB)
}
