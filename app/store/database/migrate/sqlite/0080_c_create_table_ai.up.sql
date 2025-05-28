-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE ai_configs (
    ai_id           INTEGER PRIMARY KEY AUTOINCREMENT,
    ai_space_id     INTEGER NOT NULL,
    ai_created_by   INTEGER NOT NULL,
    ai_updated_by   INTEGER NOT NULL,
    ai_created      INTEGER NOT NULL,
    ai_updated      INTEGER NOT NULL,
    ai_token        TEXT NOT NULL,
    ai_provider     TEXT NOT NULL,
    ai_model        TEXT NOT NULL,
    ai_endpoint     TEXT NOT NULL,
    ai_default      INTEGER DEFAULT 0,
    deleted_at      TEXT DEFAULT NULL,
    FOREIGN KEY (ai_space_id) REFERENCES spaces(space_id) ON DELETE CASCADE
);

CREATE TABLE ai_requests (
    ai_id INT PRIMARY KEY AUTO_INCREMENT,
    ai_created BIGINT NOT NULL,
    ai_updated BIGINT NOT NULL,
    ai_repo_id INT NOT NULL,
    ai_pullreq_id INT NOT NULL,
    ai_config_id INT NOT NULL,
    ai_token INT NOT NULL,
    ai_duration INT NOT NULL,
    ai_status VARCHAR(255) NOT NULL,
    ai_error TEXT,
    ai_client_mode BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (ai_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE
);

