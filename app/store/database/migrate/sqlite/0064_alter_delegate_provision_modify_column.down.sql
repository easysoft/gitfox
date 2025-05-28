-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

DROP TABLE delegate_provision_details;

CREATE TABLE delegate_provision_details
(
    dpdeta_id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    dpdeta_task_id              TEXT    NOT NULL,
    dpdeta_action_type          TEXT    NOT NULL,
    dpdeta_gitspace_instance_id INTEGER NOT NULL,
    dpdeta_space_id             INTEGER NOT NULL,
    dpdeta_agent_port           INTEGER NOT NULL,
    dpdeta_created              BIGINT  NOT NULL,
    dpdeta_updated              BIGINT  NOT NULL,
    CONSTRAINT fk_dpdeta_gitspace_instance_id FOREIGN KEY (dpdeta_gitspace_instance_id)
        REFERENCES gitspaces (gits_id) MATCH SIMPLE
        ON UPDATE NO ACTION
    CONSTRAINT fk_dpdeta_space_id FOREIGN KEY (dpdeta_space_id)
        REFERENCES spaces (space_id) MATCH SIMPLE
        ON UPDATE NO ACTION
);

CREATE UNIQUE INDEX delegate_provision_details_task_id_space_id ON delegate_provision_details (dpdeta_task_id, dpdeta_space_id);