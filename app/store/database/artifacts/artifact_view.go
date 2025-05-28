// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

var _ store.ArtifactViewInterface = (*views)(nil)

type views struct {
	db *gorm.DB
}

func (c *views) Create(ctx context.Context, newObj *types.ArtifactView) error {
	if err := dbtx.GetOrmAccessor(ctx, c.db).Model(new(types.ArtifactView)).Create(newObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "creating artifact view failed")
	}
	return nil
}

func (c *views) GetDefault(ctx context.Context, spaceId int64) (*types.ArtifactView, error) {
	var view types.ArtifactView
	q := types.ArtifactView{SpaceID: spaceId, Default: true}
	if err := dbtx.GetOrmAccessor(ctx, c.db).Where(q).First(&view).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec view query failed")
	}

	return &view, nil
}

func (c *views) GetByName(ctx context.Context, spaceId int64, name string) (*types.ArtifactView, error) {
	var view types.ArtifactView
	q := types.ArtifactView{SpaceID: spaceId, Name: name}
	if err := dbtx.GetOrmAccessor(ctx, c.db).Where(q).First(&view).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec view query failed")
	}

	return &view, nil
}
