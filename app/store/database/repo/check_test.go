// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/cache"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/app/store/database/repo"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const testTableCheck = "checks"

type testListCase struct {
	repoId         int64
	commitSHA      string
	listOpts       types.CheckListOptions
	wantListLength int
}

type testCountCase struct {
	repoId    int64
	commitSHA string
	wantCount int
}

type testListRecentCase struct {
	repoId     int64
	since      int64
	recentOpts types.CheckRecentOptions
	wantCount  int
}

type testFindByIdentifierCase struct {
	repoId     int64
	commitSHA  string
	identifier string
	expectErr  bool
}

type CheckSuite struct {
	testsuite.BaseSuite

	ormStore  *repo.CheckStoreOrm
	sqlxStore *database.CheckStore
}

func (suite *CheckSuite) SetupTest() {
	suite.addData()
}

func (suite *CheckSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableCheck).Where("1 = 1").Delete(nil)
}

func (suite *CheckSuite) addData() {
	addItems := []struct {
		creatorId int64
		repoId    int64
		commitSHA string
	}{
		{1, 1, "sha1:1"},
		{1, 1, "sha1:2"},
		{1, 1, "sha1:3"},
		{2, 1, "sha1:3"},
		{3, 1, "sha1:3"},
		{3, 1, "sha1:4"},
		{2, 2, "sha1:1"},
		{3, 2, "sha1:2"},
		{1, 2, "sha1:3"},
	}

	for _, item := range addItems {
		upsertCheck(suite.Ctx, suite.T(), suite.ormStore, item.creatorId, item.repoId, item.commitSHA)
	}
}

func (suite *CheckSuite) testListCases() []testListCase {
	return []testListCase{
		{repoId: 1, commitSHA: "sha1:1", listOpts: types.CheckListOptions{}, wantListLength: 1},
		{repoId: 1, commitSHA: "sha1:2", listOpts: types.CheckListOptions{}, wantListLength: 1},
		{repoId: 1, commitSHA: "sha1:3", listOpts: types.CheckListOptions{}, wantListLength: 3},
		{repoId: 1, commitSHA: "sha1:3", listOpts: types.CheckListOptions{ListQueryFilter: types.ListQueryFilter{
			Pagination: types.Pagination{Page: 1, Size: 10},
			Query:      "",
		}}, wantListLength: 3},
		{repoId: 1, commitSHA: "sha1:3", listOpts: types.CheckListOptions{ListQueryFilter: types.ListQueryFilter{
			Pagination: types.Pagination{Page: 1, Size: 2},
			Query:      "",
		}}, wantListLength: 2},
		{repoId: 1, commitSHA: "sha1:3", listOpts: types.CheckListOptions{ListQueryFilter: types.ListQueryFilter{
			Pagination: types.Pagination{Page: 2, Size: 2},
			Query:      "",
		}}, wantListLength: 1},
		{repoId: 1, commitSHA: "sha1:3", listOpts: types.CheckListOptions{ListQueryFilter: types.ListQueryFilter{
			Pagination: types.Pagination{Page: 3, Size: 2},
			Query:      "",
		}}, wantListLength: 0},
	}
}

func (suite *CheckSuite) testCountCases() []testCountCase {
	return []testCountCase{
		{repoId: 1, commitSHA: "sha1:1", wantCount: 1},
		{repoId: 1, commitSHA: "sha1:2", wantCount: 1},
		{repoId: 1, commitSHA: "sha1:3", wantCount: 3},
		{repoId: 1, commitSHA: "sha1:4", wantCount: 1},
		{repoId: 2, commitSHA: "sha1:1", wantCount: 1},
		{repoId: 2, commitSHA: "sha1:2", wantCount: 1},
		{repoId: 2, commitSHA: "sha1:3", wantCount: 1},
	}
}

func (suite *CheckSuite) testListRecentCases() []testListRecentCase {
	now := time.Now().UnixMilli()
	appendItems := []struct {
		creatorId int64
		repoId    int64
		commitSHA string
	}{
		{1, 1, "sha1:5"},
		{1, 1, "sha1:6"},
		{2, 1, "sha1:7"},
		{2, 2, "sha1:4"},
		{3, 2, "sha1:5"},
		{1, 2, "sha1:6"},
	}
	for _, item := range appendItems {
		upsertCheck(suite.Ctx, suite.T(), suite.ormStore, item.creatorId, item.repoId, item.commitSHA)
	}

	return []testListRecentCase{
		{repoId: 1, recentOpts: types.CheckRecentOptions{
			Query: "",
			Since: now - 600*1000,
		}, wantCount: 3},
		{repoId: 2, recentOpts: types.CheckRecentOptions{
			Query: "",
			Since: now - 600*1000,
		}, wantCount: 3},
		{repoId: 1, recentOpts: types.CheckRecentOptions{
			Query: "",
			Since: now,
		}, wantCount: 2},
		{repoId: 2, recentOpts: types.CheckRecentOptions{
			Query: "",
			Since: now,
		}, wantCount: 3},
	}
}

func (suite *CheckSuite) testFindByIdentifierCases() []testFindByIdentifierCase {
	return []testFindByIdentifierCase{
		{repoId: 1, commitSHA: "sha1:1", identifier: "check-1", expectErr: false},
		{repoId: 1, commitSHA: "sha1:2", identifier: "check-1", expectErr: false},
		{repoId: 1, commitSHA: "sha1:10", identifier: "check-1", expectErr: true},
		{repoId: 1, commitSHA: "sha1:1", identifier: "check-11", expectErr: true},
		{repoId: 11, commitSHA: "sha1:1", identifier: "check-1", expectErr: true},
	}
}

func (suite *CheckSuite) TestFindByIdentifier() {
	t := suite.T()

	for id, test := range suite.testFindByIdentifierCases() {
		obj, err := suite.ormStore.FindByIdentifier(suite.Ctx, test.repoId, test.commitSHA, test.identifier)
		if test.expectErr {
			require.Error(t, err)
			continue
		}

		require.Equal(t, test.commitSHA, obj.CommitSHA, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *CheckSuite) TestFindByIdentifierNoDiff() {
	t := suite.T()

	for id, test := range suite.testFindByIdentifierCases() {
		obj1, err1 := suite.ormStore.FindByIdentifier(suite.Ctx, test.repoId, test.commitSHA, test.identifier)

		obj2, err2 := suite.sqlxStore.FindByIdentifier(suite.Ctx, test.repoId, test.commitSHA, test.identifier)
		if test.expectErr {
			require.Error(t, err1)
			require.Error(t, err2)
			continue
		}

		require.EqualExportedValues(t, obj1, obj2, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *CheckSuite) TestUpsert() {
	suite.addData()
}

func (suite *CheckSuite) TestList() {
	t := suite.T()

	for id, test := range suite.testListCases() {
		items, err := suite.ormStore.List(suite.Ctx, test.repoId, test.commitSHA, test.listOpts)
		require.NoError(t, err)
		require.Equal(t, test.wantListLength, len(items), testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *CheckSuite) TestListNoDiff() {
	t := suite.T()

	for id, test := range suite.testListCases() {
		items1, err := suite.ormStore.List(suite.Ctx, test.repoId, test.commitSHA, test.listOpts)
		require.NoError(t, err)

		items2, err := suite.sqlxStore.List(suite.Ctx, test.repoId, test.commitSHA, test.listOpts)
		require.NoError(t, err)

		if test.wantListLength == 0 {
			require.Empty(t, items1)
			require.Empty(t, items2)
			continue
		}

		require.ElementsMatch(t, items1, items2, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *CheckSuite) TestCount() {
	t := suite.T()

	for id, test := range suite.testCountCases() {
		count, err := suite.ormStore.Count(suite.Ctx, test.repoId, test.commitSHA, types.CheckListOptions{})
		require.NoError(t, err)
		require.Equal(t, test.wantCount, count, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *CheckSuite) TestCountNoDiff() {
	t := suite.T()

	for id, test := range suite.testCountCases() {
		count1, err := suite.ormStore.Count(suite.Ctx, test.repoId, test.commitSHA, types.CheckListOptions{})
		require.NoError(t, err, testsuite.InvalidLoopMsgF, id)

		count2, err := suite.sqlxStore.Count(suite.Ctx, test.repoId, test.commitSHA, types.CheckListOptions{})
		require.NoError(t, err, testsuite.InvalidLoopMsgF, id)

		require.Equal(t, count1, count2, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *CheckSuite) TestListRecent() {
	t := suite.T()

	for id, test := range suite.testListRecentCases() {
		items, err := suite.ormStore.ListRecent(suite.Ctx, test.repoId, test.recentOpts)
		require.NoError(t, err)
		require.Equal(t, test.wantCount, len(items), testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *CheckSuite) TestListRecentNoDiff() {
	t := suite.T()

	for id, test := range suite.testListRecentCases() {
		items1, err := suite.ormStore.ListRecent(suite.Ctx, test.repoId, test.recentOpts)
		require.NoError(t, err, testsuite.InvalidLoopMsgF, id)

		items2, err := suite.sqlxStore.ListRecent(suite.Ctx, test.repoId, test.recentOpts)
		require.NoError(t, err, testsuite.InvalidLoopMsgF, id)

		require.ElementsMatch(t, items1, items2, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *CheckSuite) TestListResults() {
	t := suite.T()

	for id, test := range suite.testCountCases() {
		items, err := suite.ormStore.ListResults(suite.Ctx, test.repoId, test.commitSHA)
		require.NoError(t, err)
		require.Equal(t, test.wantCount, len(items), testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *CheckSuite) TestListResultsNoDiff() {
	t := suite.T()

	for id, test := range suite.testCountCases() {
		items1, err := suite.ormStore.ListResults(suite.Ctx, test.repoId, test.commitSHA)
		require.NoError(t, err, testsuite.InvalidLoopMsgF, id)

		items2, err := suite.sqlxStore.ListResults(suite.Ctx, test.repoId, test.commitSHA)
		require.NoError(t, err, testsuite.InvalidLoopMsgF, id)

		require.ElementsMatch(t, items1, items2, testsuite.InvalidLoopMsgF, id)
	}
}

func upsertCheck(
	ctx context.Context,
	t *testing.T,
	checkStore store.CheckStore,
	createBy int64,
	repoId int64,
	commitSha string,
) {
	t.Helper()

	identifier := fmt.Sprintf("check-%d", createBy)
	check := types.Check{CreatedBy: createBy, Identifier: identifier, RepoID: repoId, CommitSHA: commitSha, Payload: types.CheckPayload{
		Version: "1", Kind: enum.CheckPayloadKindEmpty,
		Data: []byte("{}")}, Metadata: []byte("{}"), Created: time.Now().UnixMilli(),
	}
	if err := checkStore.Upsert(ctx, &check); err != nil {
		t.Fatalf("failed to create check %v", err)
	}
}

func TestCheckSuite(t *testing.T) {
	ctx := context.Background()

	st := &CheckSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "check",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		testsuite.AddUser(ctx, t, ts.Principal, 1, true)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)

		testsuite.AddUser(ctx, t, ts.Principal, 2, true)
		testsuite.AddUser(ctx, t, ts.Principal, 3, true)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 2, 1, 10)

		ormPrincipalInfoView := principal.NewPrincipalOrmInfoView(st.Gdb)
		ormPrincipalInfoCache := cache.ProvidePrincipalInfoCache(ormPrincipalInfoView)

		sqlxPrincipalInfoView := database.NewPrincipalInfoView(st.Sdb)
		sqlxPrincipalInfoCache := cache.ProvidePrincipalInfoCache(sqlxPrincipalInfoView)

		st.ormStore = repo.NewCheckStoreOrm(st.Gdb, ormPrincipalInfoCache)
		st.sqlxStore = database.NewCheckStore(st.Sdb, sqlxPrincipalInfoCache)
	}

	suite.Run(t, st)
}
