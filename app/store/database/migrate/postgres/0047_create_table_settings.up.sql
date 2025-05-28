-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE settings (
 setting_id SERIAL PRIMARY KEY
,setting_space_id INTEGER
,setting_repo_id INTEGER
,setting_key TEXT NOT NULL
,setting_value JSON

,CONSTRAINT fk_settings_space_id FOREIGN KEY (setting_space_id)
    REFERENCES spaces (space_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
,CONSTRAINT fk_settings_repo_id FOREIGN KEY (setting_repo_id)
    REFERENCES repositories (repo_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);

CREATE UNIQUE INDEX settings_space_id_key
	ON settings(setting_space_id, LOWER(setting_key))
	WHERE setting_space_id IS NOT NULL;

CREATE UNIQUE INDEX settings_repo_id_key
	ON settings(setting_repo_id, LOWER(setting_key))
	WHERE setting_repo_id IS NOT NULL;