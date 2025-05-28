-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE artifact_assets ADD COLUMN asset_deleted BIGINT NOT NULL DEFAULT 0;

DROP INDEX idx_asset_ver_path ON artifact_assets;

CREATE UNIQUE INDEX idx_asset_path_ver ON artifact_assets (asset_path,asset_version_id,asset_deleted);
