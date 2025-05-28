// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/cache"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/repo"
	"github.com/easysoft/gitfox/app/store/database/space"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableRepo = "repositories"
)

var _repoMap = `
|-- space_1
|   |-> repo_1
|   |-- space_2
|   |   |-> repo_3
|   |   |-> repo_4
|   |   |-- space_4
|   |   |   |-> repo_5
|   |   |   |-> repo_7
|   |   |-- space_5
|   |   |   |-> repo_6
|   |   |   |-> repo_8
|   |-- space_3
|   |   |-> repo_2
|   |   |-> repo_9
`

type RepoSuite struct {
	testsuite.BaseSuite

	ormStore  *repo.OrmStore
	sqlxStore *database.RepoStore
}

func TestRepoSuite(t *testing.T) {
	ctx := context.Background()

	st := &RepoSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "repos",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		spacePathTransformation := store.ToLowerSpacePathTransformation

		ormPathStore := space.NewSpacePathOrmStore(st.Gdb, store.ToLowerSpacePathTransformation)
		ormPathCache := cache.New(ormPathStore, spacePathTransformation)

		sqlxPathStore := database.NewSpacePathStore(st.Sdb, store.ToLowerSpacePathTransformation)
		sqlxPathCache := cache.New(sqlxPathStore, spacePathTransformation)

		ormSpaceStore := space.NewSpaceOrmStore(st.Gdb, ormPathCache, ormPathStore)
		sqlxSpaceStore := database.NewSpaceStore(st.Sdb, sqlxPathCache, sqlxPathStore)

		st.ormStore = repo.NewRepoOrmStore(st.Gdb, ormPathCache, ormPathStore, ormSpaceStore)
		st.sqlxStore = database.NewRepoStore(st.Sdb, sqlxPathCache, sqlxPathStore, sqlxSpaceStore)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, true)

		testsuite.AddSpace(st.Ctx, t, ts.Space, ormPathStore, 1, 1, 0)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ormPathStore, 1, 2, 1)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ormPathStore, 1, 3, 1)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ormPathStore, 1, 4, 2)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ormPathStore, 2, 5, 2)
	}

	suite.Run(t, st)
}

func (suite *RepoSuite) SetupTest() {
	suite.addData()
}

func (suite *RepoSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableRepo).Where("1 = 1").Delete(nil)
}

func (suite *RepoSuite) addData() {
	now := time.Now().UnixMilli()
	addItems := []types.Repository{
		{ID: 1, ParentID: 1, Identifier: "repo_1", GitUID: genUID(), CreatedBy: 1, Created: now, Updated: now},
		{ID: 2, ParentID: 3, Identifier: "repo_2", GitUID: genUID(), CreatedBy: 1, Created: now, Updated: now},
		{ID: 3, ParentID: 2, Identifier: "repo_3", GitUID: genUID(), CreatedBy: 1, Created: now, Updated: now},
		{ID: 4, ParentID: 2, Identifier: "repo_4", GitUID: genUID(), CreatedBy: 1, Created: now, Updated: now},
		{ID: 5, ParentID: 4, Identifier: "repo_5", GitUID: genUID(), CreatedBy: 1, Created: now, Updated: now},
		{ID: 6, ParentID: 5, Identifier: "repo_6", GitUID: genUID(), CreatedBy: 1, Created: now, Updated: now},
		{ID: 7, ParentID: 4, Identifier: "repo_7", GitUID: genUID(), CreatedBy: 1, Created: now, Updated: now},
		{ID: 8, ParentID: 5, Identifier: "repo_8", GitUID: genUID(), CreatedBy: 2, Created: now, Updated: now},
		{ID: 9, ParentID: 3, Identifier: "repo_9", GitUID: genUID(), CreatedBy: 2, Created: now, Updated: now},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RepoSuite) TestFind() {
	_ = _repoMap
	tests := []struct {
		ID         int64
		ParentID   int64
		Identifier string
	}{
		{ID: 1, ParentID: 1, Identifier: "repo_1"},
		{ID: 2, ParentID: 3, Identifier: "repo_2"},
		{ID: 3, ParentID: 2, Identifier: "repo_3"},
		{ID: 5, ParentID: 4, Identifier: "repo_5"},
		{ID: 8, ParentID: 5, Identifier: "repo_8"},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.ParentID, obj.ParentID, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.Identifier, obj.Identifier, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RepoSuite) TestFindByRef() {
	_ = _repoMap
	tests := []struct {
		ID       int64
		ParentID int64
		Ref      string
	}{
		{ID: 1, ParentID: 1, Ref: "space_1/repo_1"},
		{ID: 2, ParentID: 3, Ref: "space_1/space_3/repo_2"},
		{ID: 3, ParentID: 2, Ref: "space_1/space_2/repo_3"},
		{ID: 5, ParentID: 4, Ref: "space_1/space_2/space_4/repo_5"},
		{ID: 8, ParentID: 5, Ref: "space_1/space_2/space_5/repo_8"},
		{ID: 9, ParentID: 3, Ref: "9"},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.FindByRef(suite.Ctx, test.Ref)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.ID, obj.ID, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.ParentID, obj.ParentID, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.FindByRef(suite.Ctx, test.Ref)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func genUID() string {
	uid, _ := git.NewRepositoryUID()
	return uid
}

func (suite *RepoSuite) TestUpdate() {
	_ = _repoMap
	now := time.Now().UnixMilli() - 1

	updates := []types.Repository{
		// Version and Updated will be increase always
		// Updated value will be ignored and set new in Update func
		{ID: 1, ParentID: 1, Identifier: "repo_1", GitUID: genUID(), CreatedBy: 1, Created: now, Deleted: &now},         // soft delete
		{ID: 2, ParentID: 2, Identifier: "repo_2", GitUID: genUID(), CreatedBy: 1, Created: now},                        // update parentId
		{ID: 3, ParentID: 2, Identifier: "repo_31", GitUID: genUID(), CreatedBy: 1, Created: now},                       // update Identifier
		{ID: 4, ParentID: 2, Identifier: "repo_4", CreatedBy: 1, Created: now, GitUID: "azxcdsqwerfv"},                  // update GitUID
		{ID: 5, ParentID: 4, Identifier: "repo_5", GitUID: genUID(), CreatedBy: 1, Created: now, DefaultBranch: "main"}, // update default branch
		{ID: 6, ParentID: 5, Identifier: "repo_6", GitUID: genUID(), CreatedBy: 1, Created: now, NumPulls: 10},          // update NumPulls
	}

	updates2 := []types.Repository{
		{ID: 3, ParentID: 2, Version: 1, Identifier: "repo_31", GitUID: genUID(), CreatedBy: 1, Created: now, Updated: now + 60*1000}, // update again, version is required
		{ID: 9, ParentID: 4, Identifier: "repo_9", GitUID: genUID(), CreatedBy: 1, Created: now},                                      // update parentId
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
		{ID: 2, expect: 2, fieldName: "ParentID"},
		{ID: 3, expect: "repo_31", fieldName: "Identifier"},
		{ID: 3, expect: 2, fieldName: "Version"},
		{ID: 4, expect: "azxcdsqwerfv", fieldName: "GitUID"},
		{ID: 5, expect: "main", fieldName: "DefaultBranch"},
		{ID: 6, expect: 10, fieldName: "NumPulls"},
		{ID: 9, expect: 4, fieldName: "ParentID"},
	}

	_, e := suite.ormStore.Find(suite.Ctx, 1)
	require.Error(suite.T(), e)

	_, e = suite.sqlxStore.Find(suite.Ctx, 1)
	require.Error(suite.T(), e)

	for id, test := range tests {
		obj, err := suite.ormStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		testsuite.EqualFieldValue(suite.T(), test.expect, obj, test.fieldName, testsuite.InvalidLoopMsgF, id)
		require.Greater(suite.T(), obj.Version, int64(0))
		require.Greater(suite.T(), obj.Updated, now)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RepoSuite) TestUpdateSize() {
	now := time.Now().UnixMilli() - 1
	tests := []struct {
		Id   int64
		Size int64
	}{
		{Id: 1, Size: int64(rand.Intn(2 << 16))},
		{Id: 2, Size: int64(rand.Intn(2 << 16))},
		{Id: 3, Size: int64(rand.Intn(2 << 16))},
	}

	for id, test := range tests {
		err := suite.ormStore.UpdateSize(suite.Ctx, test.Id, test.Size)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, test := range tests {
		size, err := suite.ormStore.GetSize(suite.Ctx, test.Id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.Size, size)

		obj, err := suite.ormStore.Find(suite.Ctx, test.Id)
		require.NoError(suite.T(), err)
		require.Greater(suite.T(), obj.SizeUpdated, now, testsuite.InvalidLoopMsgF, id)

		sizeB, err := suite.sqlxStore.GetSize(suite.Ctx, test.Id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.Size, sizeB)
	}
}

func (suite *RepoSuite) TestSoftDelete() {
	now := time.Now().UnixMilli()
	tests := []struct {
		Id      int64
		Deleted int64
	}{
		{Id: 1, Deleted: now - 1},
		{Id: 2, Deleted: now - 2},
		{Id: 3, Deleted: now - 3},
	}

	for id, test := range tests {
		obj, err := suite.provideStore(id).Find(suite.Ctx, test.Id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		err = suite.provideStore(id).SoftDelete(suite.Ctx, obj, test.Deleted)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, test := range tests {
		obj, err := suite.ormStore.FindByRefAndDeletedAt(suite.Ctx, fmt.Sprintf("%d", test.Id), test.Deleted)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.Id, obj.ID)

		objB, err := suite.sqlxStore.FindByRefAndDeletedAt(suite.Ctx, fmt.Sprintf("%d", test.Id), test.Deleted)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RepoSuite) provideStore(id int) store.RepoStore {
	if id%2 == 0 {
		return suite.sqlxStore
	}
	return suite.ormStore
}

func (suite *RepoSuite) TestPruneAndRestore() {
	now := time.Now().UnixMilli()

	softDeleteIds := []int64{1, 4, 5, 8}
	hardDeleteIds := []int64{2, 3, 6, 7}

	for _, pk := range softDeleteIds {
		obj, err := suite.ormStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err)

		err = suite.ormStore.SoftDelete(suite.Ctx, obj, now)
		require.NoError(suite.T(), err)
	}

	// pruned objects can't be restored
	for id, pk := range append(hardDeleteIds) {
		obj, err := suite.ormStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err)

		err = suite.provideStore(id).Purge(suite.Ctx, pk, nil)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		_, err = suite.ormStore.Find(suite.Ctx, pk)
		require.Error(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		_, err = suite.sqlxStore.Find(suite.Ctx, pk)
		require.Error(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		obj.Deleted = &now
		//_, err = suite.provideStore(id).Restore(suite.Ctx, obj, obj.Identifier+"_restore")
		//require.Error(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	// prune will remove the soft-deleted objects
	for id, pk := range softDeleteIds[2:] {
		err := suite.provideStore(id).Purge(suite.Ctx, pk, &now)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		_, err = suite.provideStore(id).FindByRefAndDeletedAt(suite.Ctx, fmt.Sprintf("%d", pk), now)
		require.Error(suite.T(), err)
	}

	// only soft deleted repos can be restored
	for id, pk := range softDeleteIds[:2] {
		_, err := suite.provideStore(id).FindByRefAndDeletedAt(suite.Ctx, fmt.Sprintf("%d", pk), now)
		require.NoError(suite.T(), err)

		//_, err = suite.provideStore(id).Restore(suite.Ctx, obj, obj.Identifier+"_restore")
		//require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RepoSuite) TestCount() {
	_ = _repoMap

	now := time.Now().UnixMilli()

	// set repo_8 to deleted
	obj, _ := suite.ormStore.Find(suite.Ctx, 8)
	_ = suite.ormStore.SoftDelete(suite.Ctx, obj, now-1)

	tests := []struct {
		parentId   int64
		filter     types.RepoFilter
		wantLength int64
	}{
		{1, types.RepoFilter{}, 1},
		{2, types.RepoFilter{}, 2},
		{3, types.RepoFilter{}, 2},
		{4, types.RepoFilter{}, 2},
		{5, types.RepoFilter{}, 1},
		{1, types.RepoFilter{Recursive: true}, 8},
		{2, types.RepoFilter{Recursive: true}, 5},
		{3, types.RepoFilter{Recursive: true}, 2},
		{4, types.RepoFilter{Recursive: true}, 2},
		//{1, types.RepoFilter{Recursive: true, Page: 1, Size: 5}, 5},
		//{1, types.RepoFilter{Recursive: true, Page: 2, Size: 5}, 3},
		//{1, types.RepoFilter{Recursive: true, Page: 3, Size: 5}, 0},
		//{1, types.RepoFilter{Recursive: true, Page: 1, Size: 5, DeletedBeforeOrAt: &now}, 1},
		//{1, types.RepoFilter{Recursive: true, Page: 2, Size: 5, Sort: enum.RepoAttrIdentifier}, 3},
		//{1, types.RepoFilter{Recursive: true, Page: 2, Size: 5, Sort: enum.RepoAttrIdentifier, Order: enum.OrderDesc}, 3},
	}

	for id, test := range tests {
		count, err := suite.ormStore.Count(suite.Ctx, test.parentId, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		countB, err := suite.sqlxStore.Count(suite.Ctx, test.parentId, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.wantLength, countB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RepoSuite) TestList() {
	_ = _repoMap

	now := time.Now().UnixMilli()

	// set repo_8 to deleted
	obj, _ := suite.ormStore.Find(suite.Ctx, 8)
	_ = suite.ormStore.SoftDelete(suite.Ctx, obj, now-1)

	tests := []struct {
		parentId   int64
		filter     types.RepoFilter
		wantLength int64
	}{
		{1, types.RepoFilter{}, 1},
		{2, types.RepoFilter{}, 2},
		{3, types.RepoFilter{}, 2},
		{4, types.RepoFilter{}, 2},
		{5, types.RepoFilter{}, 1},
		{1, types.RepoFilter{Recursive: true}, 8},
		{2, types.RepoFilter{Recursive: true}, 5},
		{3, types.RepoFilter{Recursive: true}, 2},
		{4, types.RepoFilter{Recursive: true}, 2},
		{1, types.RepoFilter{Recursive: true, Page: 1, Size: 5}, 5},
		{1, types.RepoFilter{Recursive: true, Page: 2, Size: 5}, 3},
		{1, types.RepoFilter{Recursive: true, Page: 3, Size: 5}, 0},
		{1, types.RepoFilter{Recursive: true, Page: 1, Size: 5, DeletedBeforeOrAt: &now}, 1},
		{1, types.RepoFilter{Recursive: true, Page: 2, Size: 5, Sort: enum.RepoAttrIdentifier}, 3},
		{1, types.RepoFilter{Recursive: true, Page: 2, Size: 5, Sort: enum.RepoAttrIdentifier, Order: enum.OrderDesc}, 3},
	}

	for id, test := range tests {
		repos, err := suite.ormStore.List(suite.Ctx, test.parentId, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(repos), testsuite.InvalidLoopMsgF, id)

		reposB, err := suite.sqlxStore.List(suite.Ctx, test.parentId, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), repos, reposB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RepoSuite) TestListCountMulti() {
	_ = _repoMap

	now := time.Now().UnixMilli()

	// set repo_8 to deleted
	obj, _ := suite.ormStore.Find(suite.Ctx, 8)
	_ = suite.ormStore.SoftDelete(suite.Ctx, obj, now-1)

	tests := []struct {
		parentIds  []int64
		filter     types.RepoFilter
		wantLength int64
	}{
		{[]int64{1}, types.RepoFilter{}, 1},
		{[]int64{2}, types.RepoFilter{}, 2},
		{[]int64{1, 2, 3}, types.RepoFilter{}, 5},
		{[]int64{1}, types.RepoFilter{Recursive: true}, 8},
		{[]int64{2, 3}, types.RepoFilter{Recursive: true}, 7},
	}

	for id, test := range tests {
		repos, err := suite.ormStore.ListMulti(suite.Ctx, test.parentIds, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(repos), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.CountMulti(suite.Ctx, test.parentIds, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RepoSuite) TestListAllWithoutParentAndCount() {
	count, err := suite.ormStore.CountAllWithoutParent(suite.Ctx, &types.RepoFilter{})
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 9, count)

	objs, err := suite.ormStore.ListAllWithoutParent(suite.Ctx, &types.RepoFilter{})
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 9, len(objs))
}

func (suite *RepoSuite) TestListSizeInfos() {
	_ = _repoMap
	tests := []struct {
		Id   int64
		Size int64
	}{
		{Id: 1, Size: int64(rand.Intn(2 << 16))},
		{Id: 2, Size: int64(rand.Intn(2 << 16))},
		{Id: 3, Size: int64(rand.Intn(2 << 16))},
		{Id: 4, Size: int64(rand.Intn(2 << 16))},
		{Id: 5, Size: int64(rand.Intn(2 << 16))},
	}

	for id, test := range tests {
		err := suite.ormStore.UpdateSize(suite.Ctx, test.Id, test.Size)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	infos, err := suite.ormStore.ListSizeInfos(suite.Ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 9, len(infos))

	infosB, err := suite.sqlxStore.ListSizeInfos(suite.Ctx)
	require.NoError(suite.T(), err)
	require.ElementsMatch(suite.T(), infos, infosB)
}

func (suite *RepoSuite) TestInfoViewFind() {
	viewOrm := repo.NewRepoGitOrmInfoView(suite.Gdb)
	viewSqlx := database.NewRepoGitInfoView(suite.Sdb)

	for _, pk := range []int64{1, 2, 3} {
		obj, err := viewOrm.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err)

		objB, err := viewSqlx.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err)
		require.EqualExportedValues(suite.T(), *obj, *objB)
	}
}
