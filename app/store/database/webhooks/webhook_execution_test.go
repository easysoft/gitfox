// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package webhooks_test

import (
	"context"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/repo"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableWebhookExec = "webhook_executions"
)

type WebhookExecutionSuite struct {
	testsuite.BaseSuite

	ormStore  *repo.WebhookExecutionStore
	sqlxStore *database.WebhookExecutionStore
}

func TestWebhookExecutionSuite(t *testing.T) {
	ctx := context.Background()

	st := &WebhookExecutionSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "webhook_executions",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = repo.NewWebhookExecutionOrmStore(st.Gdb)
		st.sqlxStore = database.NewWebhookExecutionStore(st.Sdb)

		// add init data
		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 3)

		whStore := repo.NewWebhookOrmStore(st.Gdb)
		addWebhook(st.Ctx, t, whStore, 1, 1, 1, enum.WebhookParentSpace)
		addWebhook(st.Ctx, t, whStore, 2, 1, 1, enum.WebhookParentRepo)
	}

	suite.Run(t, st)
}

func (suite *WebhookExecutionSuite) SetupTest() {
	suite.addData()
}

func (suite *WebhookExecutionSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableWebhookExec).Where("1 = 1").Delete(nil)
}

var testAddWebhookExecutionItems = []struct {
	id          int64
	webhookID   int64
	triggerType enum.WebhookTrigger
	triggerID   string
}{
	{id: 1, webhookID: 1, triggerType: enum.WebhookTriggerBranchCreated, triggerID: "t1"},
	{id: 2, webhookID: 1, triggerType: enum.WebhookTriggerTagCreated, triggerID: "t1"},
	{id: 3, webhookID: 2, triggerType: enum.WebhookTriggerBranchUpdated, triggerID: "t1"},
	{id: 4, webhookID: 1, triggerType: enum.WebhookTriggerBranchCreated, triggerID: "t1"},
	{id: 5, webhookID: 2, triggerType: enum.WebhookTriggerTagCreated, triggerID: "t1"},
	{id: 6, webhookID: 2, triggerType: enum.WebhookTriggerBranchCreated, triggerID: "t1"},
	{id: 7, webhookID: 1, triggerType: enum.WebhookTriggerPullReqCreated, triggerID: "t1"},
}

func (suite *WebhookExecutionSuite) addData() {
	now := time.Now().UnixMilli()
	for id, item := range testAddWebhookExecutionItems {
		obj := types.WebhookExecution{
			ID:          item.id,
			WebhookID:   item.webhookID,
			TriggerType: item.triggerType,
			TriggerID:   item.triggerID,
			Created:     now,
		}

		err := suite.ormStore.Create(suite.Ctx, &obj)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *WebhookExecutionSuite) TestFind() {
	for id, test := range testAddWebhookExecutionItems {
		obj, err := suite.ormStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.webhookID, obj.WebhookID)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *WebhookExecutionSuite) TestDeleteOld() {
	num, err := suite.ormStore.DeleteOld(suite.Ctx, time.Now().Add(time.Second))
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), len(testAddWebhookExecutionItems), num)

	// find deleted objs
	_, err = suite.ormStore.Find(suite.Ctx, 2)
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.Find(suite.Ctx, 1)
	require.Error(suite.T(), err)

	//
	num, err = suite.ormStore.DeleteOld(suite.Ctx, time.Now().Add(time.Second))
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), 0, num)
}

func (suite *WebhookExecutionSuite) TestListForWebhook() {
	tests := []struct {
		webhookID  int64
		filter     types.WebhookExecutionFilter
		wantLength int
	}{
		{webhookID: 1, filter: types.WebhookExecutionFilter{}, wantLength: 4},
		{webhookID: 2, filter: types.WebhookExecutionFilter{}, wantLength: 3},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.ListForWebhook(suite.Ctx, test.webhookID, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.ListForWebhook(suite.Ctx, test.webhookID, &test.filter)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *WebhookExecutionSuite) TestListForTrigger() {
	tests := []struct {
		triggerId  string
		wantLength int
	}{
		{triggerId: "t1", wantLength: 7},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.ListForTrigger(suite.Ctx, test.triggerId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.ListForTrigger(suite.Ctx, test.triggerId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}
