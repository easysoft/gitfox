// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package principal

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.PrincipalInfoView = (*InfoView)(nil)

// NewPrincipalOrmInfoView returns a new InfoView.
// It's used by the principal info cache.
func NewPrincipalOrmInfoView(db *gorm.DB) *InfoView {
	return &InfoView{
		db: db,
	}
}

type InfoView struct {
	db *gorm.DB
}

type Info struct {
	ID          int64              `gorm:"column:principal_id;primaryKey"`
	UID         string             `gorm:"column:principal_uid"`
	DisplayName string             `gorm:"column:principal_display_name"`
	Email       string             `gorm:"column:principal_email"`
	Type        enum.PrincipalType `gorm:"column:principal_type"`
	Created     int64              `gorm:"column:principal_created"`
	Updated     int64              `gorm:"column:principal_updated"`
}

// Find returns a single principal info object by id from the `principals` database table.
func (s *InfoView) Find(ctx context.Context, id int64) (*types.PrincipalInfo, error) {
	var info Info

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).First(&info, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to find principal info")
	}

	return mapToPrincipalInfo(&info), nil
}

// FindMany returns a several principal info objects by id from the `principals` database table.
func (s *InfoView) FindMany(ctx context.Context, ids []int64) ([]*types.PrincipalInfo, error) {
	var dst []*Info
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(principalTable).Find(&dst, ids).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to query find many principal info")
	}

	return mapToPrincipalInfos(dst), nil
}

func MapToInfo(p *Info) types.PrincipalInfo {
	return *mapToPrincipalInfo(p)
}

func mapToPrincipalInfo(p *Info) *types.PrincipalInfo {
	return &types.PrincipalInfo{
		ID:          p.ID,
		UID:         p.UID,
		DisplayName: p.DisplayName,
		Email:       p.Email,
		Type:        p.Type,
		Created:     p.Created,
		Updated:     p.Updated,
	}
}

func mapToPrincipalInfos(pList []*Info) []*types.PrincipalInfo {
	res := make([]*types.PrincipalInfo, len(pList))
	for i := range pList {
		res[i] = mapToPrincipalInfo(pList[i])
	}
	return res
}
