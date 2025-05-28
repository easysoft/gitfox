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

type CodeCommentSuite struct {
	testsuite.BaseSuite

	ormStore  *pullreq.CodeCommentView
	sqlxStore *database.CodeCommentView
}

func TestCodeCommentSuite(t *testing.T) {
	ctx := context.Background()

	st := &CodeCommentSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "code_comments",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = pullreq.NewCodeCommentOrmView(st.Gdb)
		st.sqlxStore = database.NewCodeCommentView(st.Sdb)

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

		pqActStore := pullreq.NewPullReqActivityOrmStore(st.Gdb, ts.PrincipalCache)
		addItems := []types.PullReqActivity{
			{ID: 1, PullReqID: 1, Order: 1, Type: enum.PullReqActivityTypeCodeComment, Kind: enum.PullReqActivityKindChangeComment,
				CodeComment: &types.CodeCommentFields{
					Outdated: false, MergeBaseSHA: "m1", SourceSHA: "s1", LineNew: 10, LineOld: 9, SpanNew: 2, SpanOld: 3,
				}, PayloadRaw: []byte(`{"old":"1", "new":"2"}`),
			},
			// 2 Kind unmatched
			{ID: 2, PullReqID: 1, Order: 2, Type: enum.PullReqActivityTypeCodeComment, Kind: enum.PullReqActivityKindSystem,
				CodeComment: &types.CodeCommentFields{
					Outdated: false, MergeBaseSHA: "m1", SourceSHA: "s1", LineNew: 10, LineOld: 9, SpanNew: 2, SpanOld: 3,
				}, PayloadRaw: []byte(`{"old":"1", "new":"2"}`),
			},
			// 3 Type unmatched
			{ID: 3, PullReqID: 1, Order: 3, Type: enum.PullReqActivityTypeMerge, Kind: enum.PullReqActivityKindChangeComment,
				CodeComment: &types.CodeCommentFields{
					Outdated: false, MergeBaseSHA: "m1", SourceSHA: "s1", LineNew: 10, LineOld: 9, SpanNew: 2, SpanOld: 3,
				}, PayloadRaw: []byte(`{"old":"1", "new":"2"}`),
			},
			// 4 outdated
			{ID: 4, PullReqID: 1, Order: 4, Type: enum.PullReqActivityTypeMerge, Kind: enum.PullReqActivityKindSystem,
				CodeComment: &types.CodeCommentFields{
					Outdated: true, MergeBaseSHA: "m1", SourceSHA: "s1", LineNew: 10, LineOld: 9, SpanNew: 2, SpanOld: 3,
				}, PayloadRaw: []byte(`{"old":"1", "new":"2"}`),
			},
			{ID: 5, PullReqID: 1, Order: 5, Type: enum.PullReqActivityTypeCodeComment, Kind: enum.PullReqActivityKindChangeComment,
				CodeComment: &types.CodeCommentFields{
					Outdated: false, MergeBaseSHA: "m2", SourceSHA: "s1", LineNew: 10, LineOld: 9, SpanNew: 2, SpanOld: 3,
				}, PayloadRaw: []byte(`{"old":"1", "new":"2"}`),
			},
			{ID: 6, PullReqID: 1, Order: 6, Type: enum.PullReqActivityTypeCodeComment, Kind: enum.PullReqActivityKindChangeComment,
				CodeComment: &types.CodeCommentFields{
					Outdated: false, MergeBaseSHA: "m1", SourceSHA: "s2", LineNew: 11, LineOld: 9, SpanNew: 2, SpanOld: 3,
				}, PayloadRaw: []byte(`{"old":"1", "new":"2"}`),
			},
			{ID: 7, PullReqID: 1, Order: 7, Type: enum.PullReqActivityTypeCodeComment, Kind: enum.PullReqActivityKindChangeComment,
				CodeComment: &types.CodeCommentFields{
					Outdated: false, MergeBaseSHA: "m3", SourceSHA: "s2", LineNew: 12, LineOld: 9, SpanNew: 2, SpanOld: 3,
				}, PayloadRaw: []byte(`{"old":"1", "new":"2"}`),
			},
			{ID: 8, PullReqID: 1, Order: 8, Type: enum.PullReqActivityTypeCodeComment, Kind: enum.PullReqActivityKindChangeComment,
				CodeComment: &types.CodeCommentFields{
					Outdated: false, MergeBaseSHA: "m4", SourceSHA: "s2", LineNew: 12, LineOld: 9, SpanNew: 2, SpanOld: 3,
				}, PayloadRaw: []byte(`{"old":"1", "new":"2"}`),
			},
		}

		for id, item := range addItems {
			item.CreatedBy = 1
			item.Updated = time.Now().UnixMilli()
			item.Created = time.Now().UnixMilli()
			item.RepoID = 1
			err := pqActStore.Create(st.Ctx, &item)
			require.NoError(t, err, testsuite.InvalidLoopMsgF, id)
		}
	}

	suite.Run(t, st)
}

func (suite *CodeCommentSuite) TestListNotAtSourceSHA() {
	tests := []struct {
		sha        string
		prId       int64
		wantLength int
	}{
		{sha: "s1", prId: 1, wantLength: 3},
		{sha: "s2", prId: 1, wantLength: 2},
		{sha: "s3", prId: 1, wantLength: 5},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.ListNotAtSourceSHA(suite.Ctx, test.prId, test.sha)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.ListNotAtSourceSHA(suite.Ctx, test.prId, test.sha)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB)
	}
}

func (suite *CodeCommentSuite) TestListNotAtMergeBaseSHA() {
	tests := []struct {
		sha        string
		prId       int64
		wantLength int
	}{
		{sha: "m1", prId: 1, wantLength: 3},
		{sha: "m2", prId: 1, wantLength: 4},
		{sha: "m3", prId: 1, wantLength: 4},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.ListNotAtMergeBaseSHA(suite.Ctx, test.prId, test.sha)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.ListNotAtMergeBaseSHA(suite.Ctx, test.prId, test.sha)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB)
	}
}
