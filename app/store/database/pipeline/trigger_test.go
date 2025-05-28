// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pipeline_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/pipeline"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableTrigger = "triggers"
)

type TriggerSuite struct {
	testsuite.BaseSuite

	ormStore  *pipeline.TriggerStore
	sqlxStore store.TriggerStore
}

func TestTriggerSuite(t *testing.T) {
	ctx := context.Background()

	st := &TriggerSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "triggers",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pipeline.NewTriggerOrmStore(st.Gdb)
		st.sqlxStore = database.NewTriggerStore(st.Sdb)

		// add init data
		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 2, 1, 10)

		ppStore := pipeline.NewPipelineOrmStore(st.Gdb)
		addPipeline(st.Ctx, t, ppStore, 1, 1, 1)
		addPipeline(st.Ctx, t, ppStore, 2, 2, 1)
		addPipeline(st.Ctx, t, ppStore, 3, 1, 1)
	}

	suite.Run(t, st)
}

func (suite *TriggerSuite) SetupTest() {
	suite.addData()
}

func (suite *TriggerSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableTrigger).Where("1 = 1").Delete(nil)
}

var testAddTriggerItems = []struct {
	id         int64
	pipelineID int64
	repoID     int64
	identifier string
	disabled   bool
	actions    []enum.TriggerAction
}{
	{id: 1, pipelineID: 1, repoID: 1, identifier: fmt.Sprintf("trigger_1")},
	{id: 2, pipelineID: 1, repoID: 1, identifier: fmt.Sprintf("trigger_2"),
		actions: []enum.TriggerAction{enum.TriggerActionBranchUpdated, enum.TriggerActionTagUpdated}},
	{id: 3, pipelineID: 3, repoID: 1, identifier: fmt.Sprintf("trigger_3"),
		actions: []enum.TriggerAction{enum.TriggerActionBranchUpdated, enum.TriggerActionTagUpdated}, disabled: true},
	{id: 4, pipelineID: 3, repoID: 1, identifier: fmt.Sprintf("trigger_4"),
		actions: []enum.TriggerAction{enum.TriggerActionPullReqBranchUpdated, enum.TriggerActionTagCreated}},
	{id: 5, pipelineID: 2, repoID: 2, identifier: fmt.Sprintf("trigger_5"),
		actions: []enum.TriggerAction{enum.TriggerActionPullReqClosed}},
	{id: 6, pipelineID: 2, repoID: 2, identifier: fmt.Sprintf("trigger_6"),
		actions: []enum.TriggerAction{enum.TriggerActionPullReqReopened}, disabled: true},
	{id: 7, pipelineID: 2, repoID: 2, identifier: fmt.Sprintf("trigger_7"),
		actions: []enum.TriggerAction{enum.TriggerActionTagCreated}},
}

func (suite *TriggerSuite) addData() {
	now := time.Now().UnixMilli()
	for id, item := range testAddTriggerItems {
		obj := types.Trigger{
			ID:         item.id,
			PipelineID: item.pipelineID,
			RepoID:     item.repoID,
			Identifier: item.identifier,
			Disabled:   item.disabled,
			Actions:    item.actions,
			CreatedBy:  1,
			Created:    now,
			Updated:    now,
		}

		err := suite.ormStore.Create(suite.Ctx, &obj)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TriggerSuite) TestFindByIdentifier() {
	for id, test := range testAddTriggerItems {
		obj, err := suite.ormStore.FindByIdentifier(suite.Ctx, test.pipelineID, test.identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.id, obj.ID)

		objB, err := suite.sqlxStore.FindByIdentifier(suite.Ctx, test.pipelineID, test.identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TriggerSuite) TestDeleteByIdentifier() {
	err := suite.ormStore.DeleteByIdentifier(suite.Ctx, 1, "trigger_1")
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, 3, "trigger_4")
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.FindByIdentifier(suite.Ctx, 3, "trigger_4")
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.FindByIdentifier(suite.Ctx, 1, "trigger_1")
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, 1, "trigger_1")
	require.NoError(suite.T(), err)

	err = suite.ormStore.DeleteByIdentifier(suite.Ctx, 3, "trigger_4")
	require.NoError(suite.T(), err)
}

func (suite *TriggerSuite) TestListAllEnabled() {
	tests := []struct {
		repoId     int64
		wantLength int64
	}{
		{1, 3},
		{2, 2},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.ListAllEnabled(suite.Ctx, test.repoId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.ListAllEnabled(suite.Ctx, test.repoId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TriggerSuite) TestListAncCount() {
	tests := []struct {
		pipelineId int64
		filter     types.ListQueryFilter
		wantLength int64
	}{
		{1, types.ListQueryFilter{}, 2},
		{2, types.ListQueryFilter{}, 3},
		{3, types.ListQueryFilter{}, 2},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, test.pipelineId, test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.Count(suite.Ctx, test.pipelineId, test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.List(suite.Ctx, test.pipelineId, test.filter)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}
