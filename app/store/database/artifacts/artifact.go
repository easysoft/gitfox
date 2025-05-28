// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"gorm.io/gorm"
)

var _ store.ArtifactStore = (*Store)(nil)

type Store struct {
	db         *gorm.DB
	repos      *repositories
	views      *views
	packages   *packages
	versions   *versions
	assets     *assets
	metaAssets *metaAssets
	blobs      *blobs
	nodes      *treeNode
}

func NewStore(db *gorm.DB) *Store {
	return &Store{
		db:         db,
		repos:      &repositories{db: db},
		views:      &views{db: db},
		packages:   &packages{db: db},
		versions:   &versions{db: db},
		assets:     &assets{db: db},
		metaAssets: &metaAssets{db: db},
		blobs:      &blobs{db: db},
		nodes:      &treeNode{db: db},
	}
}

func (s *Store) Repositories() store.ArtifactRepositoryInterface {
	return s.repos
}

func (s *Store) Views() store.ArtifactViewInterface {
	return s.views
}

func (s *Store) Packages() store.ArtifactPackageInterface {
	return s.packages
}

func (s *Store) Versions() store.ArtifactVersionInterface {
	return s.versions
}

func (s *Store) Assets() store.ArtifactAssetInterface {
	return s.assets
}

func (s *Store) MetaAssets() store.ArtifactMetaAssetInterface {
	return s.metaAssets
}

func (s *Store) Blobs() store.ArtifactBlobInterface {
	return s.blobs
}

func (s *Store) Nodes() store.ArtifactTreeNodeInterface {
	return s.nodes
}

func (s *Store) FindPackages(ctx context.Context, spaceId, viewId int64, filter *types.ArtifactFilter) ([]*types.ArtifactListItem, error) {
	result := make([]*types.ArtifactListItem, 0)

	q := types.ArtifactPackage{OwnerID: spaceId}

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Model(&types.ArtifactPackage{}).
		Select(`artifact_packages.package_format, artifact_packages.package_name, artifact_packages.package_namespace,
			artifact_versions.version, artifact_versions.version_updated`).
		Joins(`JOIN (
			SELECT version_package_id, MAX(version_updated) AS latest_updated
				FROM artifact_versions
				WHERE version_view_id = ?
				GROUP BY version_package_id) AS latest_versions
			ON artifact_packages.package_id = latest_versions.version_package_id
				AND artifact_packages.package_deleted = 0`, viewId).
		Joins(`JOIN artifact_versions
			ON artifact_versions.version_package_id = latest_versions.version_package_id
				AND artifact_versions.version_updated = latest_versions.latest_updated
				AND artifact_versions.version_deleted = 0`)

	if filter != nil {
		if filter.Format != "" {
			q.Format = types.ArtifactFormat(filter.Format)
		}
		if filter.Query != "" {
			stmt = stmt.Where("package_name LIKE ?", fmt.Sprintf("%%%s%%", filter.Query))
		}
	}

	stmt.Where(&q).Order("latest_versions.latest_updated desc").
		Limit(int(database.Limit(filter.Size))).
		Offset(int(database.Offset(filter.Page, filter.Size))).
		Find(&result)

	for id, p := range result {
		result[id].DisplayName = p.Name
		if p.Namespace != "" {
			result[id].DisplayName = p.Namespace + ":" + p.Name
		}
	}

	return result, nil
}

func (s *Store) GetVersion(ctx context.Context, spaceId, viewId int64,
	packageName, groupName, versionName string, format types.ArtifactFormat,
) (*types.ArtifactVersion, error) {
	dbPkg, err := s.packages.GetByName(ctx, packageName, groupName, spaceId, format)
	if err != nil {
		return nil, err
	}
	dbVer, err := s.versions.GetByVersion(ctx, dbPkg.ID, viewId, versionName)
	if err != nil {
		return nil, err
	}
	return dbVer, nil
}

func (s *Store) ListAssets(ctx context.Context, versionId int64) ([]*types.ArtifactAssetsRes, error) {
	result := make([]*types.ArtifactAssetsRes, 0)

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Model(&types.ArtifactAsset{}).
		Select(`artifact_assets.asset_id,artifact_assets.asset_path,artifact_assets.asset_content_type,
			artifact_assets.asset_checksum,artifact_assets.asset_created,
			artifact_blobs.blob_size,artifact_blobs.blob_created,principals.principal_uid_unique`).
		Joins(`LEFT JOIN artifact_blobs on artifact_assets.asset_blob_id=artifact_blobs.blob_id`).
		Joins(`LEFT JOIN principals on artifact_blobs.blob_creator=principals.principal_id`).
		Where(`artifact_assets.asset_version_id = ?`, versionId).
		Where(`artifact_assets.asset_deleted = 0`)

	if err := stmt.Find(&result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec artifact list asset by version failed")
	}

	for _, a := range result {
		var checkSum types.CheckSumRes
		if err := json.Unmarshal([]byte(a.CheckSumString), &checkSum); err != nil {
			continue
		}
		a.CheckSum = &checkSum
	}

	return result, nil
}

func (s *Store) GetAsset(ctx context.Context, assetId int64) (*types.ArtifactAssetsRes, error) {
	var result *types.ArtifactAssetsRes

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Model(&types.ArtifactAsset{}).
		Select(`artifact_assets.asset_id,artifact_assets.asset_path,artifact_assets.asset_content_type,
			artifact_assets.asset_checksum,artifact_assets.asset_created,
			artifact_blobs.blob_size,artifact_blobs.blob_created,principals.principal_uid_unique`).
		Joins(`LEFT JOIN artifact_blobs on artifact_assets.asset_blob_id=artifact_blobs.blob_id`).
		Joins(`LEFT JOIN principals on artifact_blobs.blob_creator=principals.principal_id`).
		Where(`artifact_assets.asset_id = ?`, assetId).
		Where(`artifact_assets.asset_deleted = 0`)

	if err := stmt.Take(&result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "exec artifact get asset failed")
	}

	var checkSum types.CheckSumRes
	if err := json.Unmarshal([]byte(result.CheckSumString), &checkSum); err == nil {
		result.CheckSum = &checkSum
	}

	return result, nil
}
