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
	"github.com/easysoft/gitfox/app/store/database/space"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	store2 "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var _spaceMap = `
|-- main_1 (id:1)
|   |-- sub1_1 (id:2)
|   |   |-- sub2_1 (id:5)
|   |   |-- sub2_2 (id:6)
|   |-- sub1_2 (id:3)
|   |   |-- sub2_1 (id:7)
|   |-- sub1_3 (id:4)
|-- main_2 (id:8)
|   |-- sub1_1 (id:9)
|   |-- sub1_2 (id:10)
|       |-- sub2_1 (id:11)
|       |-- sub2_2 (id:12)
|       |-- sub2_3 (id:13)
`

const (
	testTableSpace     = "spaces"
	testTableSpacePath = "space_paths"
)

type SpaceSuite struct {
	testsuite.BaseSuite

	ormStore  *space.OrmStore
	sqlxStore *database.SpaceStore

	ormPathStore  *space.SpacePathStore
	sqlxPathStore *database.SpacePathStore
}

func TestSpaceSuite(t *testing.T) {
	ctx := context.Background()

	st := &SpaceSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "spaces",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		spacePathTransformation := store.ToLowerSpacePathTransformation

		ormPathStore := space.NewSpacePathOrmStore(st.Gdb, store.ToLowerSpacePathTransformation)
		ormPathCache := cache.New(ormPathStore, spacePathTransformation)

		sqlxPathStore := database.NewSpacePathStore(st.Sdb, store.ToLowerSpacePathTransformation)
		sqlxPathCache := cache.New(sqlxPathStore, spacePathTransformation)

		st.ormStore = space.NewSpaceOrmStore(st.Gdb, ormPathCache, ormPathStore)
		st.sqlxStore = database.NewSpaceStore(st.Sdb, sqlxPathCache, sqlxPathStore)

		st.ormPathStore = ormPathStore
		st.sqlxPathStore = sqlxPathStore

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
	}

	suite.Run(t, st)
}

func (suite *SpaceSuite) SetupTest() {
	suite.addData()
}

func (suite *SpaceSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableSpacePath).Where("1 = 1").Delete(nil)
	suite.Gdb.WithContext(suite.Ctx).Table(testTableSpace).Where("1 = 1").Delete(nil)
}

func (suite *SpaceSuite) addData() {
	_ = _spaceMap

	now := time.Now().UnixMilli()
	addItems := []types.Space{
		{ID: 1, Identifier: "main_1", CreatedBy: 1, Created: now, Updated: now},
		{ID: 2, ParentID: 1, Identifier: "sub1_1", CreatedBy: 1, Created: now, Updated: now},
		{ID: 3, ParentID: 1, Identifier: "sub1_2", CreatedBy: 1, Created: now, Updated: now},
		{ID: 4, ParentID: 1, Identifier: "sub1_3", CreatedBy: 1, Created: now, Updated: now},
		{ID: 5, ParentID: 2, Identifier: "sub2_1", CreatedBy: 1, Created: now, Updated: now},
		{ID: 6, ParentID: 2, Identifier: "sub2_2", CreatedBy: 1, Created: now, Updated: now},
		{ID: 7, ParentID: 3, Identifier: "sub2_1", CreatedBy: 1, Created: now, Updated: now},

		{ID: 8, Identifier: "main_2", CreatedBy: 1, Created: now, Updated: now},
		{ID: 9, ParentID: 8, Identifier: "sub1_1", CreatedBy: 1, Created: now, Updated: now},
		{ID: 10, ParentID: 8, Identifier: "sub1_2", CreatedBy: 1, Created: now, Updated: now},
		{ID: 11, ParentID: 10, Identifier: "sub2_1", CreatedBy: 1, Created: now, Updated: now},
		{ID: 12, ParentID: 10, Identifier: "sub2_2", CreatedBy: 1, Created: now, Updated: now},
		{ID: 13, ParentID: 10, Identifier: "sub2_3", CreatedBy: 1, Created: now, Updated: now},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		segmentPath := buildPathSegment(&item)
		err = suite.ormPathStore.InsertSegment(suite.Ctx, segmentPath)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func buildPathSegment(item *types.Space) *types.SpacePathSegment {
	now := time.Now().UnixMilli()
	return &types.SpacePathSegment{
		Identifier: item.Identifier,
		IsPrimary:  true,
		SpaceID:    item.ID,
		ParentID:   item.ParentID,
		CreatedBy:  item.CreatedBy,
		Created:    now,
		Updated:    now,
	}
}

func (suite *SpaceSuite) TestFind() {
	_ = _spaceMap
	tests := []struct {
		ID         int64
		Identifier string
		ParentId   int64
	}{
		{ID: 1, Identifier: "main_1", ParentId: 0},
		{ID: 2, Identifier: "sub1_1", ParentId: 1},
		{ID: 5, Identifier: "sub2_1", ParentId: 2},
		{ID: 8, Identifier: "main_2", ParentId: 0},
		{ID: 9, Identifier: "sub1_1", ParentId: 8},
		{ID: 11, Identifier: "sub2_1", ParentId: 10},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.Identifier, obj.Identifier, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.ParentId, obj.ParentID, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SpaceSuite) TestFindByRef() {
	_ = _spaceMap
	tests := []struct {
		ID       int64
		Ref      string
		ParentId int64
	}{
		{ID: 1, Ref: "main_1", ParentId: 0},
		{ID: 2, Ref: "main_1/sub1_1", ParentId: 1},
		{ID: 3, Ref: "3", ParentId: 1},
		{ID: 5, Ref: "main_1/sub1_1/sub2_1", ParentId: 2},
		{ID: 8, Ref: "main_2", ParentId: 0},
		{ID: 9, Ref: "main_2/sub1_1", ParentId: 8},
		{ID: 11, Ref: "main_2/sub1_2/sub2_1", ParentId: 10},
		{ID: 12, Ref: "12", ParentId: 10},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.FindByRef(suite.Ctx, test.Ref)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.ID, obj.ID, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.ParentId, obj.ParentID, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.FindByRef(suite.Ctx, test.Ref)
		require.NoError(suite.T(), err)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SpaceSuite) TestGetRootSpace() {
	tests := []struct {
		Id     int64
		RootId int64
	}{
		{Id: 1, RootId: 1}, {Id: 2, RootId: 1}, {Id: 3, RootId: 1}, {Id: 4, RootId: 1},
		{Id: 5, RootId: 1}, {Id: 6, RootId: 1}, {Id: 7, RootId: 1},

		{Id: 8, RootId: 8}, {Id: 9, RootId: 8}, {Id: 10, RootId: 8},
		{Id: 11, RootId: 8}, {Id: 12, RootId: 8}, {Id: 13, RootId: 8},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.GetRootSpace(suite.Ctx, test.Id)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.RootId, obj.ID, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.GetRootSpace(suite.Ctx, test.Id)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.RootId, objB.ID, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SpaceSuite) TestUpdate() {
	_ = _spaceMap
	now := time.Now().UnixMilli() - 1

	updates := []types.Space{
		// Version and Updated will be increase always
		// Updated value will be ignored and set new in Update func
		{ID: 1, Identifier: "main_1", CreatedBy: 1, Created: now, Updated: now},               // nothing
		{ID: 5, ParentID: 2, Identifier: "sub2_1a", CreatedBy: 1, Created: now, Updated: now}, // change identifier
		{ID: 6, ParentID: 2, Identifier: "sub2_2", CreatedBy: 1, Created: now},                // ignore Updated field
		{ID: 7, ParentID: 3, Identifier: "sub2_1", CreatedBy: 1, Created: now, Updated: now},  // change IsPublic
		{ID: 11, ParentID: 9, Identifier: "sub2_1", CreatedBy: 1, Created: now, Updated: now}, // change parent
	}

	updates2 := []types.Space{
		{ID: 1, Version: 1, Identifier: "main_1", CreatedBy: 1, Created: now, Updated: now + 60*1000}, // update again, version is required
		{ID: 13, ParentID: 4, Identifier: "sub2_0", CreatedBy: 1, Created: now},                       // update parentId, Identifier
	}

	for id, up := range updates {
		err := suite.ormStore.Update(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, up := range updates2 {
		err := suite.sqlxStore.Update(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	tests := []struct {
		ID        int64
		fieldName string
		expect    interface{}
	}{
		{ID: 1, expect: 2, fieldName: "Version"},
		{ID: 5, expect: "sub2_1a", fieldName: "Identifier"},
		{ID: 6, expect: "sub2_2", fieldName: "Identifier"},
		{ID: 7, expect: false, fieldName: "IsPublic"},
		{ID: 11, expect: 9, fieldName: "ParentID"},
		{ID: 13, expect: 4, fieldName: "ParentID"},
		{ID: 13, expect: "sub2_0", fieldName: "Identifier"},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		testsuite.EqualFieldValue(suite.T(), test.expect, obj, test.fieldName, testsuite.InvalidLoopMsgF, id)
		require.Greater(suite.T(), obj.Version, int64(0))
		require.Greater(suite.T(), obj.Updated, now)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SpaceSuite) TestFindForUpdate() {
	var testId int64 = 1
	obj, err := suite.ormStore.FindForUpdate(suite.Ctx, testId)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), testId, obj.ID)

	objB, err := suite.sqlxStore.FindForUpdate(suite.Ctx, testId)
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), obj, objB)
}

func (suite *SpaceSuite) TestList() {
	_ = _spaceMap
	tests := []struct {
		Id         int64
		filter     types.SpaceFilter
		wantLength int
	}{
		{Id: 1, filter: types.SpaceFilter{}, wantLength: 3},
		{Id: 2, filter: types.SpaceFilter{}, wantLength: 2},
		{Id: 6, filter: types.SpaceFilter{}, wantLength: 0},
		{Id: 1, filter: types.SpaceFilter{Recursive: true}, wantLength: 6},
		//{Id: 0, filter: types.SpaceFilter{Recursive: true}, wantLength: 13},
		{Id: 1, filter: types.SpaceFilter{Page: 1, Size: 2}, wantLength: 2},
		{Id: 1, filter: types.SpaceFilter{Page: 2, Size: 2}, wantLength: 1},
		{Id: 1, filter: types.SpaceFilter{Page: 3, Size: 2}, wantLength: 0},
		{Id: 10, filter: types.SpaceFilter{Query: "sub"}, wantLength: 3},
		{Id: 10, filter: types.SpaceFilter{Query: "sub2"}, wantLength: 3},
		{Id: 10, filter: types.SpaceFilter{Query: "sub3"}, wantLength: 0},
		{Id: 10, filter: types.SpaceFilter{Page: 1, Size: 2, Sort: enum.SpaceAttrCreated}, wantLength: 2},
		{Id: 10, filter: types.SpaceFilter{Page: 1, Size: 2, Sort: enum.SpaceAttrUID}, wantLength: 2},
		{Id: 10, filter: types.SpaceFilter{Page: 1, Size: 2, Sort: enum.SpaceAttrIdentifier}, wantLength: 2},
		{Id: 10, filter: types.SpaceFilter{Page: 1, Size: 2, Sort: enum.SpaceAttrIdentifier, Order: enum.OrderDesc}, wantLength: 2},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, test.Id, &test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.List(suite.Ctx, test.Id, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		if test.wantLength > 0 {
			require.ElementsMatch(suite.T(), objs, objsB)
		}
	}
}

func (suite *SpaceSuite) TestCount() {
	_ = _spaceMap
	tests := []struct {
		Id         int64
		filter     types.SpaceFilter
		wantLength int
	}{
		{Id: 1, filter: types.SpaceFilter{}, wantLength: 3},
		{Id: 2, filter: types.SpaceFilter{}, wantLength: 2},
		{Id: 6, filter: types.SpaceFilter{}, wantLength: 0},
		{Id: 1, filter: types.SpaceFilter{Recursive: true}, wantLength: 6},
		//{Id: 0, filter: types.SpaceFilter{Recursive: true}, wantLength: 13},
		{Id: 1, filter: types.SpaceFilter{Page: 1, Size: 2}, wantLength: 3},
		{Id: 1, filter: types.SpaceFilter{Page: 2, Size: 2}, wantLength: 3},
		{Id: 1, filter: types.SpaceFilter{Page: 3, Size: 2}, wantLength: 3},
		{Id: 10, filter: types.SpaceFilter{Query: "sub"}, wantLength: 3},
		{Id: 10, filter: types.SpaceFilter{Query: "sub2"}, wantLength: 3},
		{Id: 10, filter: types.SpaceFilter{Query: "sub3"}, wantLength: 0},
	}

	for id, test := range tests {
		count, err := suite.ormStore.Count(suite.Ctx, test.Id, &test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		countB, err := suite.sqlxStore.Count(suite.Ctx, test.Id, &test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, countB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *SpaceSuite) TestSoftDeleteAndRestore() {
	_ = _spaceMap
	// soft delete id 7
	space7, err := suite.ormStore.Find(suite.Ctx, 7)
	require.NoError(suite.T(), err)

	deleteAt := time.Now().UnixMilli()
	err = suite.ormStore.SoftDelete(suite.Ctx, space7, deleteAt)
	require.NoError(suite.T(), err)

	// restore id 7 to 10
	space7, err = suite.ormStore.FindByRefAndDeletedAt(suite.Ctx, "main_1/sub1_2/sub2_1", deleteAt)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), space7.Deleted)

	restoreIdentifier := space7.Identifier + "_restored"
	restoreParentId := int64(10)

	restore7Obj, err := suite.ormStore.Restore(suite.Ctx, space7, &restoreIdentifier, &restoreParentId)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), restoreIdentifier, restore7Obj.Identifier)
	require.Equal(suite.T(), space7.Description, restore7Obj.Description)
}

func (suite *SpaceSuite) TestPurge() {
	// soft delete id 11
	space11, err := suite.ormStore.Find(suite.Ctx, 11)
	require.NoError(suite.T(), err)

	deleteAt := time.Now().UnixMilli()
	err = suite.ormStore.SoftDelete(suite.Ctx, space11, deleteAt)
	require.NoError(suite.T(), err)

	// Purge id 11, 12
	err = suite.ormStore.Purge(suite.Ctx, 11, &deleteAt)
	require.NoError(suite.T(), err)

	err = suite.ormStore.Purge(suite.Ctx, 12, nil)
	require.NoError(suite.T(), err)

	for id, pk := range []int64{11, 12} {
		_, err = suite.ormStore.Find(suite.Ctx, pk)
		require.ErrorIs(suite.T(), err, store2.ErrResourceNotFound, testsuite.InvalidLoopMsgF, id)
	}
}
