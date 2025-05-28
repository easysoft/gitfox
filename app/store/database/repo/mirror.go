// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

const (
	mirrorTable = "mirrors"
)

type mirrors struct {
	ID             int64  `gorm:"column:id;primaryKey"`
	RepoID         int64  `gorm:"column:repo_id"`
	SpaceID        int64  `gorm:"column:space_id"`
	SyncInterval   int64  `gorm:"column:sync_interval"`
	EnablePrune    bool   `gorm:"column:enable_prune"`
	UpdatedUnix    int64  `gorm:"column:updated_unix"`
	NextUpdateUnix int64  `gorm:"column:next_update_unix"`
	LfsEnabled     bool   `gorm:"column:lfs_enabled"`
	RemoteAddress  string `gorm:"column:remote_address"`
}

func (s *OrmStore) mapToMirror(
	ctx context.Context,
	in *mirrors,
) (*types.RepositoryMirror, error) {
	res := &types.RepositoryMirror{
		SyncInterval:   in.SyncInterval,
		EnablePrune:    in.EnablePrune,
		UpdatedUnix:    in.UpdatedUnix,
		NextUpdateUnix: in.NextUpdateUnix,
		LfsEnabled:     in.LfsEnabled,
		RemoteAddress:  in.RemoteAddress,
		RepoID:         in.RepoID,
		SpaceID:        in.SpaceID,
	}
	return res, nil
}

func (s *OrmStore) CreateOrUpdateMirror(ctx context.Context, opts *types.RepositoryMirror) error {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(mirrorTable).Where("repo_id = ? ", opts.RepoID)
	dst := new(mirrors)
	err := stmt.First(dst).Error
	now := time.Now().Unix()
	dst.UpdatedUnix = now
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			next := now + opts.SyncInterval*60
			dst.NextUpdateUnix = next
			dst.RepoID = opts.RepoID
			dst.SpaceID = opts.SpaceID
			dst.SyncInterval = opts.SyncInterval
			dst.EnablePrune = opts.EnablePrune
			dst.LfsEnabled = opts.LfsEnabled
			dst.RemoteAddress = opts.RemoteAddress
			return dbtx.GetOrmAccessor(ctx, s.db).Table(mirrorTable).Create(dst).Error
		}
		return fmt.Errorf("failed to create mirror: %w", err)
	}
	dst.NextUpdateUnix = now + dst.SyncInterval*60
	return dbtx.GetOrmAccessor(ctx, s.db).Table(mirrorTable).Save(dst).Error
}

func (s *OrmStore) ListAllMirrorRepo(ctx context.Context) ([]*types.RepositoryMirror, error) {
	var dst []*types.RepositoryMirror
	err := dbtx.GetOrmAccessor(ctx, s.db).Table(mirrorTable).
		Select("mirrors.*, repositories.repo_git_uid").
		Joins("LEFT JOIN repositories ON mirrors.repo_id = repositories.repo_id").
		Where("repositories.repo_deleted IS NULL and repositories.repo_mirror = ?", true).
		Scan(&dst).Error
	return dst, err
}

func (s *OrmStore) GetMirror(ctx context.Context, repoID int64) (*types.RepositoryMirror, error) {
	var dst mirrors
	err := dbtx.GetOrmAccessor(ctx, s.db).Table(mirrorTable).Where("repo_id = ?", repoID).First(&dst).Error
	if err != nil {
		return nil, err
	}
	return s.mapToMirror(ctx, &dst)
}

func (s *OrmStore) UpdateMirror(ctx context.Context, opts *types.RepositoryMirror) error {
	// 暂时只能修改同步周期
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(mirrorTable).Where("repo_id = ? ", opts.RepoID)
	dst := new(mirrors)
	err := stmt.First(dst).Error
	if err != nil || dst == nil {
		return fmt.Errorf("failed to update mirror: %w", err)
	}
	dst.SyncInterval = opts.SyncInterval
	dst.NextUpdateUnix = dst.UpdatedUnix + dst.SyncInterval*60
	return dbtx.GetOrmAccessor(ctx, s.db).Table(mirrorTable).Save(dst).Error
}
