-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE pullreqs (
                        pullreq_id                 INT PRIMARY KEY AUTO_INCREMENT,
                        pullreq_version            INT DEFAULT 0 NOT NULL,
                        pullreq_created_by         INT NOT NULL,
                        pullreq_created            BIGINT NOT NULL,
                        pullreq_updated            BIGINT NOT NULL,
                        pullreq_edited             BIGINT NOT NULL,
                        pullreq_number             INT NOT NULL,
                        pullreq_state              VARCHAR(255) NOT NULL,
                        pullreq_is_draft           VARCHAR(255) DEFAULT 'FALSE' NOT NULL,
                        pullreq_comment_count      INT DEFAULT 0 NOT NULL,
                        pullreq_title              TEXT NOT NULL,
                        pullreq_description        TEXT NOT NULL,
                        pullreq_source_repo_id     INT NOT NULL,
                        pullreq_source_branch      VARCHAR(255) NOT NULL,
                        pullreq_source_sha         VARCHAR(255) NOT NULL,
                        pullreq_target_repo_id     INT NOT NULL,
                        pullreq_target_branch      VARCHAR(255) NOT NULL,
                        pullreq_activity_seq       INT DEFAULT 0,
                        pullreq_merged_by          INT,
                        pullreq_merged             BIGINT,
                        pullreq_merge_method       VARCHAR(255),
                        pullreq_merge_check_status VARCHAR(255) NOT NULL,
                        pullreq_merge_target_sha   VARCHAR(255),
                        pullreq_merge_sha          VARCHAR(255),
                        pullreq_merge_conflicts    TEXT,
                        pullreq_merge_base_sha     VARCHAR(255) DEFAULT '' NOT NULL,
                        pullreq_unresolved_count   INT DEFAULT 0 NOT NULL,
                        pullreq_commit_count       INT,
                        pullreq_file_count         INT,
                        CONSTRAINT fk_pullreq_created_by FOREIGN KEY (pullreq_created_by) REFERENCES principals(principal_id),
                        CONSTRAINT fk_pullreq_source_repo_id FOREIGN KEY (pullreq_source_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE,
                        CONSTRAINT fk_pullreq_target_repo_id FOREIGN KEY (pullreq_target_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE
);

CREATE INDEX pullreqs_source_repo_branch_target_repo_branch
  ON pullreqs (pullreq_source_repo_id, pullreq_source_branch, pullreq_target_repo_id, pullreq_target_branch);

CREATE UNIQUE INDEX pullreqs_target_repo_id_number
  ON pullreqs (pullreq_target_repo_id, pullreq_number);
