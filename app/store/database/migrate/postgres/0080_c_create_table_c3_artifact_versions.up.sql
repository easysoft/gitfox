-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE artifact_versions (
  version_id          SERIAL PRIMARY KEY,
  version_package_id  INTEGER,
  version             VARCHAR(100),
  version_view_id     INTEGER,
  version_metadata    TEXT,
  version_created     BIGINT,
  version_updated     BIGINT,
  version_deleted     BIGINT NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX idx_pkg_version_view ON artifact_versions (version_view_id,version_package_id,version);
CREATE INDEX idx_version_created ON artifact_versions (version_created);
CREATE INDEX idx_version_updated ON artifact_versions (version_updated desc);
