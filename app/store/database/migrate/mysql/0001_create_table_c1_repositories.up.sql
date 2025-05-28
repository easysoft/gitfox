-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE repositories (
                            repo_id               INT PRIMARY KEY AUTO_INCREMENT,
                            repo_version          INT DEFAULT 0 NOT NULL,
                            repo_parent_id        INT NOT NULL,
                            repo_uid              VARCHAR(255) NOT NULL,
                            repo_description      TEXT,
                            repo_is_public        BOOLEAN NOT NULL,
                            repo_created_by       INT NOT NULL,
                            repo_created          BIGINT NOT NULL,
                            repo_updated          BIGINT NOT NULL,
                            repo_git_uid          VARCHAR(255) NOT NULL UNIQUE,
                            repo_default_branch   VARCHAR(255) NOT NULL,
                            repo_fork_id          INT,
                            repo_pullreq_seq      INT NOT NULL,
                            repo_num_forks        INT NOT NULL,
                            repo_num_pulls        INT NOT NULL,
                            repo_num_closed_pulls INT NOT NULL,
                            repo_num_open_pulls   INT NOT NULL,
                            repo_num_merged_pulls INT NOT NULL,
                            repo_importing        BOOLEAN DEFAULT FALSE NOT NULL,
                            repo_size             INT DEFAULT 0 NOT NULL,
                            repo_size_updated     BIGINT DEFAULT 0 NOT NULL,
                            repo_deleted          BIGINT DEFAULT NULL,
                            CONSTRAINT fk_repo_parent_id FOREIGN KEY (repo_parent_id) REFERENCES spaces(space_id) ON DELETE CASCADE
);

CREATE INDEX repositories_deleted ON repositories (repo_deleted);

CREATE UNIQUE INDEX repositories_parent_id_uid ON repositories (repo_parent_id, repo_uid);
