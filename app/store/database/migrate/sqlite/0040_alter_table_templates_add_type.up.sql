-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

DROP TABLE IF exists templates;
CREATE TABLE templates (
    template_id INTEGER PRIMARY KEY AUTOINCREMENT
    ,template_uid TEXT NOT NULL
    ,template_type TEXT NOT NULL
    ,template_description TEXT NOT NULL
    ,template_space_id INTEGER NOT NULL
    ,template_data BLOB NOT NULL
    ,template_created INTEGER NOT NULL
    ,template_updated INTEGER NOT NULL
    ,template_version INTEGER NOT NULL

    -- Ensure unique combination of space ID, UID and template type
    ,UNIQUE (template_space_id, template_uid, template_type)

    -- Foreign key to spaces table
    ,CONSTRAINT fk_templates_space_id FOREIGN KEY (template_space_id)
        REFERENCES spaces (space_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);