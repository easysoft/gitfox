// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline_test

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/pipeline"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	store2 "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableStage = "stages"
)

type StageSuite struct {
	testsuite.BaseSuite

	ormStore  *pipeline.StageStore
	sqlxStore store.StageStore
}

func TestStageSuite(t *testing.T) {
	ctx := context.Background()

	st := &StageSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "stages",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pipeline.NewStageOrmStore(st.Gdb)
		st.sqlxStore = database.NewStageStore(st.Sdb)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 2, 1, 10)
		ppStore := pipeline.NewPipelineOrmStore(st.Gdb)
		addPipeline(st.Ctx, t, ppStore, 1, 1, 1)
		addPipeline(st.Ctx, t, ppStore, 2, 2, 1)
		addPipeline(st.Ctx, t, ppStore, 3, 1, 2)

		addExecution(&st.BaseSuite, 1, 1, 1, 1)
		addExecution(&st.BaseSuite, 1, 2, 1, 2)
		addExecution(&st.BaseSuite, 2, 3, 1, 1)
	}

	suite.Run(t, st)
}

func (suite *StageSuite) SetupTest() {
	suite.addData()
}

func (suite *StageSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableStage).Where("1 = 1").Delete(nil)
}

func (suite *StageSuite) addData() {
	addItems := []types.Stage{
		{ID: 1, ExecutionID: 1, RepoID: 1, Number: 1, Kind: "test", Status: enum.CIStatusPending},
		{ID: 2, ExecutionID: 1, RepoID: 1, Number: 2, Kind: "lint", Status: enum.CIStatusRunning},
		{ID: 3, ExecutionID: 1, RepoID: 1, Number: 3, Kind: "build", Status: enum.CIStatusWaitingOnDeps},
		{ID: 4, ExecutionID: 2, RepoID: 1, Number: 1, Kind: "test", Status: enum.CIStatusError},
		{ID: 5, ExecutionID: 2, RepoID: 1, Number: 2, Kind: "lint", Status: enum.CIStatusWaitingOnDeps},
		{ID: 6, ExecutionID: 3, RepoID: 1, Number: 1, Kind: "test", Status: enum.CIStatusSuccess},
		{ID: 7, ExecutionID: 3, RepoID: 1, Number: 2, Kind: "build", Status: enum.CIStatusRunning},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *StageSuite) TestCreate() {
	addItems := []types.Stage{
		{ID: 1, ExecutionID: 1, RepoID: 1, Number: 1, Kind: "test", Status: enum.CIStatusPending},
		{ID: 2, ExecutionID: 1, RepoID: 1, Number: 2, Kind: "lint", Status: enum.CIStatusRunning},
		{ID: 3, ExecutionID: 1, RepoID: 1, Number: 3, Kind: "build", Status: enum.CIStatusWaitingOnDeps},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.ErrorIs(suite.T(), err, store2.ErrDuplicate, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *StageSuite) TestFind() {
	for id, pk := range []int64{1, 2, 3, 4, 5, 6, 7} {
		obj, err := suite.ormStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *StageSuite) TestFindByNumber() {
	tests := []struct {
		ID          int64
		ExecutionID int64
		Number      int
	}{
		{1, 1, 1},
		{2, 1, 2},
		{3, 1, 3},
		{4, 2, 1},
		{5, 2, 2},
		{6, 3, 1},
		{7, 3, 2},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.FindByNumber(suite.Ctx, test.ExecutionID, test.Number)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.ID, obj.ID, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.FindByNumber(suite.Ctx, test.ExecutionID, test.Number)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *StageSuite) TestListIncomplete() {
	wantCount := 3
	objs, err := suite.ormStore.ListIncomplete(suite.Ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), wantCount, len(objs))

	objsB, err := suite.sqlxStore.ListIncomplete(suite.Ctx)
	require.NoError(suite.T(), err)
	require.ElementsMatch(suite.T(), objs, objsB)
}

func (suite *StageSuite) TestList() {
	tests := []struct {
		ExecutionID int64
		wantLength  int
	}{
		{1, 3},
		{2, 2},
		{3, 2},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, test.ExecutionID)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.List(suite.Ctx, test.ExecutionID)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *StageSuite) TestListWithSteps() {
	// execution 1 has stage 1,2,3
	// add new step for stage 1
	addStep(&suite.BaseSuite, 1, 1, 1)

	slist1, err := suite.ormStore.ListWithSteps(suite.Ctx, 1)
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 1, len(slist1[0].Steps))
	require.EqualValues(suite.T(), 0, len(slist1[1].Steps))

	slist1B, err := suite.sqlxStore.ListWithSteps(suite.Ctx, 1)
	require.NoError(suite.T(), err)
	require.ElementsMatch(suite.T(), slist1, slist1B)

	// add new step for stage 1
	// add new step for stage 2
	addStep(&suite.BaseSuite, 2, 1, 2)
	addStep(&suite.BaseSuite, 3, 2, 1)

	slist2, err := suite.ormStore.ListWithSteps(suite.Ctx, 1)
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 2, len(slist2[0].Steps))
	require.EqualValues(suite.T(), 1, len(slist2[1].Steps))

	slist2B, err := suite.sqlxStore.ListWithSteps(suite.Ctx, 1)
	require.NoError(suite.T(), err)
	require.ElementsMatch(suite.T(), slist2, slist2B)
}

func addStage(suite *testsuite.BaseSuite, executionID, stageID, repoID, number int64) {
	suite.T().Helper()

	s := pipeline.NewStageOrmStore(suite.Gdb)
	e := types.Stage{
		ID: stageID, ExecutionID: executionID, RepoID: repoID, Number: number, Kind: "test", Status: enum.CIStatusPending,
	}

	err := s.Create(suite.Ctx, &e)
	require.NoError(suite.T(), err)
}
