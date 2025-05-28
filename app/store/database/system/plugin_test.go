// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package system_test

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/system"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTablePlugin = "plugins"
)

type SecretSuite struct {
	testsuite.BaseSuite

	ormStore  *system.PluginStore
	sqlxStore store.PluginStore
}

func TestSecretSuite(t *testing.T) {
	ctx := context.Background()

	st := &SecretSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "plugins",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = system.NewPluginOrmStore(st.Gdb)
		st.sqlxStore = database.NewPluginStore(st.Sdb)

		// add init data
		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 2, 1)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 3, 1)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 4, 2)
	}

	suite.Run(t, st)
}

func (suite *SecretSuite) SetupTest() {
	suite.addData()
}

func (suite *SecretSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTablePlugin).Where("1 = 1").Delete(nil)
}

var testAddPluginItems = []struct {
	Type       string
	identifier string
	version    string
}{
	{Type: type1, identifier: "plugin_1", version: "1.0"},
	{Type: type1, identifier: "plugin_2", version: "1.1"},
	{Type: type1, identifier: "plugin_3", version: "1.8.1"},
	{Type: type1, identifier: "plugin_4", version: "1.8.2"},
	{Type: type3, identifier: "plugin_5", version: "1.10.1"},
	{Type: type3, identifier: "plugin_6", version: "1.10.2"},
}

func (suite *SecretSuite) addData() {
	for id, item := range testAddPluginItems {
		obj := types.Plugin{
			Identifier: item.identifier,
			Type:       item.Type,
			Version:    item.version,
		}

		err := suite.ormStore.Create(suite.Ctx, &obj)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SecretSuite) TestFind() {
	for id, test := range testAddPluginItems {
		obj, err := suite.ormStore.Find(suite.Ctx, test.identifier, test.version)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.identifier, obj.Identifier)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.identifier, test.version)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SecretSuite) TestListAll() {
	objs, err := suite.ormStore.ListAll(suite.Ctx)
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 6, len(objs))

	objsB, err := suite.sqlxStore.ListAll(suite.Ctx)
	require.NoError(suite.T(), err)
	require.ElementsMatch(suite.T(), objs, objsB)
}

func (suite *SecretSuite) TestListAncCount() {
	tests := []struct {
		filter     types.ListQueryFilter
		wantLength int64
	}{
		{types.ListQueryFilter{}, 6},
		{types.ListQueryFilter{Pagination: types.Pagination{Page: 1, Size: 5}}, 5},
		{types.ListQueryFilter{Pagination: types.Pagination{Page: 2, Size: 5}}, 1},
		{types.ListQueryFilter{Pagination: types.Pagination{Page: 3, Size: 5}}, 0},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.List(suite.Ctx, test.filter)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}

	count, err := suite.ormStore.Count(suite.Ctx, types.ListQueryFilter{})

	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 6, count)
}
