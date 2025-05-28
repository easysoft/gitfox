// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package webhooks_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/repo"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableWebhook = "webhooks"
)

type WebhookSuite struct {
	testsuite.BaseSuite

	ormStore  *repo.WebhookStore
	sqlxStore *database.WebhookStore
}

func TestWebhookSuite(t *testing.T) {
	ctx := context.Background()

	st := &WebhookSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "webhooks",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = repo.NewWebhookOrmStore(st.Gdb)
		st.sqlxStore = database.NewWebhookStore(st.Sdb)

		// add init data
		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 2, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 3)
	}

	suite.Run(t, st)
}

func (suite *WebhookSuite) SetupTest() {
	suite.addData()
}

func (suite *WebhookSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableWebhook).Where("1 = 1").Delete(nil)
}

var testAddItems = []struct {
	id         int64
	parentID   int64
	parentType enum.WebhookParent
	createdBy  int64
	identifier string
}{
	{id: 1, parentID: 1, parentType: enum.WebhookParentSpace, createdBy: 1, identifier: "wh_space_1"},
	{id: 2, parentID: 1, parentType: enum.WebhookParentSpace, createdBy: 1, identifier: "wh_space_2"},
	{id: 3, parentID: 2, parentType: enum.WebhookParentSpace, createdBy: 1, identifier: "wh_space_3"},
	{id: 4, parentID: 1, parentType: enum.WebhookParentRepo, createdBy: 1, identifier: "wh_repo_1"},
	{id: 5, parentID: 1, parentType: enum.WebhookParentSpace, createdBy: 1, identifier: "wh_repo_2"},
}

func (suite *WebhookSuite) addData() {
	now := time.Now().UnixMilli()
	for id, item := range testAddItems {
		obj := types.Webhook{
			ID:         item.id,
			ParentID:   item.parentID,
			ParentType: item.parentType,
			CreatedBy:  item.createdBy,
			Created:    now,
			Updated:    now,
			Identifier: item.identifier,
		}

		err := suite.ormStore.Create(suite.Ctx, &obj)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *WebhookSuite) TestFind() {
	for id, test := range testAddItems {
		obj, err := suite.ormStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.identifier, obj.Identifier)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *WebhookSuite) TestFindByIdentifier() {
	for id, test := range testAddItems {
		obj, err := suite.ormStore.FindByIdentifier(suite.Ctx, test.parentType, test.parentID, test.identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.id, obj.ID)

		objB, err := suite.sqlxStore.FindByIdentifier(suite.Ctx, test.parentType, test.parentID, test.identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *WebhookSuite) TestDelete() {
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

func (suite *WebhookSuite) TestDeleteByIdentifier() {
	err := suite.ormStore.DeleteByIdentifier(suite.Ctx, enum.WebhookParentSpace, 1, "wh_space_1")
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, enum.WebhookParentRepo, 1, "wh_repo_1")
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.FindByIdentifier(suite.Ctx, enum.WebhookParentRepo, 1, "wh_repo_1")
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.FindByIdentifier(suite.Ctx, enum.WebhookParentSpace, 1, "wh_space_1")
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, enum.WebhookParentRepo, 1, "wh_repo_1")
	require.NoError(suite.T(), err)

	err = suite.ormStore.DeleteByIdentifier(suite.Ctx, enum.WebhookParentSpace, 1, "wh_space_1")
	require.NoError(suite.T(), err)
}

func addWebhook(ctx context.Context, t *testing.T, s store.WebhookStore,
	id, parentId, createdBy int64, parentType enum.WebhookParent) {
	obj := types.Webhook{
		ID:         id,
		ParentID:   parentId,
		ParentType: parentType,
		CreatedBy:  createdBy,
		Created:    time.Now().UnixMilli(),
		Updated:    time.Now().UnixMilli(),
		Identifier: fmt.Sprintf("webhook_%d", id),
	}

	err := s.Create(ctx, &obj)
	require.NoError(t, err)
}
