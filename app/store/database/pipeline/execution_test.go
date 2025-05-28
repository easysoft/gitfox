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

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableExecution = "executions"
)

type ExecutionSuite struct {
	testsuite.BaseSuite

	ormStore  *pipeline.ExecutionStore
	sqlxStore store.ExecutionStore
}

func TestExecutionSuite(t *testing.T) {
	ctx := context.Background()

	st := &ExecutionSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "executions",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pipeline.NewExecutionOrmStore(st.Gdb)
		st.sqlxStore = database.NewExecutionStore(st.Sdb)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 2, 1, 10)
		ppStore := pipeline.NewPipelineOrmStore(st.Gdb)
		addPipeline(st.Ctx, t, ppStore, 1, 1, 1)
		addPipeline(st.Ctx, t, ppStore, 2, 2, 1)
		addPipeline(st.Ctx, t, ppStore, 3, 1, 2)
	}

	suite.Run(t, st)
}

func (suite *ExecutionSuite) SetupTest() {
	suite.addData()
}

func (suite *ExecutionSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableExecution).Where("1 = 1").Delete(nil)
}

func (suite *ExecutionSuite) addData() {
	addItems := []types.Execution{
		{ID: 1, PipelineID: 1, CreatedBy: 1, RepoID: 1, Number: 1},
		{ID: 2, PipelineID: 1, CreatedBy: 1, RepoID: 1, Number: 2},
		{ID: 3, PipelineID: 2, CreatedBy: 1, RepoID: 2, Number: 1},
		{ID: 4, PipelineID: 2, CreatedBy: 1, RepoID: 2, Number: 2},
		{ID: 5, PipelineID: 3, CreatedBy: 1, RepoID: 1, Number: 1},
		{ID: 6, PipelineID: 3, CreatedBy: 1, RepoID: 1, Number: 2},
		{ID: 7, PipelineID: 1, CreatedBy: 1, RepoID: 1, Number: 3},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *ExecutionSuite) TestCreate() {
	addItems := []types.Execution{
		{ID: 1, PipelineID: 1, CreatedBy: 1, RepoID: 1},
		{ID: 2, PipelineID: 1, CreatedBy: 1, RepoID: 1},
		{ID: 3, PipelineID: 2, CreatedBy: 1, RepoID: 2},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.ErrorIs(suite.T(), err, store2.ErrDuplicate, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *ExecutionSuite) TestFind() {
	for id, pk := range []int64{1, 2, 3, 4, 5} {
		obj, err := suite.ormStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *ExecutionSuite) TestDelete() {
	err := suite.ormStore.Delete(suite.Ctx, 1, 1)
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.Delete(suite.Ctx, 2, 1)
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.Find(suite.Ctx, 3)
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.Find(suite.Ctx, 1)
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.Delete(suite.Ctx, 1, 1)
	require.NoError(suite.T(), err)

	err = suite.ormStore.Delete(suite.Ctx, 2, 1)
	require.NoError(suite.T(), err)
}

func (suite *ExecutionSuite) TestFindByNumber() {
	tests := []struct {
		ID         int64
		PipelineID int64
		Number     int64
	}{
		{ID: 1, PipelineID: 1, Number: 1},
		{ID: 2, PipelineID: 1, Number: 2},
		{ID: 3, PipelineID: 2, Number: 1},
		{ID: 4, PipelineID: 2, Number: 2},
		{ID: 5, PipelineID: 3, Number: 1},
		{ID: 6, PipelineID: 3, Number: 2},
		{ID: 7, PipelineID: 1, Number: 3},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.FindByNumber(suite.Ctx, test.PipelineID, test.Number)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.ID, obj.ID, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.FindByNumber(suite.Ctx, test.PipelineID, test.Number)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *ExecutionSuite) TestListAndCount() {
	tests := []struct {
		pipelineId int64
		wantLength int64
	}{
		{1, 3},
		{2, 2},
		{3, 2},
	}

	nilPG := types.Pagination{}

	for id, test := range tests {
		rules, err := suite.ormStore.List(suite.Ctx, test.pipelineId, nilPG, 0)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(rules), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.Count(suite.Ctx, test.pipelineId, 0)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		rulesB, err := suite.sqlxStore.List(suite.Ctx, test.pipelineId, nilPG, 0)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), rules, rulesB, testsuite.InvalidLoopMsgF, id)
	}
}

func addExecution(suite *testsuite.BaseSuite, pipelineID, executionID, repoID, number int64) {
	suite.T().Helper()

	s := pipeline.NewExecutionOrmStore(suite.Gdb)
	e := types.Execution{
		ID: executionID, PipelineID: pipelineID, CreatedBy: 1, RepoID: repoID, Number: number,
	}

	err := s.Create(suite.Ctx, &e)
	require.NoError(suite.T(), err)
}
