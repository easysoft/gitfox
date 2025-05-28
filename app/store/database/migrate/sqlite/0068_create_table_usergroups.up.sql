-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE usergroups
(
    usergroup_id          INTEGER PRIMARY KEY AUTOINCREMENT,
    usergroup_identifier  TEXT NOT NULL,
    usergroup_name        TEXT NOT NULL,
    usergroup_description TEXT,
    usergroup_space_id    INTEGER NOT NULL,
    usergroup_created     BIGINT,
    usergroup_updated     BIGINT,

    CONSTRAINT fk_usergroup_space_id FOREIGN KEY (usergroup_space_id)
        REFERENCES spaces (space_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX usergroups_space_id_identifier ON usergroups (usergroup_space_id, LOWER(usergroup_identifier));

