// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts_test

import (
	"context"
	"sort"
	"testing"

	"github.com/easysoft/gitfox/app/store/database/artifacts"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"

	"github.com/guregu/null"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableTreeNode = "artifact_tree_nodes"
)

type TreeNodeSuite struct {
	testsuite.BaseSuite

	ormStore *artifacts.Store
}

func TestArtifactTreeNodeSuite(t *testing.T) {
	ctx := context.Background()

	st := &TreeNodeSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "artifactTreeNode",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = artifacts.NewStore(st.Gdb)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
	}

	suite.Run(t, st)
}

func (suite *TreeNodeSuite) SetupTest() {
	suite.addData()
}

func (suite *TreeNodeSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableTreeNode).Where("1 = 1").Delete(nil)
}

func (suite *TreeNodeSuite) getTestData() []types.ArtifactTreeNode {
	return []types.ArtifactTreeNode{
		{ID: 1, OwnerID: 1, Name: "raw", Path: "/", Type: types.ArtifactTreeNodeTypeFormat, Format: types.ArtifactRawFormat},
		{ID: 2, OwnerID: 1, Name: "container", Path: "/", Type: types.ArtifactTreeNodeTypeFormat, Format: types.ArtifactContainerFormat},
		{ID: 3, ParentID: null.IntFrom(1), OwnerID: 1, Name: "runtime", Path: "/runtime", Type: types.ArtifactTreeNodeTypeDirectory, Format: types.ArtifactRawFormat},
		{ID: 4, ParentID: null.IntFrom(3), OwnerID: 1, Name: "web", Path: "/runtime/web", Type: types.ArtifactTreeNodeTypeDirectory, Format: types.ArtifactRawFormat},
		{ID: 5, ParentID: null.IntFrom(3), OwnerID: 1, Name: "db", Path: "/runtime/db", Type: types.ArtifactTreeNodeTypeDirectory, Format: types.ArtifactRawFormat},
		{ID: 6, ParentID: null.IntFrom(4), OwnerID: 1, Name: "apache", Path: "/runtime/web/apache", Type: types.ArtifactTreeNodeTypeDirectory, Format: types.ArtifactRawFormat},
		{ID: 7, ParentID: null.IntFrom(5), OwnerID: 1, Name: "mysql", Path: "/runtime/db/mysql", Type: types.ArtifactTreeNodeTypeDirectory, Format: types.ArtifactRawFormat},
		{ID: 8, ParentID: null.IntFrom(6), OwnerID: 1, Name: "2.8", Path: "/runtime/web/apache/2.8", Type: types.ArtifactTreeNodeTypeVersion, Format: types.ArtifactRawFormat},
		{ID: 9, ParentID: null.IntFrom(6), OwnerID: 1, Name: "2.9", Path: "/runtime/web/apache/2.9", Type: types.ArtifactTreeNodeTypeVersion, Format: types.ArtifactRawFormat},
		{ID: 10, ParentID: null.IntFrom(7), OwnerID: 1, Name: "8.0", Path: "/runtime/db/mysql/8.0", Type: types.ArtifactTreeNodeTypeVersion, Format: types.ArtifactRawFormat},
		{ID: 11, ParentID: null.IntFrom(2), OwnerID: 1, Name: "redis", Path: "/redis", Type: types.ArtifactTreeNodeTypeDirectory, Format: types.ArtifactContainerFormat},
		{ID: 12, ParentID: null.IntFrom(2), OwnerID: 1, Name: "influxdb", Path: "/influxdb", Type: types.ArtifactTreeNodeTypeDirectory, Format: types.ArtifactContainerFormat},
		{ID: 13, ParentID: null.IntFrom(11), OwnerID: 1, Name: "5.0", Path: "/redis/5.0", Type: types.ArtifactTreeNodeTypeVersion, Format: types.ArtifactContainerFormat},
		{ID: 14, ParentID: null.IntFrom(12), OwnerID: 1, Name: "4.3.5", Path: "/influxdb/4.3.5", Type: types.ArtifactTreeNodeTypeVersion, Format: types.ArtifactContainerFormat},
	}
}

func (suite *TreeNodeSuite) addData() {
	for id, item := range suite.getTestData() {
		err := suite.ormStore.Nodes().Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TreeNodeSuite) TestGetByPath() {
	s := suite.ormStore.Nodes()
	for id, input := range suite.getTestData() {
		obj, err := s.GetByPath(suite.Ctx, 1, input.Path, input.Format)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), input.Name, obj.Name, id)
	}
}

func (suite *TreeNodeSuite) TestListFormats() {
	s := suite.ormStore.Nodes()
	objs, err := s.ListFormats(suite.Ctx, 1)
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), 2, len(objs))

	require.Equal(suite.T(), objs[0].Name, "container")
	require.Equal(suite.T(), objs[1].Name, "raw")
}

func (suite *TreeNodeSuite) TestListByParentId() {
	tests := []struct {
		ParentID    int64
		ExpectedPks []int64
	}{
		{ParentID: 1, ExpectedPks: []int64{3}},
		{ParentID: 2, ExpectedPks: []int64{11, 12}},
		{ParentID: 3, ExpectedPks: []int64{4, 5}},
		{ParentID: 4, ExpectedPks: []int64{6}},
		{ParentID: 5, ExpectedPks: []int64{7}},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.Nodes().ListByParentId(suite.Ctx, test.ParentID)
		require.NoError(suite.T(), err, test.ParentID)

		ids := make([]int64, 0, len(objs))
		for _, obj := range objs {
			ids = append(ids, obj.ID)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

		require.Equal(suite.T(), test.ExpectedPks, ids, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TreeNodeSuite) TestRecurseCreate() {
	tests := []types.ArtifactTreeNode{
		{OwnerID: 1, Name: "8.0", Path: "/runtime/db/mysql/8.1", Type: types.ArtifactTreeNodeTypeVersion, Format: types.ArtifactRawFormat},
		{OwnerID: 1, Name: "8.1", Path: "/usr/local/php/bin/8.1", Type: types.ArtifactTreeNodeTypeVersion, Format: types.ArtifactRawFormat},
		{OwnerID: 1, Name: "4.0.1", Path: "/river/4.0.1", Type: types.ArtifactTreeNodeTypeVersion, Format: types.ArtifactHelmFormat},
	}

	for id, test := range tests {
		err := suite.ormStore.Nodes().RecurseCreate(suite.Ctx, &test)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)

		obj, err := suite.ormStore.Nodes().GetByPath(suite.Ctx, test.OwnerID, test.Path, test.Format)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.Path, obj.Path, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.Name, obj.Name, testsuite.InvalidLoopMsgF, id)
	}
}
