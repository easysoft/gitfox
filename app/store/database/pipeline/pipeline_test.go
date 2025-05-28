// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/pipeline"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	store2 "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTablePipeline = "pipelines"
)

type PipelineSuite struct {
	testsuite.BaseSuite

	ormStore  *pipeline.OrmStore
	sqlxStore store.PipelineStore
}

func TestPipelineSuite(t *testing.T) {
	ctx := context.Background()

	st := &PipelineSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "pipelines",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pipeline.NewPipelineOrmStore(st.Gdb)
		st.sqlxStore = database.NewPipelineStore(st.Sdb)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 2, 1, 10)
	}

	suite.Run(t, st)
}

func (suite *PipelineSuite) SetupTest() {
	suite.addData()
}

func (suite *PipelineSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTablePipeline).Where("1 = 1").Delete(nil)
}

func (suite *PipelineSuite) addData() {
	addItems := []types.Pipeline{
		{ID: 1, RepoID: 1, Identifier: "pp_1", CreatedBy: 1, Seq: 1, DefaultBranch: "main"},
		{ID: 2, RepoID: 1, Identifier: "pp_2", CreatedBy: 2, Seq: 1, DefaultBranch: "test"},
		{ID: 3, RepoID: 2, Identifier: "pp_1", CreatedBy: 2, Seq: 1, DefaultBranch: "feat1"},
		{ID: 4, RepoID: 2, Identifier: "pp_2", CreatedBy: 1, Seq: 1, DefaultBranch: "main"},
		{ID: 5, RepoID: 2, Identifier: "ppd1", CreatedBy: 1, Seq: 1, DefaultBranch: "test", Disabled: true},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PipelineSuite) TestCreate() {
	addItems := []types.Pipeline{
		{ID: 1, RepoID: 1, Identifier: "pp_1", CreatedBy: 1, Seq: 1, DefaultBranch: "main"},
		{ID: 2, RepoID: 1, Identifier: "pp_2", CreatedBy: 2, Seq: 1, DefaultBranch: "test"},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.ErrorIs(suite.T(), err, store2.ErrDuplicate, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PipelineSuite) TestFind() {
	for id, pk := range []int64{1, 2, 3, 4, 5} {
		obj, err := suite.ormStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PipelineSuite) TestFindByIdentifier() {
	tests := []struct {
		Id         int64
		RepoId     int64
		Identifier string
	}{
		{Id: 1, RepoId: 1, Identifier: "pp_1"},
		{Id: 2, RepoId: 1, Identifier: "pp_2"},
		{Id: 3, RepoId: 2, Identifier: "pp_1"},
		{Id: 4, RepoId: 2, Identifier: "pp_2"},
		{Id: 5, RepoId: 2, Identifier: "ppd1"},
	}
	for id, test := range tests {
		obj, err := suite.ormStore.FindByIdentifier(suite.Ctx, test.RepoId, test.Identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.Id, obj.ID, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.FindByIdentifier(suite.Ctx, test.RepoId, test.Identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PipelineSuite) TestListAndCount() {
	addPipeline(suite.Ctx, suite.T(), suite.ormStore, 11, 2, 1)
	addPipeline(suite.Ctx, suite.T(), suite.ormStore, 12, 2, 2)
	addPipeline(suite.Ctx, suite.T(), suite.ormStore, 13, 1, 1)

	tests := []struct {
		repoId     int64
		filter     types.ListQueryFilter
		wantLength int64
	}{
		{1, types.ListQueryFilter{}, 3},
		{2, types.ListQueryFilter{}, 5},
		{3, types.ListQueryFilter{}, 0},
		{2, types.ListQueryFilter{Query: "pipeline"}, 2},
		{2, types.ListQueryFilter{Query: "pp"}, 3},
		{2, types.ListQueryFilter{Query: "ppd"}, 1},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, test.repoId, test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.Count(suite.Ctx, test.repoId, test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.List(suite.Ctx, test.repoId, test.filter)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PipelineSuite) TestDelete() {
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

func (suite *PipelineSuite) TestDeleteByIdentifier() {
	//{Id: 2, RepoId: 1, Identifier: "pp_2"},
	//{Id: 3, RepoId: 2, Identifier: "pp_1"},
	err := suite.ormStore.DeleteByIdentifier(suite.Ctx, 1, "pp_2")
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, 2, "pp_1")
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.Find(suite.Ctx, 3)
	require.ErrorIs(suite.T(), err, store2.ErrResourceNotFound)

	_, err = suite.sqlxStore.Find(suite.Ctx, 2)
	require.ErrorIs(suite.T(), err, store2.ErrResourceNotFound)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, 1, "pp_2")
	require.NoError(suite.T(), err)

	err = suite.ormStore.DeleteByIdentifier(suite.Ctx, 2, "pp_1")
	require.NoError(suite.T(), err)
}

func (suite *PipelineSuite) TestIncrementSeqNum() {
	p1, _ := suite.ormStore.Find(suite.Ctx, 1)
	require.EqualValues(suite.T(), 1, p1.Seq)

	p2, err := suite.ormStore.IncrementSeqNum(suite.Ctx, p1)
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 2, p2.Seq)

	p3, err := suite.ormStore.IncrementSeqNum(suite.Ctx, p2)
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 3, p3.Seq)
}

func (suite *PipelineSuite) TestListLatest() {
	// repo1 has pipeline1,2
	// exec pipeline1
	addExecution(&suite.BaseSuite, 1, 1, 1, 1)

	plist1, err := suite.ormStore.ListLatest(suite.Ctx, 1, types.ListQueryFilter{})
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 1, plist1[0].Execution.ID)
	require.Nil(suite.T(), plist1[1].Execution)

	plist1B, err := suite.sqlxStore.ListLatest(suite.Ctx, 1, types.ListQueryFilter{})
	require.NoError(suite.T(), err)
	require.ElementsMatch(suite.T(), plist1, plist1B)

	// exec pipeline1 and pipeline2
	addExecution(&suite.BaseSuite, 1, 2, 1, 3)
	addExecution(&suite.BaseSuite, 2, 3, 1, 1)

	plist2, err := suite.ormStore.ListLatest(suite.Ctx, 1, types.ListQueryFilter{})
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 2, plist2[0].Execution.ID)
	require.EqualValues(suite.T(), 3, plist2[1].Execution.ID)

	plist2B, err := suite.sqlxStore.ListLatest(suite.Ctx, 1, types.ListQueryFilter{})
	require.NoError(suite.T(), err)
	require.ElementsMatch(suite.T(), plist2, plist2B)
}

func addPipeline(ctx context.Context, t *testing.T, s store.PipelineStore, id, repoId, createdBy int64) {
	pp := types.Pipeline{
		ID:            id,
		Identifier:    "pipeline_" + strconv.Itoa(int(id)),
		CreatedBy:     createdBy,
		Seq:           1,
		RepoID:        repoId,
		DefaultBranch: "main",
	}
	err := s.Create(ctx, &pp)
	require.NoError(t, err)
}
