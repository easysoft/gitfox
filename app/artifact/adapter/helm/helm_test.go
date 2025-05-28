// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package helm_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/helm"
	"github.com/easysoft/gitfox/app/artifact/adapter/testsuite"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HelmSuite struct {
	testsuite.BaseSuite
}

func TestHelmSuite(t *testing.T) {
	ctx := context.Background()

	st := &HelmSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "helm",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
	}

	suite.Run(t, st)
}

func (suite *HelmSuite) SetupTest() {
}

func (suite *HelmSuite) TearDownTest() {
}

func (suite *HelmSuite) TestUpload() {
	tests := []struct {
		Form      testsuite.MultiPartForm
		expectErr error
	}{
		{Form: testsuite.MultiPartForm{
			Files: map[string]string{"chart": "testdata/gitlab-2022.8.3101.tgz"}, TrimFileName: true,
		}, expectErr: nil},
		{Form: testsuite.MultiPartForm{
			Files: map[string]string{"chart": "testdata/2fauth-2023.2.801.tgz"}, TrimFileName: true,
		}, expectErr: adapter.ErrInvalidPackageName},
		{Form: testsuite.MultiPartForm{
			Files: map[string]string{"chart": "testdata/twofauth-2023.2.801.tgz"}, TrimFileName: true,
		}, expectErr: adapter.ErrInvalidPackageContent},
	}

	for _, test := range tests {
		body, cType, err := testsuite.CreateMultiPartForm(&test.Form)
		require.NoError(suite.T(), err)

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", body)
		require.NoError(suite.T(), err)
		req.Header.Set("Content-Type", cType)

		uploader := helm.NewUploader(suite.ArtifactStore, suite.Store.Artifacts, suite.DefaultView)
		err = handleUpload(suite.Ctx, uploader, req)
		if test.expectErr != nil {
			require.ErrorContains(suite.T(), err, test.expectErr.Error())
		} else {
			require.NoError(suite.T(), err)
		}
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
