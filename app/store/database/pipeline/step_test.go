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
	testTableStep = "steps"
)

type StepSuite struct {
	testsuite.BaseSuite

	ormStore  *pipeline.StepStore
	sqlxStore store.StepStore
}

func TestStepSuite(t *testing.T) {
	ctx := context.Background()

	st := &StepSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "steps",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pipeline.NewStepOrmStore(st.Gdb)
		st.sqlxStore = database.NewStepStore(st.Sdb)

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

		addStage(&st.BaseSuite, 1, 1, 1, 1)
		addStage(&st.BaseSuite, 1, 2, 1, 2)
		addStage(&st.BaseSuite, 1, 3, 1, 3)
	}

	suite.Run(t, st)
}

func (suite *StepSuite) SetupTest() {
	suite.addData()
}

func (suite *StepSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableStep).Where("1 = 1").Delete(nil)
}

func (suite *StepSuite) addData() {
	addItems := []types.Step{
		{ID: 1, StageID: 1, Number: 1, Status: enum.CIStatusRunning, DependsOn: []string{}},
		{ID: 2, StageID: 1, Number: 2, Status: enum.CIStatusPending, DependsOn: []string{"test"}},
		{ID: 3, StageID: 2, Number: 1, Status: enum.CIStatusRunning, DependsOn: []string{}},
		{ID: 4, StageID: 2, Number: 2, Status: enum.CIStatusPending, DependsOn: []string{"test", "lint"}},
		{ID: 5, StageID: 2, Number: 3, Status: enum.CIStatusPending, DependsOn: []string{"world"}},
		{ID: 6, StageID: 3, Number: 1, Status: enum.CIStatusError, DependsOn: []string{"hello"}},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *StepSuite) TestCreate() {
	addItems := []types.Step{
		{ID: 1, StageID: 1, Number: 1, Status: enum.CIStatusRunning, DependsOn: []string{}},
		{ID: 2, StageID: 1, Number: 2, Status: enum.CIStatusPending, DependsOn: []string{"test"}},
		{ID: 3, StageID: 2, Number: 1, Status: enum.CIStatusRunning, DependsOn: []string{}},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.ErrorIs(suite.T(), err, store2.ErrDuplicate, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *StepSuite) TestFindByNumber() {
	tests := []struct {
		ID      int64
		StageID int64
		Number  int
	}{
		{1, 1, 1},
		{2, 1, 2},
		{3, 2, 1},
		{4, 2, 2},
		{5, 2, 3},
		{6, 3, 1},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.FindByNumber(suite.Ctx, test.StageID, test.Number)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.ID, obj.ID, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.FindByNumber(suite.Ctx, test.StageID, test.Number)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}
func addStep(suite *testsuite.BaseSuite, stepID, stageID, number int64) {
	suite.T().Helper()

	s := pipeline.NewStepOrmStore(suite.Gdb)
	e := types.Step{
		ID: stepID, StageID: stageID, Number: number, Status: enum.CIStatusRunning, DependsOn: []string{},
	}

	err := s.Create(suite.Ctx, &e)
	require.NoError(suite.T(), err)
}
