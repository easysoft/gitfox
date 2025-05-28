// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"context"
	"testing"

	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database/dbtest"
	"github.com/easysoft/gitfox/types"

	"github.com/go-sql-driver/mysql"
	"github.com/guregu/null"
	"github.com/stretchr/testify/require"
)

func TestArtifactAsset(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	t.Parallel()
	tables := []any{new(types.ArtifactAsset), new(types.ArtifactMetaAsset)}
	ctl := &assets{
		db: dbtest.NewDB(ctx, t, "artifacts_asset", tables...),
	}

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *assets)
	}{
		{"AddArtifactAsset", addAsset},
		{"GetArtifactAsset", getAsset},
		{"UpdateArtifactAssetBlobId", updateAssetBlobId},
		//{"UpsertAssetById", upsertAssetById},
		{"DelArtifactAsset", deleteAsset},
		//{"AddMetaAsset", addMetaAsset},
		//{"GetMetaAsset", getMetaAsset},
		//{"UpdateMetaAsset", updateMetaAsset},
		//{"DelMetaAsset", deleteMetaAsset},
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

func addAsset(t *testing.T, ctx context.Context, ctl *assets) {
	var verId int64 = 1
	var blobId int64 = 1
	var assetFile = "/redis/clients/jedis/4.4.2/jedis-4.4.2-javadoc.jar"
	tests := []struct {
		Path      string
		VersionId int64
		Kind      types.AssetKind
		BlobID    int64
		wantErr   bool
		expectErr error
	}{
		{assetFile, verId, types.AssetKind(""), 0, true, nil},
		{assetFile, verId, types.AssetKind("xxx"), 0, true, nil},
		{assetFile, verId, types.AssetKindMain, 0, false, nil},
		{assetFile, verId, types.AssetKindMain, blobId, false, nil},
		{assetFile + ".asc", verId, types.AssetKindSub, blobId + 1, false, nil},
		{assetFile + ".md5", verId, types.AssetKindSub, blobId + 2, false, nil},
	}

	for id, test := range tests {
		obj := types.ArtifactAsset{
			VersionID: null.IntFrom(test.VersionId),
			Path:      test.Path,
			Kind:      test.Kind,
			BlobID:    test.BlobID,
		}
		err := ctl.Create(ctx, &obj)
		if test.wantErr {
			require.Error(t, err, "loop number: %d", id)
			if test.expectErr != nil {
				require.ErrorIs(t, err, test.expectErr, "loop number: ", id)
			}
			continue
		}

		require.Equal(t, tests[id].Path, obj.Path)
		require.Equal(t, tests[id].VersionId, obj.VersionID.Int64)
		require.Equal(t, tests[id].BlobID, obj.BlobID)
		require.Equal(t, tests[id].Kind, obj.Kind)
	}

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &assets{mockDb}

	t1 := tests[len(tests)-1]
	mock.ExpectQuery("INSERT ").WillReturnError(mysql.ErrInvalidConn)
	err := mockCtl.Create(ctx, &types.ArtifactAsset{
		VersionID: null.IntFrom(t1.VersionId),
		Path:      t1.Path,
		Kind:      t1.Kind,
		BlobID:    t1.BlobID,
	})
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func addAssetTestData(ctx context.Context, ctl *assets) []int64 {
	var verId = null.IntFrom(1)
	var blobId int64 = 1
	var data = []types.ArtifactAsset{
		{VersionID: verId, Path: "/redis/clients/jedis/4.4.2/jedis-4.4.2.jar", Kind: types.AssetKindMain, BlobID: blobId},
		{VersionID: verId, Path: "/redis/clients/jedis/4.4.2/jedis-4.4.2.jar.asc", Kind: types.AssetKindSub, BlobID: blobId + 1},
		{VersionID: verId, Path: "/redis/clients/jedis/4.4.2/jedis-4.4.2.jar.md5", Kind: types.AssetKindSub, BlobID: blobId + 2},
		{VersionID: verId, Path: "/redis/clients/jedis/5.0.0/jedis-5.0.0.jar", Kind: types.AssetKindMain, BlobID: blobId + 3},
		{VersionID: verId, Path: "/redis/clients/jedis/5.0.0/jedis-5.0.0.jar.asc", Kind: types.AssetKindSub, BlobID: blobId + 4},
		{VersionID: verId, Path: "/redis/clients/jedis/5.0.0/jedis-5.0.0.jar.md5", Kind: types.AssetKindSub, BlobID: blobId + 5},
	}

	pks := make([]int64, 0)
	for _, item := range data {
		ctl.db.WithContext(ctx).Create(&item)
		pks = append(pks, item.ID)
	}
	return pks
}

func getAsset(t *testing.T, ctx context.Context, ctl *assets) {
	pks := addAssetTestData(ctx, ctl)

	_, err := ctl.GetById(ctx, pks[0]+10086)
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	obj, err := ctl.GetById(ctx, pks[0])
	require.NoError(t, err)
	require.Equal(t, "/redis/clients/jedis/4.4.2/jedis-4.4.2.jar", obj.Path)
	require.Equal(t, types.AssetKindMain, obj.Kind)

	obj, err = ctl.GetById(ctx, pks[4])
	require.NoError(t, err)
	require.Equal(t, "/redis/clients/jedis/5.0.0/jedis-5.0.0.jar.asc", obj.Path)
	require.Equal(t, types.AssetKindSub, obj.Kind)

	obj, err = ctl.GetVersionAsset(ctx, "/redis/clients/jedis/4.4.2/jedis-4.4.2.jar", 1)
	require.NoError(t, err)
	require.Equal(t, types.AssetKindMain, obj.Kind)

	invalidTests := []struct {
		Path      string
		VersionId int64
		Err       error
	}{
		{"/redis/clients/jedis/4.4.2/jedis-4.4.2.jar", 2, gitfox_store.ErrResourceNotFound},
		{"/redis/clients/jedis/4.4.3/jedis-4.4.3.jar", 1, gitfox_store.ErrResourceNotFound},
		{"", 1, types.ErrArgsValueEmpty},
		{"/redis/clients/jedis/4.4.2/jedis-4.4.2.jar", 0, types.ErrArgsValueEmpty},
	}

	for id, test := range invalidTests {
		_, err = ctl.GetVersionAsset(ctx, test.Path, test.VersionId)
		require.ErrorIs(t, err, test.Err, "loop index: %d, err: %+v", id, err)
	}

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &assets{mockDb}

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetById(ctx, pks[4])
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetVersionAsset(ctx, "/redis/clients/jedis/4.4.2/jedis-4.4.2.jar", 1)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func updateAssetBlobId(t *testing.T, ctx context.Context, ctl *assets) {
	pks := addAssetTestData(ctx, ctl)

	var updateBlobId int64 = 10

	testPk := pks[0]
	testUpdateBlobIds := []struct {
		pk        int64
		blobId    int64
		wantErr   bool
		expectErr error
	}{
		{0, updateBlobId, true, types.ErrArgsValueEmpty},
		{testInvalidPk, updateBlobId, true, gitfox_store.ErrResourceNotFound},
		{testPk, updateBlobId, false, nil},
	}

	for id, test := range testUpdateBlobIds {
		err := ctl.UpdateBlobId(ctx, test.pk, test.blobId)
		if test.wantErr {
			require.Error(t, err)
			if test.expectErr != nil {
				require.ErrorIs(t, err, test.expectErr, "loop number: ", id)
			}
			continue
		}

		obj, err := ctl.GetById(ctx, test.pk)
		require.NoError(t, err)
		require.Equal(t, testUpdateBlobIds[id].blobId, obj.BlobID)
	}

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &assets{mockDb}

	mock.ExpectQuery("SELECT ").WillReturnError(mysql.ErrInvalidConn)
	err := mockCtl.UpdateBlobId(ctx, pks[1], updateBlobId)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func upsertAssetById(t *testing.T, ctx context.Context, ctl *assets) {
	pks := addAssetTestData(ctx, ctl)

	var updateBlobId int64 = 10

	lastPk := pks[len(pks)-1]
	lastObj, err := ctl.GetById(ctx, lastPk)
	require.NoError(t, err)

	testUpserts := []struct {
		object      *types.ArtifactAsset
		wantErr     bool
		expectErr   error
		wantCreated bool
		expectPk    int64
		equalFunc   func(t *testing.T, req, updated *types.ArtifactAsset)
	}{
		// Nothing changed
		{lastObj, false, nil, false, lastPk, nil},

		// Change lastObj with blob_id, omit kind (NOT NULL)
		// insert will success in sqlite, failed in mysql/postgres, this case will be commented for default sqlite db,
		//{&model.ArtifactAsset{ID: lastPk, VersionID: lastObj.VersionID, Path: lastObj.Path, BlobID: updateBlobId},
		//	false, nil, false, lastPk,
		//	func(t *testing.T, req, updated *model.ArtifactAsset) {
		//		require.Equal(t, req.BlobID, updated.BlobID)
		//		require.Equal(t, lastObj.Kind, updated.Kind, "%+v", updated)
		//	},
		//},

		// Create new object same as lastObj, should not be created by conflict index path,version_id
		{&types.ArtifactAsset{VersionID: lastObj.VersionID, Path: lastObj.Path, Kind: types.AssetKindMain, BlobID: updateBlobId},
			false, nil, false, 0, nil,
		},

		// Create new object with different path
		{&types.ArtifactAsset{VersionID: lastObj.VersionID, Path: lastObj.Path + "v1", Kind: types.AssetKindMain, BlobID: updateBlobId},
			false, nil, true, lastPk + 1, nil,
		},
	}

	for id, test := range testUpserts {
		t.Logf("loop index: %d", id)
		e := ctl.Update(ctx, test.object)
		if test.wantErr {
			require.Error(t, e)
			if test.expectErr != nil {
				require.ErrorIs(t, e, test.expectErr)
			}
			continue
		}

		if test.wantCreated {
			require.Equal(t, testUpserts[id].expectPk, test.object.ID)
		}

		if test.equalFunc != nil {
			require.Equal(t, test.expectPk, test.object.ID)
			upObj, err := ctl.GetById(ctx, test.object.ID)
			require.NoError(t, err)
			test.equalFunc(t, test.object, upObj)
		}
	}

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &assets{mockDb}

	mock.ExpectQuery("UPDATE").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.Update(ctx, &types.ArtifactAsset{})
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func deleteAsset(t *testing.T, ctx context.Context, ctl *assets) {
	pks := addAssetTestData(ctx, ctl)

	for id, pk := range pks {
		err := ctl.DeleteById(ctx, pk)
		require.NoError(t, err, "loop index: %d, pk: %d", id, pk)
	}

	err := ctl.DeleteById(ctx, pks[0]+testInvalidPk)
	require.ErrorIs(t, err, types.ErrAssetNoItemDeleted)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &assets{mockDb}

	mock.ExpectQuery("delete").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.DeleteById(ctx, 1)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error(), err)
}
