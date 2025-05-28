// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/app/store/database/repo"
	"github.com/easysoft/gitfox/app/store/database/space"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/cache"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableRule = "rules"

	testRuleTypeMonitor types.RuleType = "monitor"
	testRuleTypeBranch  types.RuleType = "branch"
	testRuleTypeTag     types.RuleType = "tag"
)

var _ruleRepoMap = `
|-- space_1
|   |-> repo_1
|   |-- space_2
|   |   |-> repo_2

rule1 -> nil, nil
rule2 -> 1,1
rule3 -> 2,2
rule4 -> 2,2
`

type RuleSuite struct {
	testsuite.BaseSuite

	ormStore  *repo.RuleStore
	sqlxStore *database.RuleStore
}

func TestRuleSuite(t *testing.T) {
	ctx := context.Background()

	st := &RuleSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "rules",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		ormPrincipalView := principal.NewPrincipalOrmInfoView(st.Gdb)
		ormPrincipalCache := cache.NewExtended[int64, *types.PrincipalInfo](ormPrincipalView, 30*time.Second)

		sqlxPrincipalView := database.NewPrincipalInfoView(st.Sdb)
		sqlxPrincipalCache := cache.NewExtended[int64, *types.PrincipalInfo](sqlxPrincipalView, 30*time.Second)

		st.ormStore = repo.NewRuleOrmStore(st.Gdb, ormPrincipalCache)
		st.sqlxStore = database.NewRuleStore(st.Sdb, sqlxPrincipalCache)

		ormPathStore := space.NewSpacePathOrmStore(st.Gdb, store.ToLowerSpacePathTransformation)

		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)

		testsuite.AddSpace(st.Ctx, t, ts.Space, ormPathStore, 1, 1, 0)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ormPathStore, 1, 2, 1)

		testsuite.AddRepo(st.Ctx, t, ts.Repo, 1, 1, 10)
		testsuite.AddRepo(st.Ctx, t, ts.Repo, 2, 2, 11)
	}

	suite.Run(t, st)
}

func (suite *RuleSuite) SetupTest() {
	suite.addData()
}

func (suite *RuleSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableRule).Where("1 = 1").Delete(nil)
}

func (suite *RuleSuite) addData() {
	_ = _ruleRepoMap

	now := time.Now().UnixMilli()
	addItems := []types.Rule{
		{ID: 1, Identifier: "rule_1", Type: testRuleTypeMonitor, State: enum.RuleStateMonitor, CreatedBy: 1, Created: now, Updated: now},
		{ID: 2, SpaceID: ptrInt64(1), RepoID: ptrInt64(1), Identifier: "rule_2", Type: testRuleTypeBranch, State: enum.RuleStateActive, CreatedBy: 1, Created: now, Updated: now},
		{ID: 3, SpaceID: ptrInt64(2), RepoID: ptrInt64(2), Identifier: "rule_3", Type: testRuleTypeBranch, State: enum.RuleStateActive, CreatedBy: 1, Created: now, Updated: now},
		{ID: 4, SpaceID: ptrInt64(2), RepoID: ptrInt64(2), Identifier: "rule_4", Type: testRuleTypeTag, State: enum.RuleStateActive, CreatedBy: 1, Created: now, Updated: now},
	}

	for id, item := range addItems {
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func ptrInt64(i int64) *int64 {
	return &i
}

func (suite *RuleSuite) TestFind() {
	tests := []struct {
		Id         int64
		RepoID     *int64
		Identifier string
	}{
		{Id: 1, RepoID: nil, Identifier: "rule_1"},
		{Id: 2, RepoID: ptrInt64(1), Identifier: "rule_2"},
		{Id: 3, RepoID: ptrInt64(2), Identifier: "rule_3"},
		{Id: 4, RepoID: ptrInt64(2), Identifier: "rule_4"},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.Find(suite.Ctx, test.Id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.Identifier, obj.Identifier)
		if test.RepoID == nil {
			require.Nil(suite.T(), obj.RepoID)
		} else {
			require.Equal(suite.T(), test.RepoID, obj.RepoID)
		}

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.Id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RuleSuite) TestFindByIdentifier() {
	tests := []struct {
		Id         int64
		SpaceID    *int64
		RepoID     *int64
		Identifier string
	}{
		{Id: 1, SpaceID: nil, RepoID: nil, Identifier: "rule_1"},
		{Id: 2, SpaceID: ptrInt64(1), RepoID: ptrInt64(1), Identifier: "rule_2"},
		{Id: 3, SpaceID: ptrInt64(2), RepoID: ptrInt64(2), Identifier: "rule_3"},
		{Id: 4, SpaceID: ptrInt64(2), RepoID: ptrInt64(2), Identifier: "Rule_4"},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.FindByIdentifier(suite.Ctx, test.SpaceID, test.RepoID, test.Identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.Id, obj.ID)

		objB, err := suite.sqlxStore.FindByIdentifier(suite.Ctx, test.SpaceID, test.RepoID, test.Identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RuleSuite) TestUpdate() {
	now := time.Now().UnixMilli() - 1
	updates := []types.Rule{
		// do nothing
		{ID: 1, Identifier: "rule_1", Type: testRuleTypeMonitor, State: enum.RuleStateMonitor, CreatedBy: 1, Created: now},
		// change Identifier
		{ID: 2, SpaceID: ptrInt64(1), RepoID: ptrInt64(1), Identifier: "rule_2+2", Type: testRuleTypeBranch, State: enum.RuleStateActive, CreatedBy: 1, Created: now},
		// change state
		{ID: 3, SpaceID: ptrInt64(2), RepoID: ptrInt64(2), Identifier: "rule_3", Type: testRuleTypeBranch, State: enum.RuleStateDisabled, CreatedBy: 1, Created: now},
		// change pattern
		{ID: 4, SpaceID: ptrInt64(2), RepoID: ptrInt64(2), Identifier: "rule_4", Type: testRuleTypeTag, State: enum.RuleStateActive, CreatedBy: 1, Created: now, Pattern: []byte("{\"a\":1}")},
	}

	updates2 := []types.Rule{
		// change type, version is required, but Type won't be changed
		{ID: 1, Version: 1, Identifier: "rule_1", Type: testRuleTypeBranch, State: enum.RuleStateMonitor, CreatedBy: 1, Created: now},
	}

	for id, up := range updates {
		err := suite.ormStore.Update(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	for id, up := range updates2 {
		err := suite.sqlxStore.Update(suite.Ctx, &up)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}

	tests := []struct {
		ID        int64
		fieldName string
		expect    interface{}
	}{
		{ID: 1, expect: testRuleTypeMonitor, fieldName: "Type"},
		{ID: 1, expect: 2, fieldName: "Version"},
		{ID: 2, expect: "rule_2+2", fieldName: "Identifier"},
		{ID: 3, expect: enum.RuleStateDisabled, fieldName: "State"},
		{ID: 4, expect: []byte("{\"a\":1}"), fieldName: "Pattern"},
	}

	for id, test := range tests {
		obj, err := suite.ormStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		testsuite.EqualFieldValue(suite.T(), test.expect, obj, test.fieldName, testsuite.InvalidLoopMsgF, id)
		require.Greater(suite.T(), obj.Version, int64(0))
		require.Greater(suite.T(), obj.Updated, now)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.ID)
		require.NoError(suite.T(), err)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RuleSuite) TestDelete() {
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

func (suite *RuleSuite) TestDeleteByIdentifier() {
	err := suite.ormStore.DeleteByIdentifier(suite.Ctx, nil, nil, "rule_1")
	require.NoError(suite.T(), err)

	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, ptrInt64(1), ptrInt64(1), "rule_2")
	require.NoError(suite.T(), err)

	// find deleted objs
	_, err = suite.ormStore.Find(suite.Ctx, 2)
	require.Error(suite.T(), err)

	_, err = suite.sqlxStore.Find(suite.Ctx, 1)
	require.Error(suite.T(), err)

	// no rows affected doesn't return err now
	err = suite.sqlxStore.DeleteByIdentifier(suite.Ctx, nil, nil, "rule_1+")
	require.NoError(suite.T(), err)

	err = suite.ormStore.DeleteByIdentifier(suite.Ctx, nil, ptrInt64(2), "rule_2+")
	require.NoError(suite.T(), err)
}

func (suite *RuleSuite) TestListAndCount() {
	_ = _ruleRepoMap
	tests := []struct {
		spaceId    *int64
		repoId     *int64
		filter     types.RuleFilter
		wantLength int64
	}{
		{spaceId: nil, repoId: nil, filter: types.RuleFilter{}, wantLength: 4},
		{spaceId: ptrInt64(1), repoId: nil, filter: types.RuleFilter{}, wantLength: 1},
		{spaceId: nil, repoId: ptrInt64(1), filter: types.RuleFilter{}, wantLength: 1},
		{spaceId: ptrInt64(1), repoId: ptrInt64(1), filter: types.RuleFilter{}, wantLength: 1},
		{spaceId: ptrInt64(2), filter: types.RuleFilter{}, wantLength: 2},
		{spaceId: ptrInt64(2), filter: types.RuleFilter{
			States: []enum.RuleState{enum.RuleStateMonitor},
		}, wantLength: 0},
		{spaceId: ptrInt64(2), filter: types.RuleFilter{
			States: []enum.RuleState{enum.RuleStateMonitor, enum.RuleStateActive},
		}, wantLength: 2},
		{spaceId: ptrInt64(2), filter: types.RuleFilter{
			States: []enum.RuleState{enum.RuleStateMonitor, enum.RuleStateActive, enum.RuleStateDisabled},
		}, wantLength: 2},
	}

	for id, test := range tests {
		rules, err := suite.ormStore.List(suite.Ctx, test.spaceId, test.repoId, &test.filter)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(rules), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.Count(suite.Ctx, test.spaceId, test.repoId, &test.filter)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		rulesB, err := suite.sqlxStore.List(suite.Ctx, test.spaceId, test.repoId, &test.filter)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), rules, rulesB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *RuleSuite) TestListAllRepoRules() {
	for _, repoId := range []int64{1, 2} {
		infos, err := suite.ormStore.ListAllRepoRules(suite.Ctx, repoId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, repoId)

		infosB, err := suite.sqlxStore.ListAllRepoRules(suite.Ctx, repoId)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, repoId)
		require.ElementsMatch(suite.T(), infos, infosB, testsuite.InvalidLoopMsgF, repoId)
	}
}
