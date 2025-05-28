-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE artifact_assets (
  asset_id            INTEGER PRIMARY KEY AUTOINCREMENT,
  asset_view_id       INTEGER,
  asset_version_id    INTEGER,
  asset_format        TEXT,
  asset_path          TEXT,
  asset_content_type  TEXT,
  asset_kind          TEXT,
  asset_metadata      TEXT,
  asset_blob_id       INTEGER,
  asset_checksum      TEXT,
  asset_created       BIGINT,
  asset_updated       BIGINT
);

CREATE UNIQUE INDEX idx_asset_ver_path ON artifact_assets (asset_version_id,asset_path);
