// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package testsuite

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"os"
	"path"
	"time"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/store/cache"
	"github.com/easysoft/gitfox/app/store/database/artifacts"
	"github.com/easysoft/gitfox/app/store/database/principal"
	"github.com/easysoft/gitfox/app/store/database/repo"
	"github.com/easysoft/gitfox/app/store/database/space"
	cache2 "github.com/easysoft/gitfox/cache"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/store/database/dbtest"
	"github.com/easysoft/gitfox/types"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TestStore struct {
	Principal      store.PrincipalStore
	PrincipalCache store.PrincipalInfoCache
	Space          store.SpaceStore
	SpacePath      store.SpacePathStore
	Repo           store.RepoStore
	check          store.CheckStore

	Artifacts store.ArtifactStore
}

func newTestStore(gdb *gorm.DB) *TestStore {
	principalStore := principal.NewPrincipalOrmStore(gdb, store.ToLowerPrincipalUIDTransformation)
	principalViewInfo := principal.NewPrincipalOrmInfoView(gdb)
	pCache := cache2.NewExtended[int64, *types.PrincipalInfo](principalViewInfo, 30*time.Second)

	spacePathTransformation := store.ToLowerSpacePathTransformation
	spacePathStore := space.NewSpacePathOrmStore(gdb, store.ToLowerSpacePathTransformation)
	spacePathCache := cache.New(spacePathStore, spacePathTransformation)

	spaceStore := space.NewSpaceOrmStore(gdb, spacePathCache, spacePathStore)
	repoStore := repo.NewRepoOrmStore(gdb, spacePathCache, spacePathStore, spaceStore)

	artifactStore := artifacts.NewStore(gdb)

	return &TestStore{
		Principal:      principalStore,
		PrincipalCache: pCache,
		Space:          spaceStore,
		SpacePath:      spacePathStore,
		Repo:           repoStore,
		Artifacts:      artifactStore,
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
	Gdb   *gorm.DB

	artifactRootDir string
	ArtifactStore   storage.ContentStorage

	DefaultView *adapter.ViewDescriptor
}

// SetupSuite implements [suite.SetupAllSuite] interface.
func (suite *BaseSuite) SetupSuite() {
	_, gdb := dbtest.New(suite.Ctx, suite.T(), suite.Name)
	suite.Gdb = gdb
	suite.Store = newTestStore(gdb)

	rootDir, err := os.MkdirTemp("", "*")
	suite.Require().NoError(err)
	suite.artifactRootDir = rootDir

	driver, err := storage.NewDriver(suite.Ctx, "filesystem", map[string]interface{}{"rootdirectory": rootDir})
	suite.Require().NoError(err)

	suite.ArtifactStore, err = storage.NewCommonContentStore(suite.Ctx, driver)
	suite.Require().NoError(err)

	suite.DefaultView = &adapter.ViewDescriptor{
		ViewID:    1,
		OwnerID:   1,
		Store:     suite.ArtifactStore,
		StorageID: 1,
	}
	suite.Constructor(suite.Store)
}

// TearDownSuite implements [suite.TearDownAllSuite].
func (suite *BaseSuite) TearDownSuite() {
	_ = os.RemoveAll(suite.artifactRootDir)
	if suite.Teardown != nil {
		suite.Require().NoError(suite.Teardown())
	}
}

type MultiPartForm struct {
	Fields       map[string]string
	Files        map[string]string
	TrimFileName bool
}

func CreateMultiPartForm(m *MultiPartForm) (io.Reader, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for field, value := range m.Fields {
		err := writer.WriteField(field, value)
		if err != nil {
			return nil, "", err
		}
	}

	for field, fileName := range m.Files {
		uploadFileName := fileName
		if m.TrimFileName {
			uploadFileName = path.Base(fileName)
		}
		part, err := writer.CreateFormFile(field, uploadFileName)
		if err != nil {
			return nil, "", err
		}
		file, err := os.Open(fileName)
		if err != nil {
			return nil, "", err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, "", err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}
	return body, writer.FormDataContentType(), nil
}
