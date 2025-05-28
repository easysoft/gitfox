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

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTablePullReviews = "pullreq_reviews"
)

type PullReqReviewsSuite struct {
	testsuite.BaseSuite

	ormStore  *pullreq.ReviewOrmStore
	sqlxStore *database.PullReqReviewStore
}

func TestPullReqReviewsSuite(t *testing.T) {
	ctx := context.Background()

	st := &PullReqReviewsSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "pullreqs_reviews",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pullreq.NewPullReqReviewOrmStore(st.Gdb)
		st.sqlxStore = database.NewPullReqReviewStore(st.Sdb)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 3, false)

		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)

		pqStore := pullreq.NewPullReqOrmStore(st.Gdb, ts.PrincipalCache)
		addPullReq(st.Ctx, t, pqStore, 1, "test")
		addPullReq(st.Ctx, t, pqStore, 2, "feat1")
		addPullReq(st.Ctx, t, pqStore, 3, "feat2")
	}

	suite.Run(t, st)
}

func (suite *PullReqReviewsSuite) SetupTest() {
	suite.addData()
}

func (suite *PullReqReviewsSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTablePullReviews).Where("1 = 1").Delete(nil)
}

func (suite *PullReqReviewsSuite) addData() {
	addItems := []types.PullReqReview{
		{ID: 1, PullReqID: 1, CreatedBy: 3},
		{ID: 2, PullReqID: 2, CreatedBy: 2},
		{ID: 3, PullReqID: 3, CreatedBy: 1},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqReviewsSuite) TestFind() {
	for id, pk := range []int64{1, 2, 3} {
		obj, err := suite.ormStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		objB, err := suite.sqlxStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}
