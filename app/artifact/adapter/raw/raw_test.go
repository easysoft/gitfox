// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package raw_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/raw"
	"github.com/easysoft/gitfox/app/artifact/adapter/testsuite"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RawSuite struct {
	testsuite.BaseSuite
}

func TestRawSuite(t *testing.T) {
	ctx := context.Background()

	st := &RawSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "raw",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
	}

	suite.Run(t, st)
}

func (suite *RawSuite) SetupTest() {
}

func (suite *RawSuite) TearDownTest() {
}

func (suite *RawSuite) TestUpload() {
	invalidTests := []struct {
		Form      testsuite.MultiPartForm
		expectErr error
	}{
		{Form: testsuite.MultiPartForm{
			Files: map[string]string{"file": "testdata/install.sh"}, TrimFileName: true,
		}, expectErr: adapter.ErrMissFormField},
		{Form: testsuite.MultiPartForm{
			Fields: map[string]string{"name": "gitfox"},
			Files:  map[string]string{"file": "testdata/install.sh"}, TrimFileName: true,
		}, expectErr: adapter.ErrMissFormField},
		{Form: testsuite.MultiPartForm{
			Fields: map[string]string{"version": "1.0.0"},
			Files:  map[string]string{"file": "testdata/install.sh"}, TrimFileName: true,
		}, expectErr: adapter.ErrMissFormField},
		{Form: testsuite.MultiPartForm{
			Fields: map[string]string{"name": ".gitfox", "version": "1.0.0"},
			Files:  map[string]string{"file": "testdata/install.sh"}, TrimFileName: true,
		}, expectErr: adapter.ErrInvalidPackageName},
		{Form: testsuite.MultiPartForm{
			Fields: map[string]string{"name": "gitfox&", "version": "1.0.0"},
			Files:  map[string]string{"file": "testdata/install.sh"}, TrimFileName: true,
		}, expectErr: adapter.ErrInvalidPackageName},
		{Form: testsuite.MultiPartForm{
			Fields: map[string]string{"name": "gitfox", "version": "1.0.0/"},
			Files:  map[string]string{"file": "testdata/install.sh"}, TrimFileName: true,
		}, expectErr: adapter.ErrInvalidPackageVersion},
		{Form: testsuite.MultiPartForm{
			Fields: map[string]string{"name": "gitfox", "version": "^1.0.0"},
			Files:  map[string]string{"file": "testdata/install.sh"}, TrimFileName: true,
		}, expectErr: adapter.ErrInvalidPackageVersion},
		{Form: testsuite.MultiPartForm{
			Fields: map[string]string{"name": "gitfox", "version": "1.0.0", "group": "a.9 "},
			Files:  map[string]string{"file": "testdata/install.sh"}, TrimFileName: true,
		}, expectErr: adapter.ErrInvalidGroupName},
		{Form: testsuite.MultiPartForm{
			Fields: map[string]string{"name": "gitfox", "version": "1.0.0", "group": "a.9+"},
			Files:  map[string]string{"file": "testdata/install.sh"}, TrimFileName: true,
		}, expectErr: adapter.ErrInvalidGroupName},
	}

	successTests := []testsuite.MultiPartForm{
		{Fields: map[string]string{"name": "gitfox", "group": "easycorp.pangu", "version": "1.0.0"},
			Files: map[string]string{"file": "testdata/install.sh"}, TrimFileName: true},
		{Fields: map[string]string{"name": "gitfox", "group": "", "version": "1.0.0"},
			Files: map[string]string{"file": "testdata/install.sh"}, TrimFileName: true},
		{Fields: map[string]string{"name": "a", "group": "g", "version": "1"},
			Files: map[string]string{"file": "testdata/install.sh"}, TrimFileName: true},
		{Fields: map[string]string{"name": "a_b-c", "group": "g1.g_2.g-3", "version": "1.0.0-beta1"},
			Files: map[string]string{"file": "testdata/install.sh"}, TrimFileName: true},
	}

	for idx, test := range invalidTests {
		body, cType, err := testsuite.CreateMultiPartForm(&test.Form)
		require.NoError(suite.T(), err)

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", body)
		require.NoError(suite.T(), err)
		req.Header.Set("Content-Type", cType)

		uploader := raw.NewUploader(suite.ArtifactStore, suite.Store.Artifacts, suite.DefaultView)
		err = handleUpload(suite.Ctx, uploader, req)
		require.ErrorContains(suite.T(), err, test.expectErr.Error(), "loop id: %d", idx)
	}

	for idx, test := range successTests {
		body, cType, err := testsuite.CreateMultiPartForm(&test)
		require.NoError(suite.T(), err)

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", body)
		require.NoError(suite.T(), err)
		req.Header.Set("Content-Type", cType)

		uploader := raw.NewUploader(suite.ArtifactStore, suite.Store.Artifacts, suite.DefaultView)
		err = handleUpload(suite.Ctx, uploader, req)
		require.NoError(suite.T(), err, "loop id: %d", idx)
	}
}

func handleUpload(ctx context.Context, uploader adapter.ArtifactPackageUploader, req *http.Request) error {
	_, err := uploader.Serve(ctx, req)
	if err != nil {
		return err
	}

	err = uploader.IsValid(ctx)
	if err != nil {
		return err
	}

	err = uploader.Save(ctx)
	if err != nil {
		return err
	}

	return nil
}
