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

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func testMetaArtifactAsset(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	t.Parallel()
	tables := []any{new(types.ArtifactAsset), new(types.ArtifactMetaAsset)}
	ctl := &metaAssets{
		db: dbtest.NewDB(ctx, t, "artifacts_meta_asset", tables...),
	}

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *metaAssets)
	}{
		{"AddMetaAsset", addMetaAsset},
		{"GetMetaAsset", getMetaAsset},
		{"UpdateMetaAsset", updateMetaAsset},
		{"DelMetaAsset", deleteMetaAsset},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := dbtest.ClearTables(t, ctl.db, tables...)
				require.NoError(t, err)
			})
			tc.test(t, ctx, ctl)
		})
		if t.Failed() {
			break
		}
	}
}

func addMetaAsset(t *testing.T, ctx context.Context, ctl *metaAssets) {
	var ownerId int64 = 1
	var blobId int64 = 1
	var viewId int64 = 1
	var metaAssetFile = "/redis/clients/jedis/maven-metadata.xml"
	tests := []struct {
		Path      string
		OwnerId   int64
		Format    types.ArtifactFormat
		Kind      types.AssetKind
		BlobID    int64
		wantErr   bool
		expectErr error
	}{
		// empty path, ownerId, format
		{"", ownerId, types.ArtifactMavenFormat, types.AssetKind(""), 0, true, types.ErrArgsValueEmpty},
		{metaAssetFile, 0, types.ArtifactMavenFormat, types.AssetKind(""), 0, true, types.ErrArgsValueEmpty},
		{metaAssetFile, ownerId, types.ArtifactFormat(""), types.AssetKind(""), 0, true, types.ErrArgsValueEmpty},

		// unidentified kind
		{metaAssetFile, ownerId, types.ArtifactMavenFormat, types.AssetKind(""), 0, true, nil},
		{metaAssetFile, ownerId, types.ArtifactMavenFormat, types.AssetKind("xxx"), 0, true, nil},

		// blobId allow empty
		{metaAssetFile, ownerId, types.ArtifactMavenFormat, types.AssetKindMain, 0, false, nil},

		// success create
		{metaAssetFile, ownerId, types.ArtifactMavenFormat, types.AssetKindMain, blobId, false, nil},
		{metaAssetFile + ".md5", ownerId, types.ArtifactMavenFormat, types.AssetKindMain, blobId + 1, false, nil},
		{metaAssetFile + ".sha1", ownerId, types.ArtifactMavenFormat, types.AssetKindMain, blobId + 2, false, nil},
	}

	for id, test := range tests {
		obj := types.ArtifactMetaAsset{
			OwnerID: test.OwnerId,
			Format:  test.Format,
			Path:    test.Path,
			ViewID:  viewId,
			Kind:    test.Kind,
			BlobID:  test.BlobID,
		}
		err := ctl.Create(ctx, &obj)
		if test.wantErr {
			require.Error(t, err)
			if test.expectErr != nil {
				require.ErrorIs(t, err, test.expectErr, "loop is %d", id)
			}
			continue
		}

		require.Equal(t, tests[id].Path, obj.Path)
		require.Equal(t, tests[id].OwnerId, obj.OwnerID)
		require.Equal(t, tests[id].BlobID, obj.BlobID)
		require.Equal(t, tests[id].Kind, obj.Kind)
		require.Equal(t, tests[id].Format, obj.Format)
	}

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &metaAssets{mockDb}

	t1 := tests[len(tests)-1]
	mock.ExpectQuery("INSERT ").WillReturnError(mysql.ErrInvalidConn)
	err := mockCtl.Create(ctx, &types.ArtifactMetaAsset{
		OwnerID: t1.OwnerId,
		Format:  t1.Format,
		Path:    t1.Path,
		ViewID:  viewId,
		Kind:    t1.Kind,
		BlobID:  t1.BlobID,
	})
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func addMetaAssetTestData(ctx context.Context, ctl *metaAssets) []int64 {
	var viewId int64 = 1
	var blobId int64 = 1
	var data = []types.ArtifactMetaAsset{
		{ViewID: viewId, Path: "/redis/clients/jedis/maven-metadata.xml", Format: types.ArtifactMavenFormat, Kind: types.AssetKindMain, BlobID: blobId},
		{ViewID: viewId, Path: "/redis/clients/jedis/maven-metadata.xml.md5", Format: types.ArtifactMavenFormat, Kind: types.AssetKindSub, BlobID: blobId + 1},
		{ViewID: viewId, Path: "/redis/clients/jedis/maven-metadata.xml.sha1", Format: types.ArtifactMavenFormat, Kind: types.AssetKindSub, BlobID: blobId + 2},
		{ViewID: viewId, Path: "/redis/clients/jedis/maven-metadata.xml.sha256", Format: types.ArtifactMavenFormat, Kind: types.AssetKindSub, BlobID: blobId + 3},
		{ViewID: viewId, Path: "/redis/clients/jedis/maven-metadata.xml.sha512", Format: types.ArtifactMavenFormat, Kind: types.AssetKindSub, BlobID: blobId + 4},
		{ViewID: viewId, Path: "/archetype-catalog.xml", Format: types.ArtifactMavenFormat, Kind: types.AssetKindMain, BlobID: blobId + 5},
		{ViewID: viewId, Path: "/archetype-catalog.xml.md5", Format: types.ArtifactMavenFormat, Kind: types.AssetKindSub, BlobID: blobId + 6},
		{ViewID: viewId, Path: "/archetype-catalog.xml.sha1", Format: types.ArtifactMavenFormat, Kind: types.AssetKindSub, BlobID: blobId + 7},
	}

	pks := make([]int64, 0)
	for _, item := range data {
		ctl.db.WithContext(ctx).Create(&item)
		pks = append(pks, item.ID)
	}
	return pks
}

func getMetaAsset(t *testing.T, ctx context.Context, ctl *metaAssets) {
	pks := addMetaAssetTestData(ctx, ctl)

	_, err := ctl.GetById(ctx, pks[0]+10086)
	require.ErrorIs(t, err, types.ErrMetaAssetNotFound)

	obj, err := ctl.GetById(ctx, pks[0])
	require.NoError(t, err)
	require.Equal(t, "/redis/clients/jedis/maven-metadata.xml", obj.Path)
	require.Equal(t, types.AssetKindMain, obj.Kind)

	obj, err = ctl.GetById(ctx, pks[4])
	require.NoError(t, err)
	require.Equal(t, "/redis/clients/jedis/maven-metadata.xml.sha512", obj.Path)
	require.Equal(t, types.AssetKindSub, obj.Kind)

	obj, err = ctl.GetByPath(ctx, "/redis/clients/jedis/maven-metadata.xml", 1, types.ArtifactMavenFormat)
	require.NoError(t, err)
	require.Equal(t, types.AssetKindMain, obj.Kind)

	invalidTests := []struct {
		Path   string
		ViewId int64
		Format types.ArtifactFormat
		Err    error
	}{
		{"/redis/clients/jedis/maven-metadata.xml", 2, types.ArtifactMavenFormat, types.ErrMetaAssetNotFound},
		{"/redis/clients/jedis/maven-metadata.html", 1, types.ArtifactMavenFormat, types.ErrMetaAssetNotFound},
		{"/redis/clients/jedis/maven-metadata.xml", 1, types.ArtifactContainerFormat, types.ErrMetaAssetNotFound},
		{"", 1, types.ArtifactMavenFormat, types.ErrArgsValueEmpty},
		{"/redis/clients/jedis/maven-metadata.xml", 0, types.ArtifactMavenFormat, types.ErrArgsValueEmpty},
		{"/redis/clients/jedis/maven-metadata.xml", 2, types.ArtifactFormat(""), types.ErrArgsValueEmpty},
	}

	for id, test := range invalidTests {
		_, err = ctl.GetByPath(ctx, test.Path, test.ViewId, test.Format)
		require.ErrorIs(t, err, test.Err, "loop index: %d, err: %+v", id, err)
	}

	// test GetMetaAssetOption Apply()
	_, err = ctl.GetByPath(ctx, "/redis/clients/jedis/maven-metadata.xml", -1, types.ArtifactMavenFormat)
	require.ErrorIs(t, err, types.ErrMetaAssetNotFound)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &metaAssets{mockDb}

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetById(ctx, pks[4])
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetByPath(ctx, "/redis/clients/jedis/maven-metadata.xml", 1, types.ArtifactMavenFormat)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func updateMetaAsset(t *testing.T, ctx context.Context, ctl *metaAssets) {
	pks := addMetaAssetTestData(ctx, ctl)

	var updateBlobId int64 = 10

	err := ctl.UpdateBlobId(ctx, pks[0], updateBlobId)
	require.NoError(t, err)

	obj, err := ctl.GetById(ctx, pks[0])
	require.NoError(t, err)
	require.Equal(t, updateBlobId, obj.BlobID)

	// invalid test
	err = ctl.UpdateBlobId(ctx, pks[1], 0)
	require.ErrorIs(t, err, types.ErrArgsValueEmpty)

	err = ctl.UpdateBlobId(ctx, pks[1]+10086, updateBlobId+1)
	require.ErrorIs(t, err, types.ErrMetaAssetNotFound)

	obj.BlobID = updateBlobId + 3
	err = ctl.Update(ctx, obj)
	require.NoError(t, err)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &metaAssets{mockDb}

	mock.ExpectQuery("UPDATE").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.UpdateBlobId(ctx, pks[1], updateBlobId+2)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())

	obj, _ = ctl.GetById(ctx, pks[0])
	obj.BlobID = updateBlobId + 5
	mock.ExpectQuery("INSERT INTO").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.Update(ctx, obj)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func deleteMetaAsset(t *testing.T, ctx context.Context, ctl *metaAssets) {
	pks := addMetaAssetTestData(ctx, ctl)

	for id, pk := range pks {
		err := ctl.DeleteById(ctx, pk)
		require.NoError(t, err, "loop index: %d, pk: %d", id, pk)
	}

	err := ctl.DeleteById(ctx, pks[0]+10086)
	require.ErrorIs(t, err, types.ErrMetaAssetNoItemDeleted)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &metaAssets{mockDb}

	mock.ExpectQuery("delete").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.DeleteById(ctx, 1)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error(), err)
}
