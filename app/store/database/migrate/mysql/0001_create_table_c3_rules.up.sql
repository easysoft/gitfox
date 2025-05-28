-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE rules (
                     rule_id          INT PRIMARY KEY AUTO_INCREMENT,
                     rule_version     INT NOT NULL,
                     rule_created_by  INT,
                     rule_created     BIGINT NOT NULL,
                     rule_updated     BIGINT NOT NULL,
                     rule_space_id    INT,
                     rule_repo_id     INT,
                     rule_uid         VARCHAR(255) NOT NULL,
                     rule_description TEXT NOT NULL,
                     rule_type        VARCHAR(255) NOT NULL,
                     rule_state       VARCHAR(255) NOT NULL,
                     rule_pattern     TEXT NOT NULL,
                     rule_definition  TEXT NOT NULL,
                     CONSTRAINT fk_rule_created_by FOREIGN KEY (rule_created_by) REFERENCES principals(principal_id) ON DELETE SET NULL,
                     CONSTRAINT fk_rule_space_id FOREIGN KEY (rule_space_id) REFERENCES spaces(space_id) ON DELETE CASCADE,
                     CONSTRAINT fk_rule_repo_id FOREIGN KEY (rule_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX rules_repo_id_uid ON rules (rule_repo_id, rule_uid);
CREATE UNIQUE INDEX rules_space_id_uid ON rules (rule_space_id, rule_uid);
