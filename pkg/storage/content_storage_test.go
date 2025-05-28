// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"io"
	"os"
	"path"
	"testing"

	storagedriver "github.com/easysoft/gitfox/pkg/storage/driver"

	"github.com/mailru/easyjson/buffer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestBuildPath(t *testing.T) {
	tests := []struct {
		path       string
		prefixPath string
		expect     string
	}{
		{"a", "", "/a"},
		{"a", "demo", "/demo/a"},
		{"a", "/demo", "/demo/a"},
		{"a/", "/demo", "/demo/a/"},
		{"", "", "/"},
		{"", "demo", "/demo/"},
		{"/", "", "/"},
		{"/", "/demo", "/demo/"},
		{"//abc", "", "//abc"},
		{"//abc", "/demo", "/demo//abc"},
		{"//abc/", "/demo", "/demo//abc/"},
	}

	var store *commonContentStore
	for _, test := range tests {
		var s ContentStorage
		if test.prefixPath != "" {
			s, _ = NewCommonContentStore(context.Background(), nil, prefixDirOption{test.prefixPath})
		} else {
			s, _ = NewCommonContentStore(context.Background(), nil)
		}

		store = s.(*commonContentStore)

		content := store.buildPath(test.path)
		assert.Equal(t, test.expect, content, "case: %+v", test)
	}
}

//func TestContentStorage(t *testing.T) {
//	ctx := context.Background()
//
//	storageDirectory, err := os.MkdirTemp(os.TempDir(), "*")
//	assert.NoError(t, err)
//
//	defer os.RemoveAll(storageDirectory)
//
//	var driverName = "filesystem"
//	var driverConfig = map[string]interface{}{
//		"rootdirectory": storageDirectory,
//		"maxthreads":    10,
//	}
//	var prefixDir = "demo"
//
//	driver, err := NewDriver(ctx, driverName, driverConfig)
//	assert.NoError(t, err)
//
//	cs, err := NewCommonContentStore(ctx, driver, prefixDirOption{prefixDir})
//	if !assert.NoError(t, err) {
//		return
//	}
//
//	textContext := []byte("hello world")
//	testPath := "path/to/testdir/init.txt"
//
//	err = cs.Put(ctx, testPath, textContext)
//	if !assert.NoError(t, err) {
//		return
//	}
//
//	finfo, err := cs.Stat(ctx, testPath)
//	if !assert.NoError(t, err) {
//		assert.Equal(t, storageDirectory+"/"+prefixDir+"/"+testPath, finfo.Path())
//	}
//
//	readContent, err := cs.Get(ctx, testPath)
//	if assert.NoError(t, err) {
//		assert.Equal(t, textContext, readContent)
//	}
//
//	fh, err := cs.Open(ctx, testPath)
//	if !assert.NoError(t, err) {
//		return
//	}
//
//	//readContent, err = io.ReadAll(fh)
//	//if assert.NoError(t, err) {
//	//	assert.Equal(t, textContext, readContent)
//	//}
//
//	w, err := os.CreateTemp(os.TempDir(), "write-*")
//	if !assert.NoError(t, err) {
//		return
//	}
//	n, err := io.Copy(w, fh)
//	if assert.NoError(t, err) {
//		assert.Equal(t, int64(len(textContext)), n)
//		e := fh.Close()
//		assert.Error(t, e)
//	}
//
//	readContent, err = os.ReadFile(w.Name())
//	if assert.NoError(t, err) {
//		assert.Equal(t, textContext, readContent)
//	}
//}

type contentStorageTestSuite struct {
	suite.Suite
	ctx         context.Context
	constructor constructor
	store       ContentStorage

	testContents []byte
}

type constructor func(suite *contentStorageTestSuite) error

func (suite *contentStorageTestSuite) SetupTest() {
	suite.testContents = []byte("hello world")

	err := suite.constructor(suite)
	suite.Require().NoError(err)
}

func (suite *contentStorageTestSuite) TestPathWithPrefixDir() {
	if suite.store.PrefixDir() == "" {
		return
	}

	var err error
	testPath := "a/b1"
	err = suite.store.Put(suite.ctx, testPath, suite.testContents)
	suite.Require().NoError(err)

	fInfo, err := suite.store.Stat(suite.ctx, testPath)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.store.PrefixDir()+"/"+testPath, fInfo.Path())

	err = suite.store.Delete(suite.ctx, path.Dir(testPath))
	suite.Require().NoError(err)

	_, err = suite.store.Stat(suite.ctx, path.Dir(testPath))
	suite.Require().ErrorAs(err, &storagedriver.PathNotFoundError{})
}

func (suite *contentStorageTestSuite) TestValidPath() {
	validFiles := []string{
		"/a",
		"/2",
		"/aa",
		"/a.a",
		"/0-9/abcdefg",
		"/abcdefg/z.75",
		"/abc/1.2.3.4.5-6_zyx/123.z/4",
		"/docker/docker-registry",
		"/123.abc",
		"/abc./abc",
		"/.abc",
		"/a--b",
		"/a-.b",
		"/_.abc",
		"/Docker/docker-registry",
		"/Abc/Cba",
	}

	for _, filename := range validFiles {
		err := suite.store.Put(suite.ctx, filename, suite.testContents)
		defer suite.store.Delete(suite.ctx, filename)
		suite.Require().NoError(err)

		received, err := suite.store.Get(suite.ctx, filename)
		suite.Require().NoError(err)
		suite.Require().Equal(suite.testContents, received)
	}
}

func (suite *contentStorageTestSuite) TestInvalidPath() {
	invalidFiles := []string{
		"",
		"/",
		"//bcd",
		"/abc_123/",
	}

	s := suite.store.(*commonContentStore)
	for _, filename := range invalidFiles {
		err := suite.store.Put(suite.ctx, filename, suite.testContents)
		// only delete if file was successfully written
		if err == nil {
			defer suite.store.Delete(suite.ctx, filename)
		}
		suite.Require().Errorf(err, "filename: %s, buildedPath: %s", filename, s.buildPath(filename))
		suite.Require().IsType(err, storagedriver.InvalidPathError{})

		_, err = suite.store.Get(suite.ctx, filename)
		suite.Require().Errorf(err, "path: %s", filename)
		suite.Require().IsType(err, storagedriver.InvalidPathError{})
	}
}

func (suite *contentStorageTestSuite) TestOpenFile() {
	testPath := "/TestOpen/data"
	var err error

	err = suite.store.Put(suite.ctx, testPath, suite.testContents)
	suite.Require().NoError(err)

	fh, err := suite.store.Open(suite.ctx, testPath)
	suite.Require().NoError(err)

	readContent, err := io.ReadAll(fh)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.testContents, readContent)

	err = suite.store.Delete(suite.ctx, path.Dir(testPath))
	suite.Require().NoError(err)
}

func (suite *contentStorageTestSuite) TestSaveFile() {
	testPath := "/TestSaveFile/data"
	var err error

	b := buffer.Buffer{}
	b.AppendBytes(suite.testContents)

	err = suite.store.Save(suite.ctx, testPath, b.ReadCloser(), int64(len(suite.testContents)))
	suite.Require().NoError(err)

	readContent, err := suite.store.Get(suite.ctx, testPath)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.testContents, readContent)

	err = suite.store.Delete(suite.ctx, path.Dir(testPath))
	suite.Require().NoError(err)
}

func (suite *contentStorageTestSuite) TestListPath() {
	testBaseDir := "/TestList"

	testSubFiles := []string{
		"a1",
		"a2",
		"c/c1",
		"b/b3",
		"b/b1",
	}

	tests := []struct {
		listRoot    string
		expectNames []string
	}{
		{testBaseDir, []string{
			path.Join(suite.store.PrefixDir(), testBaseDir, "a1"),
			path.Join(suite.store.PrefixDir(), testBaseDir, "a2"),
			path.Join(suite.store.PrefixDir(), testBaseDir, "b"),
			path.Join(suite.store.PrefixDir(), testBaseDir, "c"),
		}},
		{testBaseDir + "/b", []string{
			path.Join(suite.store.PrefixDir(), testBaseDir, "b", "b1"),
			path.Join(suite.store.PrefixDir(), testBaseDir, "b", "b3"),
		}},
	}
	var err error

	for _, fName := range testSubFiles {
		fileName := path.Join(testBaseDir, fName)
		err = suite.store.Put(suite.ctx, fileName, suite.testContents)
		suite.Require().NoError(err)
	}

	for _, test := range tests {
		receivePaths, err := suite.store.List(suite.ctx, test.listRoot)
		suite.Require().NoError(err)
		suite.Equal(test.expectNames, receivePaths)
	}

	err = suite.store.Delete(suite.ctx, testBaseDir)
	suite.Require().NoError(err)
}

func filesystemConstructor(ctx context.Context, rootDir, prefixDir string) func(suite *contentStorageTestSuite) error {
	return func(suite *contentStorageTestSuite) error {
		var cs ContentStorage
		var err error

		var driverConfig = map[string]interface{}{
			"rootdirectory": rootDir,
			"maxthreads":    10,
		}

		driver, err := NewDriver(ctx, "filesystem", driverConfig)
		if err != nil {
			return err
		}

		if prefixDir != "" {
			cs, err = NewCommonContentStore(ctx, driver, prefixDirOption{prefixDir})
		} else {
			cs, err = NewCommonContentStore(ctx, driver)
		}

		if err != nil {
			return err
		}

		suite.store = cs
		return nil
	}
}

func TestFileSystemContentStorageSuite(t *testing.T) {
	ctx := context.Background()
	rootDir, err := os.MkdirTemp(os.TempDir(), "*")
	require.NoError(t, err)
	//defer os.RemoveAll(rootDir)

	s1 := filesystemConstructor(ctx, rootDir, "")
	suite.Run(t, &contentStorageTestSuite{constructor: s1, ctx: context.Background()})

	s2 := filesystemConstructor(ctx, rootDir, "demo/")
	suite.Run(t, &contentStorageTestSuite{constructor: s2, ctx: context.Background()})
}
