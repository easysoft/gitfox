-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE stages (
                      stage_id              INT PRIMARY KEY AUTO_INCREMENT,
                      stage_execution_id    INT NOT NULL,
                      stage_repo_id         INT NOT NULL,
                      stage_number          INT NOT NULL,
                      stage_kind            VARCHAR(255) NOT NULL,
                      stage_type            VARCHAR(255) NOT NULL,
                      stage_name            VARCHAR(255) NOT NULL,
                      stage_status          VARCHAR(255) NOT NULL,
                      stage_error           TEXT NOT NULL,
                      stage_parent_group_id INT NOT NULL,
                      stage_errignore       BOOLEAN NOT NULL,
                      stage_exit_code       INT NOT NULL,
                      stage_limit           INT NOT NULL,
                      stage_os              VARCHAR(255) NOT NULL,
                      stage_arch            VARCHAR(255) NOT NULL,
                      stage_variant         VARCHAR(255) NOT NULL,
                      stage_kernel          VARCHAR(255) NOT NULL,
                      stage_machine         VARCHAR(255) NOT NULL,
                      stage_started         BIGINT NOT NULL,
                      stage_stopped         BIGINT NOT NULL,
                      stage_created         BIGINT NOT NULL,
                      stage_updated         BIGINT NOT NULL,
                      stage_version         INT NOT NULL,
                      stage_on_success      BOOLEAN NOT NULL,
                      stage_on_failure      BOOLEAN NOT NULL,
                      stage_depends_on      TEXT NOT NULL,
                      stage_labels          TEXT NOT NULL,
                      stage_limit_repo      INT DEFAULT 0 NOT NULL,
                      UNIQUE (stage_execution_id, stage_number),
                      CONSTRAINT fk_stages_execution_id FOREIGN KEY (stage_execution_id) REFERENCES executions(execution_id) ON DELETE CASCADE,
                      INDEX ix_stage_in_progress (stage_status) USING BTREE
);
