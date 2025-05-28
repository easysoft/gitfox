-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE triggers (
                        trigger_id          INT PRIMARY KEY AUTO_INCREMENT,
                        trigger_uid         VARCHAR(255) NOT NULL,
                        trigger_pipeline_id INT NOT NULL,
                        trigger_type        VARCHAR(255) NOT NULL,
                        trigger_repo_id     INT NOT NULL,
                        trigger_secret      VARCHAR(255) NOT NULL,
                        trigger_description VARCHAR(255) NOT NULL,
                        trigger_disabled    BOOLEAN NOT NULL,
                        trigger_created_by  INT NOT NULL,
                        trigger_actions     TEXT NOT NULL,
                        trigger_created     BIGINT NOT NULL,
                        trigger_updated     BIGINT NOT NULL,
                        trigger_version     INT NOT NULL,
                        UNIQUE (trigger_pipeline_id, trigger_uid),
                        CONSTRAINT fk_triggers_pipeline_id FOREIGN KEY (trigger_pipeline_id) REFERENCES pipelines(pipeline_id) ON DELETE CASCADE,
                        CONSTRAINT fk_triggers_repo_id FOREIGN KEY (trigger_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE
);
