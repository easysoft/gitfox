-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.


CREATE TABLE mirrors (
                       id INTEGER PRIMARY KEY AUTOINCREMENT,
                       space_id                INTEGER,
                       repo_id                 INTEGER,
                       sync_interval BIGINT DEFAULT NULL,
                       enable_prune BOOLEAN NOT NULL DEFAULT TRUE,
                       updated_unix BIGINT DEFAULT NULL,
                       next_update_unix BIGINT DEFAULT NULL,
                       lfs_enabled BOOLEAN NOT NULL DEFAULT FALSE,
                       remote_address VARCHAR(2048) DEFAULT NULL,
                       CONSTRAINT fk_mirror_space_id FOREIGN KEY (space_id) REFERENCES spaces(space_id) ON DELETE CASCADE,
                       CONSTRAINT fk_mirror_repo_id FOREIGN KEY (repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE
);

CREATE INDEX mirrors_repo_id ON mirrors (repo_id);
CREATE INDEX mirrors_space_id ON mirrors (space_id);
CREATE INDEX mirrors_next_update_unix ON mirrors (next_update_unix);
CREATE INDEX mirrors_updated_unix ON mirrors (updated_unix);
