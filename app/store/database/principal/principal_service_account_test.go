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
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PrincipalSASuite struct {
	testsuite.BaseSuite

	ormStore  *principal.PrincipalOrmStore
	sqlxStore *database.PrincipalStore
}

func TestPrincipalSASuite(t *testing.T) {
	ctx := context.Background()

	st := &PrincipalSASuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "principals_sa",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = principal.NewPrincipalOrmStore(st.Gdb, store.ProvidePrincipalUIDTransformation())
		st.sqlxStore = database.NewPrincipalStore(st.Sdb, store.ProvidePrincipalUIDTransformation())
	}

	suite.Run(t, st)
}

func (suite *PrincipalSASuite) SetupTest() {
	suite.addData()
}

func (suite *PrincipalSASuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table("principals").Where("1 = 1").Delete(nil)
}

func (suite *PrincipalSASuite) addData() {
	now := time.Now().UnixMilli()
	items := []types.ServiceAccount{
		{ID: 1, UID: "sa_1", Email: "sa_1" + emailSuffix, DisplayName: "ServiceAccount 1", Updated: now,
			ParentType: enum.ParentResourceTypeSpace, ParentID: 1},
		{ID: 2, UID: "sa_2", Email: "sa_2" + emailSuffix, DisplayName: "ServiceAccount 2", Updated: now,
			ParentType: enum.ParentResourceTypeSpace, ParentID: 1},
		{ID: 3, UID: "sa_3", Email: "sa_3" + emailSuffix, DisplayName: "ServiceAccount 3", Updated: now,
			ParentType: enum.ParentResourceTypeSpace, ParentID: 2},
		{ID: 4, UID: "sa_4", Email: "sa_4" + emailSuffix, DisplayName: "ServiceAccount 4", Updated: now,
			ParentType: enum.ParentResourceTypeRepo, ParentID: 1},
		{ID: 5, UID: "sa_5", Email: "sa_5" + emailSuffix, DisplayName: "ServiceAccount 5", Updated: now,
			ParentType: enum.ParentResourceTypeRepo, ParentID: 2},
		{ID: 6, UID: "sa_6", Email: "sa_6" + emailSuffix, DisplayName: "ServiceAccount 6", Updated: now,
			ParentType: enum.ParentResourceTypeRepo, ParentID: 3},
	}

	for id, item := range items {
		err := suite.ormStore.CreateServiceAccount(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PrincipalSASuite) TestCreateServiceAccount() {
}

func (suite *PrincipalSASuite) TestFindServiceAccount() {
	tests := []struct {
		ID        int64
		UID       string
		expectErr bool
	}{
		{ID: 1, UID: "sa_1"}, {ID: 2, UID: "sa_2"},
		{ID: 3, UID: "sa_3"}, {ID: 4, UID: "sa_4"},

		// invalid test
		{ID: 11, expectErr: true},
	}

	for id, test := range tests {
		sa, err := suite.ormStore.FindServiceAccount(suite.Ctx, test.ID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.UID, sa.UID, testsuite.InvalidLoopMsgF, id)
		}

		// check same
		saB, err := suite.sqlxStore.FindServiceAccount(suite.Ctx, test.ID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.EqualExportedValues(suite.T(), *sa, *saB, testsuite.InvalidLoopMsgF, id)
		}
	}
}

func (suite *PrincipalSASuite) TestFindServiceAccountByUID() {
	tests := []struct {
		ID        int64
		UID       string
		expectErr bool
	}{
		{ID: 1, UID: "sa_1"}, {ID: 2, UID: "sa_2"},
		{ID: 3, UID: "sa_3"}, {ID: 4, UID: "sa_4"},

		// invalid test
		{ID: 11, UID: "sa_11", expectErr: true},
	}

	for id, test := range tests {
		sa, err := suite.ormStore.FindServiceAccountByUID(suite.Ctx, test.UID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.ID, sa.ID, testsuite.InvalidLoopMsgF, id)
		}

		// check same
		saB, err := suite.sqlxStore.FindServiceAccountByUID(suite.Ctx, test.UID)
		if test.expectErr {
			require.Error(suite.T(), err)
		} else {
			require.NoError(suite.T(), err)
			require.EqualExportedValues(suite.T(), *sa, *saB, testsuite.InvalidLoopMsgF, id)
		}
	}
}

func (suite *PrincipalSASuite) TestUpdateServiceAccount() {
	upEmailSuffix := "@local.test"
	now := time.Now().UnixMilli()

	updates := []types.ServiceAccount{
		{ID: 1, UID: "sa_1", Email: "sa_1" + upEmailSuffix, DisplayName: "ServiceAccount 1", Updated: now,
			ParentType: enum.ParentResourceTypeSpace, ParentID: 1}, // change email
		{ID: 2, UID: "sa_2", Email: "sa_2" + emailSuffix, DisplayName: "ServiceAccount 2", Updated: now,
			ParentType: enum.ParentResourceTypeSpace, ParentID: 1}, // nothing
		{ID: 3, UID: "sa_3.1", Email: "sa_3" + emailSuffix, DisplayName: "ServiceAccount 3", Updated: now,
			ParentType: enum.ParentResourceTypeSpace, ParentID: 3}, // uid, parentId won't be changed
	}

	updates2 := []types.ServiceAccount{
		{ID: 4, UID: "sa_4", Email: "sa_4" + emailSuffix, DisplayName: "ServiceAccount 4", Updated: now,
			ParentType: enum.ParentResourceTypeRepo, ParentID: 1}, // nothing
		{ID: 5, UID: "sa_5", Email: "sa_5" + upEmailSuffix, DisplayName: "ServiceAccount 5", Updated: now,
			ParentType: enum.ParentResourceTypeRepo, ParentID: 2}, // change email
		{ID: 6, UID: "sa_6.2", Email: "sa_6.2" + emailSuffix, DisplayName: "ServiceAccount 6.2", Updated: now,
			ParentType: enum.ParentResourceTypeRepo, ParentID: 3}, // uid won't be changed; changed email displayName
	}

	updatesInvalids := []types.ServiceAccount{
		{ID: 1, UID: "sa_1", Email: "sa_2" + emailSuffix, DisplayName: "ServiceAccount 1", Updated: now,
			ParentType: enum.ParentResourceTypeSpace, ParentID: 1}, // duplicate email
		{ID: 4, UID: "sa_4", Email: "sa_6.2" + emailSuffix, DisplayName: "ServiceAccount 4", Updated: now,
			ParentType: enum.ParentResourceTypeRepo, ParentID: 1}, // duplicate email
	}

	for id, up := range updates {
		err := suite.ormStore.UpdateServiceAccount(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, up := range updates2 {
		err := suite.sqlxStore.UpdateServiceAccount(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, up := range updatesInvalids {
		err := suite.ormStore.UpdateServiceAccount(suite.Ctx, &up)
		require.Error(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		err = suite.sqlxStore.UpdateServiceAccount(suite.Ctx, &up)
		require.Error(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	tests := []struct {
		ID        int64
		fieldName string
		expect    interface{}
	}{
		{ID: 1, expect: "sa_1" + upEmailSuffix, fieldName: "Email"},
		{ID: 2, expect: now, fieldName: "Updated"},
		{ID: 3, expect: "sa_3", fieldName: "UID"},
		{ID: 3, expect: int64(2), fieldName: "ParentID"},
		{ID: 4, expect: "ServiceAccount 4", fieldName: "DisplayName"},
		{ID: 5, expect: "sa_5" + upEmailSuffix, fieldName: "Email"},
		{ID: 6, expect: "sa_6", fieldName: "UID"},
		{ID: 6, expect: "sa_6.2" + emailSuffix, fieldName: "Email"},
		{ID: 6, expect: "ServiceAccount 6.2", fieldName: "DisplayName"},
	}

	for id, test := range tests {
		sa, err := suite.ormStore.FindServiceAccount(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		testsuite.EqualFieldValue(suite.T(), test.expect, sa, test.fieldName, testsuite.InvalidLoopMsgF, id)

		saB, err := suite.sqlxStore.FindServiceAccount(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		testsuite.EqualFieldValue(suite.T(), test.expect, saB, test.fieldName, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PrincipalSASuite) TestDeleteServiceAccount() {
	err := suite.ormStore.DeleteServiceAccount(suite.Ctx, 1)
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.DeleteServiceAccount(suite.Ctx, 2)
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.FindServiceAccount(suite.Ctx, 2)
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.FindServiceAccount(suite.Ctx, 1)
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.DeleteServiceAccount(suite.Ctx, 1)
	require.NoError(suite.T(), err)

	err = suite.ormStore.DeleteServiceAccount(suite.Ctx, 2)
	require.NoError(suite.T(), err)
}

func (suite *PrincipalSASuite) TestListServiceAccountAndCount() {
	tests := []struct {
		parentType   enum.ParentResourceType
		parentID     int64
		expectLength int
	}{
		{parentType: enum.ParentResourceTypeSpace, parentID: 1, expectLength: 2},
		{parentType: enum.ParentResourceTypeSpace, parentID: 2, expectLength: 1},
		{parentType: enum.ParentResourceTypeSpace, parentID: 3, expectLength: 0},
		{parentType: enum.ParentResourceTypeRepo, parentID: 1, expectLength: 1},
		{parentType: enum.ParentResourceTypeRepo, parentID: 2, expectLength: 1},
		{parentType: enum.ParentResourceTypeRepo, parentID: 3, expectLength: 1},
		{parentType: enum.ParentResourceTypeRepo, parentID: 4, expectLength: 0},
	}

	for id, test := range tests {
		saList, err := suite.ormStore.ListServiceAccounts(suite.Ctx, test.parentType, test.parentID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.expectLength, len(saList), testsuite.InvalidLoopMsgF, id)

		saCount, err := suite.ormStore.CountServiceAccounts(suite.Ctx, test.parentType, test.parentID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.expectLength, int(saCount), testsuite.InvalidLoopMsgF, id)

		saListB, err := suite.sqlxStore.ListServiceAccounts(suite.Ctx, test.parentType, test.parentID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.expectLength, len(saListB), testsuite.InvalidLoopMsgF, id)
		if test.expectLength > 0 {
			require.ElementsMatch(suite.T(), saList, saListB, testsuite.InvalidLoopMsgF, id)
		}
	}
}
