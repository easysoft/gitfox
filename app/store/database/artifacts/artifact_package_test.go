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

func TestArtifactPackage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	t.Parallel()
	tables := []any{new(types.ArtifactPackage)}
	ctl := &packages{
		db: dbtest.NewDB(ctx, t, "artifacts_pkg", tables...),
	}

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *packages)
	}{
		{"AddArtifactPkg", addPkg},
		{"GetArtifactPkg", getPackage},
		{"DelArtifactPkg", delPackage},
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

func addPkg(t *testing.T, ctx context.Context, pkg *packages) {
	a1 := types.ArtifactPackage{Name: "a1", OwnerID: 1, Format: types.ArtifactRawFormat}
	err := pkg.Create(ctx, &a1)
	require.NoError(t, err)

	require.Equal(t, "a1", a1.Name)
	require.Equal(t, int64(1), a1.ID)
	require.Equal(t, int64(1), a1.OwnerID)

	a2 := types.ArtifactPackage{Name: "a2", OwnerID: 1, Format: types.ArtifactRawFormat}
	err = pkg.Create(ctx, &a2)
	require.NoError(t, err)

	require.Equal(t, "a2", a2.Name)
	require.Equal(t, "", a2.Namespace)

	a3 := types.ArtifactPackage{Name: "a3", OwnerID: 1, Format: types.ArtifactRawFormat, Namespace: "com.java"}
	err = pkg.Create(ctx, &a3)
	require.NoError(t, err)

	require.Equal(t, "a3", a3.Name)
	require.Equal(t, "com.java", a3.Namespace)

	err = pkg.Create(ctx, &a3)
	require.Error(t, err)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &packages{db: mockDb}

	mock.ExpectQuery("INSERT ").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.Create(ctx, &a3)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func getPackage(t *testing.T, ctx context.Context, pkg *packages) {
	a3 := types.ArtifactPackage{Name: "a3", OwnerID: 1, Format: types.ArtifactRawFormat, Namespace: "com.java"}
	err := pkg.Create(ctx, &a3)
	require.NoError(t, err)

	_, err = pkg.GetByID(ctx, a3.ID)
	require.NoError(t, err)

	_, err = pkg.GetByID(ctx, a3.ID+10086)
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	_, err = pkg.GetByName(ctx, "a3", "com.java", 1, types.ArtifactRawFormat)
	require.NoError(t, err)

	_, err = pkg.GetByName(ctx, "a3", "vv", 1, types.ArtifactRawFormat)
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	_, err = pkg.GetByName(ctx, "a3", "", 1, types.ArtifactRawFormat)
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	a32 := types.ArtifactPackage{Name: "a3", OwnerID: 1, Format: types.ArtifactRawFormat, Namespace: "com.java.gitfox"}
	err = pkg.Create(ctx, &a32)
	require.NoError(t, err)
	_, err = pkg.GetByName(ctx, "a3", "", 1, types.ArtifactRawFormat)
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &packages{db: mockDb}

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetByName(ctx, "a3", "com.java", 1, types.ArtifactRawFormat)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetByID(ctx, a3.ID)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func delPackage(t *testing.T, ctx context.Context, pkg *packages) {
	err := pkg.DeleteById(ctx, 10086)
	require.ErrorIs(t, err, types.ErrPkgNoItemDeleted)

	a3 := types.ArtifactPackage{Name: "a3", OwnerID: 1, Format: types.ArtifactRawFormat, Namespace: "com.java"}
	err = pkg.Create(ctx, &a3)
	require.NoError(t, err)

	err = pkg.DeleteById(ctx, a3.ID)
	require.NoError(t, err)

	_, err = pkg.GetByID(ctx, a3.ID)
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &packages{db: mockDb}

	mock.ExpectQuery("delete").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.DeleteById(ctx, a3.ID)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}
