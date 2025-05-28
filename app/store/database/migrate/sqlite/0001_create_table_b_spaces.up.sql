-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE spaces (
 space_id           INTEGER PRIMARY KEY AUTOINCREMENT
,space_version      INTEGER NOT NULL DEFAULT 0
,space_parent_id    INTEGER DEFAULT NULL
,space_uid          TEXT NOT NULL
,space_description  TEXT
,space_is_public    BOOLEAN NOT NULL
,space_created_by   INTEGER NOT NULL
,space_created      BIGINT NOT NULL
,space_updated      BIGINT NOT NULL

,CONSTRAINT fk_space_parent_id FOREIGN KEY (space_parent_id)
    REFERENCES spaces (space_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);