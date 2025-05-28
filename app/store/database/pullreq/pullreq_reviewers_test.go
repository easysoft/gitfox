// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package pullreq_test

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/pullreq"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTablePullReviewers = "pullreq_reviewers"
)

type PullReqReviewersSuite struct {
	testsuite.BaseSuite

	ormStore  *pullreq.ReviewerOrmStore
	sqlxStore *database.PullReqReviewerStore
}

func TestPullReqReviewersSuite(t *testing.T) {
	ctx := context.Background()

	st := &PullReqReviewersSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "pullreqs_reviewers",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pullreq.NewPullReqReviewerOrmStore(st.Gdb, ts.PrincipalCache)
		st.sqlxStore = database.NewPullReqReviewerStore(st.Sdb, ts.PrincipalCache)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 3, false)

		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)

		pqStore := pullreq.NewPullReqOrmStore(st.Gdb, ts.PrincipalCache)
		addPullReq(st.Ctx, t, pqStore, 1, "test")
		addPullReq(st.Ctx, t, pqStore, 2, "feat1")
	}

	suite.Run(t, st)
}

func (suite *PullReqReviewersSuite) SetupTest() {
	suite.addData()
}

func (suite *PullReqReviewersSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTablePullReviewers).Where("1 = 1").Delete(nil)
}

func (suite *PullReqReviewersSuite) addData() {
	addItems := []types.PullReqReviewer{
		{PullReqID: 1, PrincipalID: 2, RepoID: 1, CreatedBy: 1, Type: enum.PullReqReviewerTypeRequested, ReviewDecision: enum.PullReqReviewDecisionPending},
		{PullReqID: 2, PrincipalID: 3, RepoID: 1, CreatedBy: 1, Type: enum.PullReqReviewerTypeRequested, ReviewDecision: enum.PullReqReviewDecisionPending},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqReviewersSuite) TestFind() {
	tests := []struct {
		prId        int64
		principalId int64
	}{
		{1, 2},
		{2, 3},
	}
	for id, test := range tests {
		obj, err := suite.ormStore.Find(suite.Ctx, test.prId, test.principalId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.prId, test.principalId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqReviewersSuite) TestDelete() {
	err := suite.ormStore.Delete(suite.Ctx, 1, 2)
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.Delete(suite.Ctx, 2, 3)
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.Find(suite.Ctx, 2, 3)
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.Find(suite.Ctx, 1, 2)
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.Delete(suite.Ctx, 1, 2)
	require.NoError(suite.T(), err)

	err = suite.ormStore.Delete(suite.Ctx, 2, 3)
	require.NoError(suite.T(), err)
}

func (suite *PullReqReviewersSuite) TestList() {
	tests := []struct {
		prId       int64
		wantLength int64
	}{
		{1, 1}, {2, 1},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, test.prId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.ormStore.List(suite.Ctx, test.prId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}
