// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package space_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/space"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableSecret = "secrets"
)

type SecretSuite struct {
	testsuite.BaseSuite

	ormStore  *space.SecretStore
	sqlxStore store.SecretStore
}

func TestSecretSuite(t *testing.T) {
	ctx := context.Background()

	st := &SecretSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "secrets",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = space.NewSecretOrmStore(st.Gdb)
		st.sqlxStore = database.NewSecretStore(st.Sdb)

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
	suite.Gdb.WithContext(suite.Ctx).Table(testTableSecret).Where("1 = 1").Delete(nil)
}

var testAddSecretItems = []struct {
	id         int64
	spaceId    int64
	identifier string
}{
	{id: 1, spaceId: 1, identifier: fmt.Sprintf("secret_1")},
	{id: 2, spaceId: 2, identifier: fmt.Sprintf("secret_2")},
	{id: 3, spaceId: 2, identifier: fmt.Sprintf("secret_3")},
	{id: 4, spaceId: 2, identifier: fmt.Sprintf("secret_4")},
	{id: 5, spaceId: 3, identifier: fmt.Sprintf("secret_5")},
	{id: 6, spaceId: 3, identifier: fmt.Sprintf("secret_6")},
	{id: 7, spaceId: 4, identifier: fmt.Sprintf("secret_7")},
}

func (suite *SecretSuite) addData() {
	now := time.Now().UnixMilli()
	for id, item := range testAddSecretItems {
		obj := types.Secret{
			ID:         item.id,
			SpaceID:    item.spaceId,
			Identifier: item.identifier,
			Data:       "",
			CreatedBy:  1,
			Created:    now,
			Updated:    now,
		}

		err := suite.ormStore.Create(suite.Ctx, &obj)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SecretSuite) TestFind() {
	for id, test := range testAddSecretItems {
		obj, err := suite.ormStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.identifier, obj.Identifier)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SecretSuite) TestFindByIdentifier() {
	for id, test := range testAddSecretItems {
		obj, err := suite.ormStore.FindByIdentifier(suite.Ctx, test.spaceId, test.identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.id, obj.ID)

		objB, err := suite.sqlxStore.FindByIdentifier(suite.Ctx, test.spaceId, test.identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SecretSuite) TestDelete() {
	err := suite.ormStore.Delete(suite.Ctx, 1)
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.Delete(suite.Ctx, 2)
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.Find(suite.Ctx, 2)
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.Find(suite.Ctx, 1)
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.Delete(suite.Ctx, 1)
	require.NoError(suite.T(), err)

	err = suite.ormStore.Delete(suite.Ctx, 2)
	require.NoError(suite.T(), err)
}

func (suite *SecretSuite) TestDeleteByIdentifier() {
	err := suite.ormStore.DeleteByIdentifier(suite.Ctx, 1, "secret_1")
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, 3, "secret_6")
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.FindByIdentifier(suite.Ctx, 3, "secret_6")
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.FindByIdentifier(suite.Ctx, 1, "secret_1")
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, 1, "secret_1")
	require.NoError(suite.T(), err)

	err = suite.ormStore.DeleteByIdentifier(suite.Ctx, 3, "secret_6")
	require.NoError(suite.T(), err)
}

func (suite *SecretSuite) TestListAllEnabled() {
	tests := []struct {
		spaceId    int64
		wantLength int64
	}{
		{1, 1},
		{2, 3},
		{3, 2},
		{4, 1},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.ListAll(suite.Ctx, test.spaceId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.ListAll(suite.Ctx, test.spaceId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SecretSuite) TestListAncCount() {
	tests := []struct {
		spaceId    int64
		filter     types.ListQueryFilter
		wantLength int64
	}{
		{1, types.ListQueryFilter{}, 1},
		{2, types.ListQueryFilter{}, 3},
		{3, types.ListQueryFilter{}, 2},
		{4, types.ListQueryFilter{}, 1},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, test.spaceId, test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.Count(suite.Ctx, test.spaceId, test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.List(suite.Ctx, test.spaceId, test.filter)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}
