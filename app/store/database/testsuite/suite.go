// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package testsuite

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/cache"
	"github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/app/store/database/repo"
	"github.com/easysoft/gitfox/app/store/database/space"
	cache2 "github.com/easysoft/gitfox/cache"
	"github.com/easysoft/gitfox/store/database/dbtest"
	"github.com/easysoft/gitfox/types"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

const (
	InvalidLoopMsgF = "\"invalid loop id: %d\""
)

type TestStore struct {
	Principal      store.PrincipalStore
	PrincipalCache store.PrincipalInfoCache
	Space          store.SpaceStore
	SpacePath      store.SpacePathStore
	Repo           store.RepoStore
	check          store.CheckStore
}

func newTestStore(sdb *sqlx.DB, gdb *gorm.DB) *TestStore {
	principalStore := principal.NewPrincipalOrmStore(gdb, store.ToLowerPrincipalUIDTransformation)
	principalViewInfo := principal.NewPrincipalOrmInfoView(gdb)
	pCache := cache2.NewExtended[int64, *types.PrincipalInfo](principalViewInfo, 30*time.Second)

	spacePathTransformation := store.ToLowerSpacePathTransformation
	spacePathStore := space.NewSpacePathOrmStore(gdb, store.ToLowerSpacePathTransformation)
	spacePathCache := cache.New(spacePathStore, spacePathTransformation)

	spaceStore := space.NewSpaceOrmStore(gdb, spacePathCache, spacePathStore)
	repoStore := repo.NewRepoOrmStore(gdb, spacePathCache, spacePathStore, spaceStore)

	return &TestStore{
		Principal:      principalStore,
		PrincipalCache: pCache,
		Space:          spaceStore,
		SpacePath:      spacePathStore,
		Repo:           repoStore,
	}
}

type Constructor func(ts *TestStore)

type Teardown func() error

type BaseSuite struct {
	suite.Suite

	Constructor Constructor
	Teardown    Teardown

	Ctx  context.Context
	Name string

	Store *TestStore
	Sdb   *sqlx.DB
	Gdb   *gorm.DB
}

// SetupSuite implements [suite.SetupAllSuite] interface.
func (suite *BaseSuite) SetupSuite() {
	sdb, gdb := dbtest.New(suite.Ctx, suite.T(), suite.Name)
	suite.Sdb = sdb
	suite.Gdb = gdb
	suite.Store = newTestStore(sdb, gdb)

	suite.Constructor(suite.Store)
}

// TearDownSuite implements [suite.TearDownAllSuite].
func (suite *BaseSuite) TearDownSuite() {
	if suite.Teardown != nil {
		suite.Require().NoError(suite.Teardown())
	}
}

func AddUser(
	ctx context.Context,
	t *testing.T,
	principalStore store.PrincipalStore,
	id int64, admin bool,
) {
	t.Helper()

	uid := "user_" + strconv.FormatInt(id, 10)
	if err := principalStore.CreateUser(ctx,
		&types.User{
			ID: id, UID: uid, Email: uid + "@local.dev", DisplayName: strings.ToUpper(uid), Admin: admin,
			Created: time.Now().UnixMilli(), Updated: time.Now().UnixMilli(),
		}); err != nil {
		t.Fatalf("failed to create user '%s': %v", uid, err)
	}
}

func AddSpace(
	ctx context.Context,
	t *testing.T,
	spaceStore store.SpaceStore,
	spacePathStore store.SpacePathStore,
	userID int64,
	spaceID int64,
	parentID int64,
) {
	t.Helper()

	identifier := "space_" + strconv.FormatInt(spaceID, 10)

	space := types.Space{ID: spaceID, Identifier: identifier, CreatedBy: userID, ParentID: parentID}
	if err := spaceStore.Create(ctx, &space); err != nil {
		t.Fatalf("failed to create space %v", err)
	}

	if err := spacePathStore.InsertSegment(ctx, &types.SpacePathSegment{
		ID: space.ID, Identifier: identifier, ParentID: parentID, CreatedBy: userID, SpaceID: spaceID, IsPrimary: true,
	}); err != nil {
		t.Fatalf("failed to insert segment %v", err)
	}
}

func AddRepo(
	ctx context.Context,
	t *testing.T,
	repoStore store.RepoStore,
	id int64,
	spaceID int64,
	size int64,
) {
	t.Helper()

	identifier := "repo_" + strconv.FormatInt(id, 10)
	repo := types.Repository{Identifier: identifier, ID: id, ParentID: spaceID, GitUID: identifier, Size: size}
	if err := repoStore.Create(ctx, &repo); err != nil {
		t.Fatalf("failed to create repo %v", err)
	}
}
