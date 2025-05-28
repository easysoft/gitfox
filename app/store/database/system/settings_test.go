// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package system_test

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/system"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableSetting = "settings"
)

type SettingsSuite struct {
	testsuite.BaseSuite

	ormStore  *system.SettingsStore
	sqlxStore *database.SettingsStore
}

func TestSettingsSuite(t *testing.T) {
	ctx := context.Background()

	st := &SettingsSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "settings",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = system.NewSettingsOrmStore(st.Gdb)
		st.sqlxStore = database.NewSettingsStore(st.Sdb)

		// add init data
		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)

		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 2, 1)

		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 20)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 2, 2, 20)
	}

	suite.Run(t, st)
}

func (suite *SettingsSuite) SetupTest() {
	suite.addData()
}

func (suite *SettingsSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableSetting).Where("1 = 1").Delete(nil)
}

type testSettings struct {
	scope   enum.SettingsScope
	scopeID int64
	key     string
	value   string
}

var testAddSettings = []testSettings{
	{scope: enum.SettingsScopeSpace, scopeID: 1, key: "space1_1", value: `{"k1": "v1"}`},
	{scope: enum.SettingsScopeSpace, scopeID: 1, key: "space1_2", value: `{"k1": "v2"}`},
	{scope: enum.SettingsScopeSpace, scopeID: 2, key: "space2_1", value: `{"k1": "v3"}`},
	{scope: enum.SettingsScopeSpace, scopeID: 2, key: "space2_2", value: `{"k1": "v4"}`},
	{scope: enum.SettingsScopeRepo, scopeID: 1, key: "repo1_1", value: `{"k1": "v1"}`},
	{scope: enum.SettingsScopeRepo, scopeID: 1, key: "repo1_2", value: `{"k1": "v2"}`},
	{scope: enum.SettingsScopeRepo, scopeID: 2, key: "repo1_1", value: `{"k1": "v3"}`},
	{scope: enum.SettingsScopeRepo, scopeID: 2, key: "repo1_2", value: `{"k1": "v4"}`},
}

func (suite *SettingsSuite) addData() {
	for id, item := range testAddSettings {
		err := suite.ormStore.Upsert(suite.Ctx, item.scope, item.scopeID, item.key, []byte(item.value))
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SettingsSuite) TestFind() {
	for id, test := range testAddSettings {
		val, err := suite.ormStore.Find(suite.Ctx, test.scope, test.scopeID, test.key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), []byte(test.value), val)

		valB, err := suite.sqlxStore.Find(suite.Ctx, test.scope, test.scopeID, test.key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), val, valB)
	}
}

func (suite *SettingsSuite) TestFindMany() {
	tests := []struct {
		scope   enum.SettingsScope
		scopeID int64
		keys    []string
	}{
		{scope: enum.SettingsScopeSpace, scopeID: 1, keys: []string{"space1_1", "space1_2"}},
		{scope: enum.SettingsScopeSpace, scopeID: 2, keys: []string{"space2_1", "space2_2"}},
		{scope: enum.SettingsScopeRepo, scopeID: 1, keys: []string{"repo1_1", "repo1_2"}},
		{scope: enum.SettingsScopeRepo, scopeID: 2, keys: []string{"repo2_1", "repo2_2"}},
	}

	for id, test := range tests {
		vals, err := suite.ormStore.FindMany(suite.Ctx, test.scope, test.scopeID, test.keys...)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		valsB, err := suite.sqlxStore.FindMany(suite.Ctx, test.scope, test.scopeID, test.keys...)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), vals, valsB)
	}
}

func (suite *SettingsSuite) TestUpsert() {
	tests := []testSettings{
		{scope: enum.SettingsScopeSpace, scopeID: 1, key: "space1_1", value: `{"k1": "v1_u1"}`},
		{scope: enum.SettingsScopeSpace, scopeID: 1, key: "space1_2", value: `{"k1": "v2_u1"}`},
		{scope: enum.SettingsScopeRepo, scopeID: 1, key: "repo1_1", value: `{"k1": "v1_u1"}`},
		{scope: enum.SettingsScopeRepo, scopeID: 1, key: "repo1_2", value: `{"k1": "v2_u1"}`},
	}

	for id, test := range tests {
		// updated value
		err := suite.ormStore.Upsert(suite.Ctx, test.scope, test.scopeID, test.key, []byte(test.value))
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		val, err := suite.ormStore.Find(suite.Ctx, test.scope, test.scopeID, test.key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), []byte(test.value), val)

		// nothing changed
		err = suite.sqlxStore.Upsert(suite.Ctx, test.scope, test.scopeID, test.key, []byte(test.value))
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		valB, err := suite.sqlxStore.Find(suite.Ctx, test.scope, test.scopeID, test.key)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), val, valB)
	}
}
