// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package principal_test

import (
	"context"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableToken = "tokens"
)

type TokenSuite struct {
	testsuite.BaseSuite

	ormStore  *principal.TokenStore
	sqlxStore *database.TokenStore
}

func TestTokenSuite(t *testing.T) {
	ctx := context.Background()

	st := &TokenSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "tokens",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = principal.NewTokenOrmStore(st.Gdb)
		st.sqlxStore = database.NewTokenStore(st.Sdb)

		// add init data
		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 3, true)
	}

	suite.Run(t, st)
}

func (suite *TokenSuite) SetupTest() {
	suite.addData()
}

func (suite *TokenSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableToken).Where("1 = 1").Delete(nil)
}

var testAddTokenItems = []struct {
	id            int64
	principalID   int64
	tokenType     enum.TokenType
	expiredMinute int
	identifier    string
}{
	{id: 1, principalID: 1, tokenType: enum.TokenTypeSAT, expiredMinute: 0, identifier: "token_1"},
	{id: 2, principalID: 1, tokenType: enum.TokenTypePAT, expiredMinute: 1, identifier: "token_2"},
	{id: 3, principalID: 2, tokenType: enum.TokenTypeSession, expiredMinute: 5, identifier: "token_3"},
	{id: 4, principalID: 3, tokenType: enum.TokenTypeSAT, expiredMinute: 10, identifier: "token_4"},
	{id: 5, principalID: 2, tokenType: enum.TokenTypePAT, expiredMinute: 30, identifier: "token_5"},
}

func (suite *TokenSuite) addData() {
	now := time.Now()
	for id, item := range testAddTokenItems {
		obj := types.Token{
			ID:          item.id,
			PrincipalID: item.principalID,
			Type:        item.tokenType,
			Identifier:  item.identifier,
			ExpiresAt:   nil,
			CreatedBy:   1,
		}
		if item.expiredMinute > 0 {
			expTime := now.Add(time.Duration(item.expiredMinute) * time.Minute).UnixMilli()
			obj.ExpiresAt = &expTime
		}

		err := suite.ormStore.Create(suite.Ctx, &obj)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TokenSuite) TestFind() {
	for id, test := range testAddTokenItems {
		obj, err := suite.ormStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.identifier, obj.Identifier)

		objB, err := suite.sqlxStore.Find(suite.Ctx, test.id)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TokenSuite) TestFindByIdentifier() {
	for id, test := range testAddTokenItems {
		obj, err := suite.ormStore.FindByIdentifier(suite.Ctx, test.principalID, test.identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.Equal(suite.T(), test.id, obj.ID)

		objB, err := suite.sqlxStore.FindByIdentifier(suite.Ctx, test.principalID, test.identifier)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualExportedValues(suite.T(), *obj, *objB, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TokenSuite) TestDelete() {
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

func (suite *TokenSuite) TestDeleteExpiredBefore() {
	now := time.Now().Add(time.Second)
	tests := []struct {
		before     time.Time
		tokenTypes []enum.TokenType
		wantLength int64
	}{
		{before: now, tokenTypes: []enum.TokenType{}, wantLength: 0},
		{before: now.Add(time.Minute), tokenTypes: []enum.TokenType{}, wantLength: 1},
		{before: now.Add(5 * time.Minute), tokenTypes: []enum.TokenType{enum.TokenTypePAT}, wantLength: 0},
		{before: now.Add(10 * time.Minute), tokenTypes: []enum.TokenType{enum.TokenTypeSAT, enum.TokenTypeSession}, wantLength: 2},
	}

	for id, test := range tests {
		num, err := suite.ormStore.DeleteExpiredBefore(suite.Ctx, test.before, test.tokenTypes)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
		require.EqualValues(suite.T(), test.wantLength, num, testsuite.InvalidLoopMsgF, id)
	}

	remains := []int64{1, 5}
	for id, pk := range remains {
		_, err := suite.ormStore.Find(suite.Ctx, pk)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

func (suite *TokenSuite) TestListAncCount() {
	tests := []struct {
		principalId int64
		tokenType   enum.TokenType
		wantLength  int64
	}{
		{1, enum.TokenTypePAT, 1},
		{1, enum.TokenTypeSAT, 1},
		{1, enum.TokenTypeSession, 0},
		{2, enum.TokenTypePAT, 1},
		{2, enum.TokenTypeSAT, 0},
		{2, enum.TokenTypeSession, 1},
		{3, enum.TokenTypePAT, 0},
		{3, enum.TokenTypeSAT, 1},
		{3, enum.TokenTypeSession, 0},
	}

	for id, test := range tests {
		objs, err := suite.ormStore.List(suite.Ctx, test.principalId, test.tokenType)
		require.NoError(suite.T(), err)
		require.EqualValues(suite.T(), test.wantLength, len(objs), testsuite.InvalidLoopMsgF, id)

		count, err := suite.ormStore.Count(suite.Ctx, test.principalId, test.tokenType)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), test.wantLength, count, testsuite.InvalidLoopMsgF, id)

		objsB, err := suite.sqlxStore.List(suite.Ctx, test.principalId, test.tokenType)
		require.NoError(suite.T(), err)
		require.ElementsMatch(suite.T(), objs, objsB, testsuite.InvalidLoopMsgF, id)
	}
}
