// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox/store/database/dbtest"
	"github.com/easysoft/gitfox/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	t.Parallel()

	db := dbtest.NewDB(ctx, t, "artifact_store")
	s := NewStore(db)

	_ = s.Packages()
	_ = s.Versions()
	_ = s.Assets()
	_ = s.MetaAssets()
	_ = s.Blobs()
}

func TestFindPackages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	t.Parallel()
	tables := []any{new(types.ArtifactPackage), new(types.ArtifactVersion)}

	s := NewStore(dbtest.NewDB(ctx, t, "artifacts_store_0", tables...))

	inPkgs := []struct {
		Name   string
		Format types.ArtifactFormat
	}{
		{"raw1", types.ArtifactRawFormat},
		{"raw2", types.ArtifactRawFormat},
		{"maven1", types.ArtifactMavenFormat},
		{"raw3", types.ArtifactRawFormat},
		{"helm1", types.ArtifactHelmFormat},
	}

	for _, obj := range inPkgs {
		err := s.Packages().Create(ctx, &types.ArtifactPackage{OwnerID: 1, Name: obj.Name, Format: obj.Format})
		assert.NoError(t, err)
	}

	err := s.Versions().Create(ctx, &types.ArtifactVersion{PackageID: 1, Version: "0.0.1", ViewID: 1})
	require.NoError(t, err)

	res, err := s.FindPackages(ctx, 1, 1, &types.ArtifactFilter{})
	require.NoError(t, err)

	require.Equal(t, 1, len(res))
	require.Equal(t, "0.0.1", res[0].Version)

	// raw2 create version
	err = s.Versions().Create(ctx, &types.ArtifactVersion{PackageID: 2, Version: "1.0.1", ViewID: 1})
	require.NoError(t, err)

	res, err = s.FindPackages(ctx, 1, 1, &types.ArtifactFilter{})
	require.NoError(t, err)

	require.Equal(t, 2, len(res))
	require.Equal(t, "1.0.1", res[0].Version)
	require.Equal(t, "raw2", res[0].Name)

	// maven1 create version
	err = s.Versions().Create(ctx, &types.ArtifactVersion{PackageID: 3, Version: "1.2.1", ViewID: 1})
	require.NoError(t, err)

	// filter format
	f := types.ArtifactFilter{Format: "raw"}

	res, err = s.FindPackages(ctx, 1, 1, &f)
	require.NoError(t, err)

	require.Equal(t, 2, len(res))
	require.Equal(t, "1.0.1", res[0].Version)
	require.Equal(t, "raw2", res[0].Name)

	// add raw new version
	err = s.Versions().Create(ctx, &types.ArtifactVersion{PackageID: 2, Version: "1.0.2", ViewID: 1})
	require.NoError(t, err)

	res, err = s.FindPackages(ctx, 1, 1, &f)
	require.NoError(t, err)

	require.Equal(t, 2, len(res))
	require.Equal(t, "1.0.2", res[0].Version)
	require.Equal(t, "raw2", res[0].Name)
}
