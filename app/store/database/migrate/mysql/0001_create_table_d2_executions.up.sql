-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE executions (
                          execution_id            INT PRIMARY KEY AUTO_INCREMENT,
                          execution_pipeline_id   INT NOT NULL,
                          execution_repo_id       INT NOT NULL,
                          execution_created_by    INT NOT NULL,
                          execution_trigger       VARCHAR(255) NOT NULL,
                          execution_number        INT NOT NULL,
                          execution_parent        INT NOT NULL,
                          execution_status        VARCHAR(255) NOT NULL,
                          execution_error         TEXT NOT NULL,
                          execution_event         TEXT NOT NULL,
                          execution_action        VARCHAR(255) NOT NULL,
                          execution_link          VARCHAR(255) NOT NULL,
                          execution_timestamp     INT NOT NULL,
                          execution_title         TEXT NOT NULL,
                          execution_message       TEXT NOT NULL,
                          execution_before        VARCHAR(255) NOT NULL,
                          execution_after         VARCHAR(255) NOT NULL,
                          execution_ref           VARCHAR(255) NOT NULL,
                          execution_source_repo   VARCHAR(255) NOT NULL,
                          execution_source        VARCHAR(255) NOT NULL,
                          execution_target        VARCHAR(255) NOT NULL,
                          execution_author        VARCHAR(255) NOT NULL,
                          execution_author_name   VARCHAR(255) NOT NULL,
                          execution_author_email  VARCHAR(255) NOT NULL,
                          execution_author_avatar VARCHAR(255) NOT NULL,
                          execution_sender        VARCHAR(255) NOT NULL,
                          execution_params        TEXT NOT NULL,
                          execution_cron          VARCHAR(255) NOT NULL,
                          execution_deploy        VARCHAR(255) NOT NULL,
                          execution_deploy_id     INT NOT NULL,
                          execution_debug         BOOLEAN DEFAULT 0 NOT NULL,
                          execution_started       BIGINT NOT NULL,
                          execution_finished      BIGINT NOT NULL,
                          execution_created       BIGINT NOT NULL,
                          execution_updated       BIGINT NOT NULL,
                          execution_version       INT NOT NULL,
                          UNIQUE (execution_pipeline_id, execution_number),
                          CONSTRAINT fk_executions_pipeline_id FOREIGN KEY (execution_pipeline_id) REFERENCES pipelines(pipeline_id) ON DELETE CASCADE,
                          CONSTRAINT fk_executions_repo_id FOREIGN KEY (execution_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE,
                          CONSTRAINT fk_executions_created_by FOREIGN KEY (execution_created_by) REFERENCES principals(principal_id)
);
