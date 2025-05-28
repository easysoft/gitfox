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
	"github.com/stretchr/testify/require"
)

func TestArtifactBlob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	t.Parallel()
	tables := []any{new(types.ArtifactBlob)}
	_, gdb := dbtest.New(ctx, t, "artifacts_blob", tables...)
	ctl := &blobs{
		db: gdb,
	}

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, ctl *blobs)
	}{
		{"AddArtifactBlob", addBlob},
		{"GetArtifactBlob", getBlob},
		{"DelArtifactBlob", deleteBlob},
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

func addBlob(t *testing.T, ctx context.Context, ctl *blobs) {
	tests := []struct {
		StorageID int64
		Ref       string
		Size      int64
		wantErr   bool
		expectErr error
	}{
		{0, "f1aeb843414d377424dff9b54d7bcf3a6137b0e8", 2<<7 + 1, true, types.ErrArgsValueEmpty},
		{1, "", 2<<7 + 1, true, types.ErrArgsValueEmpty},
		{1, "f1aeb843414d377424dff9b54d7bcf3a6137b0e8", 2<<7 + 1, false, nil},
		{1, "f1aeb843414d377424dff9b54d7bcf3a6137b0e8", 2<<7 + 1, true, nil}, // unique key for storage_id, path
	}

	for id, test := range tests {
		obj := types.ArtifactBlob{StorageID: test.StorageID, Ref: test.Ref, Size: test.Size}
		err := ctl.Create(ctx, &obj)
		if test.wantErr {
			require.Error(t, err)
			if test.expectErr != nil {
				require.ErrorIs(t, err, test.expectErr)
			}
			continue
		}

		require.Equal(t, tests[id].StorageID, obj.StorageID)
		require.Equal(t, tests[id].Ref, obj.Ref)
		require.Equal(t, tests[id].Size, obj.Size)
	}

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &blobs{db: mockDb}

	t1 := tests[len(tests)-1]
	mock.ExpectQuery("INSERT ").WillReturnError(mysql.ErrInvalidConn)
	err := mockCtl.Create(ctx, &types.ArtifactBlob{StorageID: t1.StorageID, Ref: t1.Ref, Size: t1.Size})
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func getBlob(t *testing.T, ctx context.Context, ctl *blobs) {
	newObj := types.ArtifactBlob{StorageID: 1, Ref: "f1aeb843414d377424dff9b54d7bcf3a6137b0e8", Size: 2<<7 + 1}
	err := ctl.Create(ctx, &newObj)
	require.NoError(t, err)

	obj, err := ctl.GetById(ctx, newObj.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), obj.StorageID)
	require.Equal(t, "f1aeb843414d377424dff9b54d7bcf3a6137b0e8", obj.Ref)
	require.Equal(t, int64(257), obj.Size)

	_, err = ctl.GetById(ctx, newObj.ID+10086)
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	_, err = ctl.GetByRef(ctx, "", 1)
	require.ErrorIs(t, err, types.ErrArgsValueEmpty)

	_, err = ctl.GetByRef(ctx, "xx", 0)
	require.ErrorIs(t, err, types.ErrArgsValueEmpty)

	_, err = ctl.GetByRef(ctx, "xx", 1)
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	_, err = ctl.GetByRef(ctx, "f1aeb843414d377424dff9b54d7bcf3a6137b0e8", 1)
	require.NoError(t, err)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &blobs{db: mockDb}

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetById(ctx, newObj.ID)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetByRef(ctx, "f1aeb843414d377424dff9b54d7bcf3a6137b0e8", 1)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func deleteBlob(t *testing.T, ctx context.Context, ctl *blobs) {
	newObj := types.ArtifactBlob{StorageID: 1, Ref: "f1aeb843414d377424dff9b54d7bcf3a6137b0e8", Size: 2<<7 + 1}
	err := ctl.Create(ctx, &newObj)
	require.NoError(t, err)

	err = ctl.DeleteById(ctx, newObj.ID+10086)
	require.ErrorIs(t, err, types.ErrBlobNoItemDeleted)

	err = ctl.DeleteById(ctx, newObj.ID)
	require.NoError(t, err)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &blobs{db: mockDb}

	mock.ExpectQuery("delete").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.DeleteById(ctx, 1)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error(), err)
}
