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
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableTemplate = "templates"
)

type TemplateSuite struct {
	testsuite.BaseSuite

	ormStore  *space.TemplateStore
	sqlxStore store.TemplateStore
}

func TestTemplateSuite(t *testing.T) {
	ctx := context.Background()

	st := &TemplateSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "templates",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = space.NewTemplateOrmStore(st.Gdb)
		st.sqlxStore = database.NewTemplateStore(st.Sdb)

		// add init data
		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 2, 1)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 3, 1)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 4, 2)
	}

	suite.Run(t, st)
}

func (suite *TemplateSuite) SetupTest() {
	suite.addData()
}

func (suite *TemplateSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableTemplate).Where("1 = 1").Delete(nil)
}

var testAddTemplateItems = []struct {
	id         int64
	tplType    enum.ResolverType
	spaceId    int64
	identifier string
}{
	{id: 1, tplType: enum.ResolverTypeStage, spaceId: 1, identifier: fmt.Sprintf("template_1")},
	{id: 2, tplType: enum.ResolverTypeStage, spaceId: 2, identifier: fmt.Sprintf("template_2")},
	{id: 3, tplType: enum.ResolverTypeStep, spaceId: 2, identifier: fmt.Sprintf("template_3")},
	{id: 4, tplType: enum.ResolverTypeStage, spaceId: 2, identifier: fmt.Sprintf("template_4")},
	{id: 5, tplType: enum.ResolverTypeStep, spaceId: 3, identifier: fmt.Sprintf("template_5")},
	{id: 6, tplType: enum.ResolverTypeStage, spaceId: 3, identifier: fmt.Sprintf("template_6")},
	{id: 7, tplType: enum.ResolverTypeStep, spaceId: 4, identifier: fmt.Sprintf("template_7")},
}

func (suite *TemplateSuite) addData() {
	now := time.Now().UnixMilli()
	for id, item := range testAddTemplateItems {
		obj := types.Template{
			ID:         item.id,
			Type:       item.tplType,
			SpaceID:    item.spaceId,
			Identifier: item.identifier,
			Data:       "",
			Created:    now,
			Updated:    now,
		}

		err := suite.ormStore.Create(suite.Ctx, &obj)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TemplateSuite) TestFind() {
	for id, test := range testAddTemplateItems {
		obj, err := suite.ormStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.identifier, obj.Identifier)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TemplateSuite) TestFindByIdentifier() {
	for id, test := range testAddTemplateItems {
		obj, err := suite.ormStore.FindByIdentifierAndType(suite.Ctx, test.spaceId, test.identifier, test.tplType)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.id, obj.ID)

		objB, err := suite.sqlxStore.FindByIdentifierAndType(suite.Ctx, test.spaceId, test.identifier, test.tplType)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TemplateSuite) TestDelete() {
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

func (suite *TemplateSuite) TestDeleteByIdentifier() {
	err := suite.ormStore.DeleteByIdentifierAndType(suite.Ctx, 1, "template_1", enum.ResolverTypeStage)
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.DeleteByIdentifierAndType(suite.Ctx, 3, "template_6", enum.ResolverTypeStage)
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.FindByIdentifierAndType(suite.Ctx, 3, "template_6", enum.ResolverTypeStage)
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.FindByIdentifierAndType(suite.Ctx, 1, "template_1", enum.ResolverTypeStage)
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.DeleteByIdentifierAndType(suite.Ctx, 1, "template_1", enum.ResolverTypeStage)
	require.NoError(suite.T(), err)

	err = suite.ormStore.DeleteByIdentifierAndType(suite.Ctx, 3, "template_6", enum.ResolverTypeStage)
	require.NoError(suite.T(), err)
}

func (suite *TemplateSuite) TestListAncCount() {
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
