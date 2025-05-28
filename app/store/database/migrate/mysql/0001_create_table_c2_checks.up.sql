-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE checks (
                      check_id              INT PRIMARY KEY AUTO_INCREMENT,
                      check_created_by      INT NOT NULL,
                      check_created         BIGINT NOT NULL,
                      check_updated         BIGINT NOT NULL,
                      check_repo_id         INT NOT NULL,
                      check_commit_sha      VARCHAR(255) NOT NULL,
                      check_uid             VARCHAR(255) NOT NULL,
                      check_status          VARCHAR(255) NOT NULL,
                      check_summary         VARCHAR(255) NOT NULL,
                      check_link            VARCHAR(255) NOT NULL,
                      check_payload         TEXT NOT NULL,
                      check_metadata        TEXT NOT NULL,
                      check_payload_version VARCHAR(255) DEFAULT '' NOT NULL,
                      check_payload_kind    VARCHAR(255) DEFAULT '' NOT NULL,
                      check_started         BIGINT DEFAULT 0 NOT NULL,
                      check_ended           BIGINT DEFAULT 0 NOT NULL,
                      CONSTRAINT fk_check_created_by FOREIGN KEY (check_created_by) REFERENCES principals(principal_id),
                      CONSTRAINT fk_check_repo_id FOREIGN KEY (check_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX checks_repo_id_commit_sha_uid ON checks (check_repo_id, check_commit_sha, check_uid);
CREATE INDEX checks_repo_id_created ON checks (check_repo_id, check_created);
