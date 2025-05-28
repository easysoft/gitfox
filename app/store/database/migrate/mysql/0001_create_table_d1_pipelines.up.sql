-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE pipelines (
                         pipeline_id             INT PRIMARY KEY AUTO_INCREMENT,
                         pipeline_description    VARCHAR(255) NOT NULL,
                         pipeline_uid            VARCHAR(255) NOT NULL,
                         pipeline_seq            INT DEFAULT 0 NOT NULL,
                         pipeline_disabled       BOOLEAN NOT NULL,
                         pipeline_repo_id        INT NOT NULL,
                         pipeline_default_branch VARCHAR(255) NOT NULL,
                         pipeline_created_by     INT NOT NULL,
                         pipeline_config_path    VARCHAR(255) NOT NULL,
                         pipeline_created        BIGINT NOT NULL,
                         pipeline_updated        BIGINT NOT NULL,
                         pipeline_version        INT NOT NULL,
                         UNIQUE (pipeline_repo_id, pipeline_uid),
                         CONSTRAINT fk_pipelines_repo_id FOREIGN KEY (pipeline_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE,
                         CONSTRAINT fk_pipelines_created_by FOREIGN KEY (pipeline_created_by) REFERENCES principals(principal_id)
);
