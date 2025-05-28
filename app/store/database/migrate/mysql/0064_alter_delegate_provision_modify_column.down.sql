-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE delegate_provision_details
  DROP FOREIGN KEY fk_dpdeta_gitspace_instance_identifier_space_id,
  DROP COLUMN dpdeta_gitspace_instance_identifier,
  ADD COLUMN dpdeta_gitspace_instance_id INTEGER NOT NULL,
  ADD CONSTRAINT fk_dpdeta_gitspace_instance_id FOREIGN KEY (dpdeta_gitspace_instance_id)
    REFERENCES gitspaces (gits_id) ON UPDATE NO ACTION;
