-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE artifact_views (
  view_id           INTEGER PRIMARY KEY AUTO_INCREMENT,
  view_name         VARCHAR(100),
  view_description  TEXT,
  view_space_id     INTEGER,
  view_is_default   BOOLEAN
);

