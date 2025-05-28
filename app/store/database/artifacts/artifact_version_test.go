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

func TestArtifactVersion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	t.Parallel()
	tables := []any{new(types.ArtifactVersion)}
	ctl := &versions{
		db: dbtest.NewDB(ctx, t, "artifacts_pkg_ver", tables...),
	}

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *versions)
	}{
		{"AddArtifactPkgVer", addPackageVer},
		{"GetArtifactPkgVer", getPackageVer},
		{"DelArtifactPkgVer", delPackageVer},
		{"SearchArtifactVer", searchVersion},
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

func addPackageVer(t *testing.T, ctx context.Context, ctl *versions) {
	obj100 := types.ArtifactVersion{PackageID: 1, Version: "1.0.0"}
	err := ctl.Create(ctx, &obj100)
	require.NoError(t, err)
	require.Equal(t, "1.0.0", obj100.Version)
	require.Equal(t, int64(1), obj100.PackageID)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &versions{db: mockDb}

	mock.ExpectQuery("INSERT ").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.Create(ctx, &obj100)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func addVersionTestData(ctx context.Context, ctl *versions) []int64 {
	var data = []types.ArtifactVersion{
		{PackageID: 1, Version: "1.0.0", ViewID: 1},
		{PackageID: 1, Version: "1.0.1", ViewID: 1},
		{PackageID: 1, Version: "1.0.2", ViewID: 1},
		{PackageID: 1, Version: "1.0.1", ViewID: 2}, // promote No.2
		{PackageID: 2, Version: "1.0.0.beta1", ViewID: 1},
		{PackageID: 2, Version: "1.0.0.rc1", ViewID: 1},
		{PackageID: 2, Version: "1.0.0", ViewID: 1},
		{PackageID: 2, Version: "1.0.0", ViewID: 2}, // promote previous 1.0.0
	}

	pks := make([]int64, 0)
	for _, item := range data {
		ctl.db.WithContext(ctx).Create(&item)
		pks = append(pks, item.ID)
	}
	return pks
}

func getPackageVer(t *testing.T, ctx context.Context, ctl *versions) {
	testData := [][3]interface{}{
		{int64(1), "1.0.0", int64(1)},
		{int64(1), "1.0.1", int64(1)},
		{int64(2), "1.3.1", int64(1)},
		{int64(3), "1.0.0-beta1", int64(1)},
	}

	pks := make([]int64, 0)

	for _, d := range testData {
		obj := types.ArtifactVersion{PackageID: d[0].(int64), ViewID: d[2].(int64), Version: d[1].(string)}
		err := ctl.Create(ctx, &obj)
		require.NoError(t, err)

		pks = append(pks, obj.ID)
	}

	for n, d := range testData {
		pk := pks[n]
		objById, err := ctl.GetByID(ctx, pk)
		require.NoError(t, err)
		require.Equal(t, pk, objById.ID)
		require.Equal(t, d[0].(int64), objById.PackageID)
		require.Equal(t, d[1].(string), objById.Version)

		objByName, err := ctl.GetByVersion(ctx, d[0].(int64), d[2].(int64), d[1].(string))
		require.NoError(t, err)
		require.Equal(t, pk, objByName.ID)
	}

	// invalid test
	_, err := ctl.GetByID(ctx, pks[len(pks)-1]+1)
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	_, err = ctl.GetByVersion(ctx, 1, 1, "1.3.1")
	require.ErrorIs(t, err, gitfox_store.ErrResourceNotFound)

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &versions{db: mockDb}

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetByID(ctx, pks[len(pks)-1]+1)
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())

	mock.ExpectQuery("SELECT").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.GetByVersion(ctx, 1, 1, "1.3.1")
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func delPackageVer(t *testing.T, ctx context.Context, ctl *versions) {
	err := ctl.DeleteById(ctx, 10086)
	require.ErrorIs(t, err, types.ErrPkgVersionNoItemDeleted)

	testData := [][2]interface{}{
		{int64(1), "1.0.0"},
		{int64(1), "1.0.1"},
		{int64(2), "1.3.1"},
		{int64(3), "1.0.0-beta1"},
	}

	pks := make([]int64, 0)

	for _, d := range testData {
		obj := types.ArtifactVersion{PackageID: d[0].(int64), Version: d[1].(string)}
		e := ctl.Create(ctx, &obj)
		require.NoError(t, e)

		pks = append(pks, obj.ID)
	}

	err = ctl.DeleteById(ctx, pks[0])
	t.Logf("pks[0]: %d", pks[0])
	require.NoError(t, err)

	delNum, err := ctl.DeleteByIds(ctx, pks[1:3]...)
	require.NoError(t, err)
	require.Equal(t, int64(2), delNum, pks[1:3])

	// mocker sql exec error
	mockDb, mock := dbtest.NewMockDB()
	mockCtl := &versions{db: mockDb}

	mock.ExpectQuery("delete").WillReturnError(mysql.ErrInvalidConn)
	_, err = mockCtl.DeleteByIds(ctx, pks[len(pks)-1])
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())

	mock.ExpectQuery("delete").WillReturnError(mysql.ErrInvalidConn)
	err = mockCtl.DeleteById(ctx, pks[len(pks)-1])
	require.ErrorContains(t, err, mysql.ErrInvalidConn.Error())
}

func searchVersion(t *testing.T, ctx context.Context, ctl *versions) {
	_ = addVersionTestData(ctx, ctl)

	tests := []struct {
		q            types.SearchVersionOption
		expectCount  int
		firstVersion string
	}{
		{types.SearchVersionOption{PackageId: 1}, 4, "1.0.0"},
		{types.SearchVersionOption{PackageId: 1, ViewId: 1}, 3, "1.0.0"},
		{types.SearchVersionOption{PackageId: 1, ViewId: 2}, 1, "1.0.1"},
		{types.SearchVersionOption{PackageId: 2}, 4, "1.0.0.beta1"},
		{types.SearchVersionOption{PackageId: 2, ViewId: 1}, 3, "1.0.0.beta1"},
		{types.SearchVersionOption{PackageId: 2, ViewId: 2}, 1, "1.0.0"},
		{types.SearchVersionOption{ViewId: 1}, 6, "1.0.0"},
		{types.SearchVersionOption{ViewId: 2}, 2, "1.0.1"},
	}

	for _, test := range tests {
		items, err := ctl.Find(ctx, test.q)
		require.NoError(t, err)
		require.Equal(t, test.expectCount, len(items))
		require.Equalf(t, test.firstVersion, items[0].Version, "items: %+v", items)
	}
}
