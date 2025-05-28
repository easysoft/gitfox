// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pullreq_test

import (
	"context"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/pullreq"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTablePullActivity = "pullreq_activities"
)

type PullReqActivitySuite struct {
	testsuite.BaseSuite

	ormStore  *pullreq.ActivityOrmStore
	sqlxStore *database.PullReqActivityStore
}

func TestPullReqActivitySuite(t *testing.T) {
	ctx := context.Background()

	st := &PullReqActivitySuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "pullreqs_activity",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pullreq.NewPullReqActivityOrmStore(st.Gdb, ts.PrincipalCache)
		st.sqlxStore = database.NewPullReqActivityStore(st.Sdb, ts.PrincipalCache)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 3, false)

		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 2, 1, 10)

		pqStore := pullreq.NewPullReqOrmStore(st.Gdb, ts.PrincipalCache)
		addPullReq(st.Ctx, t, pqStore, 1, "test")
		addPullReq(st.Ctx, t, pqStore, 2, "feat1")
		addPullReq(st.Ctx, t, pqStore, 3, "feat2")
	}

	suite.Run(t, st)
}

func (suite *PullReqActivitySuite) SetupTest() {
	suite.addData()
}

func (suite *PullReqActivitySuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTablePullActivity).Where("1 = 1").Delete(nil)
}

func ptrInt64(i int64) *int64 {
	return &i
}

func (suite *PullReqActivitySuite) addData() {
	addItems := []types.PullReqActivity{
		{ID: 1, CreatedBy: 1, RepoID: 1, PullReqID: 1, Order: 1, SubOrder: 1, Type: enum.PullReqActivityTypeMerge, Kind: enum.PullReqActivityKindSystem,
			PayloadRaw: []byte(`{"old":"1", "new":"2"}`)},
		{ID: 2, CreatedBy: 2, RepoID: 1, PullReqID: 1, Order: 2, SubOrder: 1, Type: enum.PullReqActivityTypeComment, Kind: enum.PullReqActivityKindComment,
			PayloadRaw: []byte(`{"old":"1", "new":"2"}`)},
		{ID: 3, CreatedBy: 1, RepoID: 1, PullReqID: 1, Order: 3, SubOrder: 1, ParentID: ptrInt64(2), Type: enum.PullReqActivityTypeComment, Kind: enum.PullReqActivityKindComment,
			PayloadRaw: []byte(`{"old":"1", "new":"2"}`)},
		{ID: 4, CreatedBy: 1, RepoID: 1, PullReqID: 1, Order: 4, SubOrder: 1, ParentID: ptrInt64(2), Type: enum.PullReqActivityTypeCodeComment, Kind: enum.PullReqActivityKindChangeComment,
			PayloadRaw: []byte(`{"old":"1", "new":"2"}`)},
		{ID: 5, CreatedBy: 2, RepoID: 1, PullReqID: 1, Order: 5, SubOrder: 1, ParentID: ptrInt64(2), Type: enum.PullReqActivityTypeStateChange, Kind: enum.PullReqActivityKindComment,
			PayloadRaw: []byte(`{"old":"1", "new":"2"}`)},
		{ID: 6, CreatedBy: 3, RepoID: 1, PullReqID: 1, Order: 2, SubOrder: 2, Type: enum.PullReqActivityTypeTitleChange, Kind: enum.PullReqActivityKindComment,
			PayloadRaw: []byte(`{"old":"1", "new":"2"}`)},
		{ID: 7, CreatedBy: 3, RepoID: 1, PullReqID: 1, Order: 2, SubOrder: 3, Type: enum.PullReqActivityTypeStateChange, Kind: enum.PullReqActivityKindComment,
			PayloadRaw: []byte(`{"old":"1", "new":"2"}`)},
	}

	now := time.Now().UnixMilli()

	for id, item := range addItems {
		item.Created = now + int64(id)
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqActivitySuite) TestListAuthorIDs() {
	tests := []struct {
		prID   int64
		order  int64
		expect []int64
	}{
		{1, 2, []int64{2, 3}},
	}

	for id, test := range tests {
		authors, err := suite.ormStore.ListAuthorIDs(suite.Ctx, test.prID, test.order)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), test.expect, authors)

		authorsB, err := suite.sqlxStore.ListAuthorIDs(suite.Ctx, test.prID, test.order)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), authors, authorsB)
	}
}

func (suite *PullReqActivitySuite) TestFind() {
	for id, pk := range []int64{1, 2, 3, 4} {
		obj, err := suite.ormStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB)
	}
}

func (suite *PullReqActivitySuite) TestListAndCount() {
	pr4, err := suite.ormStore.Find(suite.Ctx, 4)
	require.NoError(suite.T(), err)

	tests := []struct {
		prId       int64
		filter     types.PullReqActivityFilter
		wantLength int64
	}{
		{1, types.PullReqActivityFilter{}, 7},
		{1, types.PullReqActivityFilter{Types: []enum.PullReqActivityType{
			enum.PullReqActivityTypeStateChange,
		}}, 2},
		{1, types.PullReqActivityFilter{Types: []enum.PullReqActivityType{
			enum.PullReqActivityTypeStateChange, enum.PullReqActivityTypeTitleChange,
		}}, 3},
		{1, types.PullReqActivityFilter{Kinds: []enum.PullReqActivityKind{
			enum.PullReqActivityKindComment,
		}}, 5},
		{1, types.PullReqActivityFilter{Kinds: []enum.PullReqActivityKind{
			enum.PullReqActivityKindSystem, enum.PullReqActivityKindChangeComment,
		}}, 2},
		{1, types.PullReqActivityFilter{Before: pr4.Created}, 3},
		{1, types.PullReqActivityFilter{After: pr4.Created}, 3},
	}

	for id, test := range tests {
		rules, err := suite.ormStore.List(suite.Ctx, test.prId, &test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(rules), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.Count(suite.Ctx, test.prId, &test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		rulesB, err := suite.sqlxStore.List(suite.Ctx, test.prId, &test.filter)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), rules, rulesB, testsuite.InvalidLoopMsgF, id)
	}
}
