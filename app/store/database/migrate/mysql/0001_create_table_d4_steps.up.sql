-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE steps (
                     step_id              INT PRIMARY KEY AUTO_INCREMENT,
                     step_stage_id        INT NOT NULL,
                     step_number          INT NOT NULL,
                     step_name            VARCHAR(100) NOT NULL,
                     step_status          VARCHAR(50) NOT NULL,
                     step_error           VARCHAR(500) NOT NULL,
                     step_parent_group_id INT NOT NULL,
                     step_errignore       BOOLEAN NOT NULL,
                     step_exit_code       INT NOT NULL,
                     step_started         BIGINT NOT NULL,
                     step_stopped         BIGINT NOT NULL,
                     step_version         INT NOT NULL,
                     step_depends_on      TEXT NOT NULL,
                     step_image           VARCHAR(255) NOT NULL,
                     step_detached        BOOLEAN NOT NULL,
                     step_schema          TEXT NOT NULL,
                     UNIQUE (step_stage_id, step_number),
                     CONSTRAINT fk_steps_stage_id FOREIGN KEY (step_stage_id) REFERENCES stages(stage_id) ON DELETE CASCADE
);
