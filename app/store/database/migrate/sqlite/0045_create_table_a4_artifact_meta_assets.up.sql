-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE artifact_meta_assets (
  meta_asset_id             INTEGER PRIMARY KEY AUTOINCREMENT,
  meta_asset_owner_id       INTEGER,
  meta_asset_view_id        INTEGER,
  meta_asset_format         TEXT,
  meta_asset_path           TEXT,
  meta_asset_content_type   TEXT,
  meta_asset_kind           TEXT,
  meta_asset_blob_id        INTEGER,
  meta_asset_checksum       TEXT,
  meta_asset_created        BIGINT,
  meta_asset_updated        BIGINT
);

CREATE UNIQUE INDEX idx_view_format_path ON artifact_meta_assets (meta_asset_view_id,meta_asset_path,meta_asset_format);
