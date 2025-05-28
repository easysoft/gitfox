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
	testTablePullFileViews = "pullreq_file_views"
)

type PullReqFileViewsSuite struct {
	testsuite.BaseSuite

	ormStore  *pullreq.FileViewOrmStore
	sqlxStore *database.PullReqFileViewStore
}

func TestPullReqFileViewSuite(t *testing.T) {
	ctx := context.Background()

	st := &PullReqFileViewsSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "pullreqs_file_views",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pullreq.NewPullReqFileViewOrmStore(st.Gdb)
		st.sqlxStore = database.NewPullReqFileViewStore(st.Sdb)

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

func (suite *PullReqFileViewsSuite) SetupTest() {
	suite.addData()
}

func (suite *PullReqFileViewsSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTablePullFileViews).Where("1 = 1").Delete(nil)
}

var _addFileViews = []types.PullReqFileView{
	{PullReqID: 1, PrincipalID: 1, Path: "a/b/1", SHA: "sha1:1"},
	{PullReqID: 2, PrincipalID: 3, Path: "a/b/c/1", SHA: "sha1:2"},
	{PullReqID: 3, PrincipalID: 2, Path: "a/b/d/2", SHA: "sha1:3"},
	{PullReqID: 1, PrincipalID: 1, Path: "a/b/1", SHA: "sha1:4"},
}

func (suite *PullReqFileViewsSuite) addData() {
	for id, item := range _addFileViews {
		err := suite.ormStore.Upsert(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqFileViewsSuite) TestFindByFileForPrincipal() {
	tests := []types.PullReqFileView{
		{PullReqID: 1, PrincipalID: 1, Path: "a/b/1", SHA: "sha1:4"},
		{PullReqID: 2, PrincipalID: 3, Path: "a/b/c/1", SHA: "sha1:2"},
		{PullReqID: 3, PrincipalID: 2, Path: "a/b/d/2", SHA: "sha1:3"},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.FindByFileForPrincipal(suite.Ctx, test.PullReqID, test.PrincipalID, test.Path)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.SHA, obj.SHA, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqFileViewsSuite) TestDeleteByFileForPrincipal() {
	for id, test := range _addFileViews[0:3] {
		err := suite.ormStore.DeleteByFileForPrincipal(suite.Ctx, test.PullReqID, test.PrincipalID, test.Path)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *PullReqFileViewsSuite) TestMarkObsolete() {
	err := suite.ormStore.MarkObsolete(suite.Ctx, 1, []string{"a/b/1", "a/b/c/1"})
	require.NoError(suite.T(), err)

	obj, err := suite.ormStore.FindByFileForPrincipal(suite.Ctx, 1, 1, "a/b/1")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), true, obj.Obsolete)
}

func (suite *PullReqFileViewsSuite) TestList() {
	tests := []struct {
		prId        int64
		principalID int64
		wantLength  int64
	}{
		{1, 1, 1},
		{2, 3, 1},
		{3, 2, 1},
		{3, 3, 0},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, test.prId, test.principalID)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(objs))

		objsB, err := suite.sqlxStore.List(suite.Ctx, test.prId, test.principalID)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}
