-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE artifact_packages (
  package_id          SERIAL PRIMARY KEY,
  package_owner_id    INTEGER,
  package_name        VARCHAR(100),
  package_namespace   VARCHAR(100),
  package_format      VARCHAR(20),
  package_created     BIGINT,
  package_updated     BIGINT,
  package_deleted     BIGINT NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX idx_unique_pkg ON artifact_packages (package_owner_id,package_name,package_namespace,package_format);
