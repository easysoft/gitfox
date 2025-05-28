// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pullreq_test

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/pullreq"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	store2 "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTablePullReqs = "pullreqs"
)

type PullReqSuite struct {
	testsuite.BaseSuite

	ormStore  *pullreq.OrmStore
	sqlxStore *database.PullReqStore
}

func TestPullReqSuite(t *testing.T) {
	ctx := context.Background()

	st := &PullReqSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "pullreqs",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pullreq.NewPullReqOrmStore(st.Gdb, ts.PrincipalCache)
		st.sqlxStore = database.NewPullReqStore(st.Sdb, ts.PrincipalCache)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 2, 1, 10)
	}

	suite.Run(t, st)
}

func (suite *PullReqSuite) SetupTest() {
	suite.addData()
}

func (suite *PullReqSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTablePullReqs).Where("1 = 1").Delete(nil)
}

var _addPullReqItems = []types.PullReq{
	{ID: 1, CreatedBy: 2, State: enum.PullReqStateOpen, SourceRepoID: 1, SourceBranch: "feat1", TargetRepoID: 1, Number: 1, TargetBranch: "main", SourceSHA: "sha1_1", MergeCheckStatus: enum.MergeCheckStatusUnchecked},
	{ID: 2, CreatedBy: 2, State: enum.PullReqStateOpen, SourceRepoID: 2, SourceBranch: "main", TargetRepoID: 1, Number: 2, TargetBranch: "main", SourceSHA: "sha1_2", MergeCheckStatus: enum.MergeCheckStatusUnchecked},
	{ID: 3, CreatedBy: 2, State: enum.PullReqStateMerged, SourceRepoID: 2, SourceBranch: "test", TargetRepoID: 1, Number: 3, TargetBranch: "test", SourceSHA: "sha1_3", MergeCheckStatus: enum.MergeCheckStatusUnchecked},
	{ID: 4, CreatedBy: 2, State: enum.PullReqStateClosed, SourceRepoID: 2, SourceBranch: "test", TargetRepoID: 1, Number: 4, TargetBranch: "test", SourceSHA: "sha1_4", MergeCheckStatus: enum.MergeCheckStatusUnchecked},
	{ID: 5, CreatedBy: 1, State: enum.PullReqStateOpen, SourceRepoID: 2, SourceBranch: "test", TargetRepoID: 2, Number: 1, TargetBranch: "test", SourceSHA: "sha1_5", MergeCheckStatus: enum.MergeCheckStatusUnchecked},
}

func (suite *PullReqSuite) addData() {
	for id, item := range _addPullReqItems {
		item.Version = 1
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqSuite) TestCreate() {
	for id, item := range _addPullReqItems {
		item.Version = 1
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.ErrorIs(suite.T(), err, store2.ErrDuplicate, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqSuite) TestFind() {
	for id, pk := range []int64{1, 2, 3, 4} {
		obj, err := suite.ormStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqSuite) TestFindByNumber() {
	tests := []struct {
		repoId int64
		number int64
		sha    string
	}{
		{1, 1, "sha1_1"},
		{1, 2, "sha1_2"},
		{1, 3, "sha1_3"},
		{1, 4, "sha1_4"},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.FindByNumber(suite.Ctx, test.repoId, test.number)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.sha, obj.SourceSHA, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.FindByNumber(suite.Ctx, test.repoId, test.number)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqSuite) TestResetMergeCheckStatus() {
	err := suite.ormStore.ResetMergeCheckStatus(suite.Ctx, 1, "main")
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.ResetMergeCheckStatus(suite.Ctx, 1, "test")
	require.NoError(suite.T(), err)

	tests := []struct {
		ID      int64
		status  enum.MergeCheckStatus
		version int64
	}{
		{ID: 1, status: enum.MergeCheckStatusUnchecked, version: 2},
		{ID: 2, status: enum.MergeCheckStatusUnchecked, version: 2},
		{ID: 3, status: enum.MergeCheckStatusUnchecked, version: 1}, // pk 3,4 closed or merged req can be updated
		{ID: 4, status: enum.MergeCheckStatusUnchecked, version: 1},
		{ID: 5, status: enum.MergeCheckStatusUnchecked, version: 1}, // pk 5 won't be changed for targetRepo
	}

	for id, test := range tests {
		obj, e := suite.ormStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), e, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.status, obj.MergeCheckStatus, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.version, obj.Version, testsuite.InvalidLoopMsgF, id)

		objB, e := suite.sqlxStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), e, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB)
	}
}

func (suite *PullReqSuite) TestDelete() {
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

func (suite *PullReqSuite) TestListAndCount() {
	_ = _addPullReqItems
	tests := []struct {
		filter     types.PullReqFilter
		wantLength int64
	}{
		{types.PullReqFilter{}, 5},
		{types.PullReqFilter{SourceRepoID: 1}, 1},
		{types.PullReqFilter{SourceRepoID: 2}, 4},
		{types.PullReqFilter{SourceRepoID: 2, SourceBranch: "test"}, 3},
		{types.PullReqFilter{SourceRepoID: 2, TargetRepoID: 2}, 1},
		{types.PullReqFilter{SourceRepoID: 2, TargetRepoID: 1}, 3},
		{types.PullReqFilter{SourceRepoID: 2, TargetRepoID: 1, TargetBranch: "test"}, 2},
		{types.PullReqFilter{SourceRepoID: 2, CreatedBy: []int64{1}}, 1},
		{types.PullReqFilter{SourceRepoID: 2, States: []enum.PullReqState{enum.PullReqStateClosed}}, 1},
		{types.PullReqFilter{SourceRepoID: 2, States: []enum.PullReqState{enum.PullReqStateClosed, enum.PullReqStateMerged}}, 2},
		{types.PullReqFilter{SourceRepoID: 2, TargetRepoID: 1, States: []enum.PullReqState{enum.PullReqStateOpen}}, 1},
	}

	for id, test := range tests {
		rules, err := suite.ormStore.List(suite.Ctx, &test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(rules), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.Count(suite.Ctx, &test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		rulesB, err := suite.sqlxStore.List(suite.Ctx, &test.filter)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), rules, rulesB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqSuite) TestRecreatePr() {
	// pr1 is opened
	t0 := _addPullReqItems[0]
	recreatePr := types.PullReq{CreatedBy: 2, State: enum.PullReqStateOpen,
		SourceRepoID: t0.SourceRepoID, SourceBranch: t0.SourceBranch, TargetRepoID: t0.TargetRepoID, TargetBranch: t0.TargetBranch,
		Number: t0.Number + 1000, SourceSHA: t0.SourceSHA, MergeCheckStatus: enum.MergeCheckStatusUnchecked}

	testPr1 := recreatePr
	testPr1.Number = recreatePr.Number + 1
	// create testPr1 failed for conflict source/target repo/branch with status open
	err := suite.ormStore.Create(suite.Ctx, &testPr1)
	require.ErrorIs(suite.T(), err, store2.ErrDuplicate)

	// set t0 closed
	t0.State = enum.PullReqStateClosed
	err = suite.ormStore.Update(suite.Ctx, &t0)
	require.NoError(suite.T(), err)

	// create testPr1 again
	err = suite.ormStore.Create(suite.Ctx, &testPr1)
	require.NoError(suite.T(), err)

	// close testPr1
	testPr1.State = enum.PullReqStateClosed
	err = suite.ormStore.Update(suite.Ctx, &testPr1)
	require.NoError(suite.T(), err)

	// create testPr2
	testPr2 := recreatePr
	testPr2.Number = testPr1.Number + 1
	err = suite.ormStore.Create(suite.Ctx, &testPr2)
	require.NoError(suite.T(), err)
}

func addPullReq(ctx context.Context, t *testing.T, st store.PullReqStore, id int64, branch string) {
	obj := types.PullReq{ID: id, CreatedBy: 1, State: enum.PullReqStateOpen, SourceRepoID: 1, SourceBranch: branch, TargetRepoID: 1, Number: id, TargetBranch: "main", SourceSHA: "sha1_1", MergeCheckStatus: enum.MergeCheckStatusUnchecked}
	err := st.Create(ctx, &obj)
	require.NoError(t, err)
}
