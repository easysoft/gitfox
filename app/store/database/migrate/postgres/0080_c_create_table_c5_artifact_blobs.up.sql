-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE artifact_blobs (
  blob_id           SERIAL PRIMARY KEY,
  storage_id        INTEGER,
  blob_ref          VARCHAR(50),
  blob_size         BIGINT,
  blob_metadata     TEXT default NULL,
  blob_downloads    BIGINT,
  blob_deleted      BIGINT DEFAULT NULL,
  blob_created      BIGINT,
  blob_creator      BIGINT NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX idx_store_ref ON artifact_blobs (blob_ref,storage_id);
