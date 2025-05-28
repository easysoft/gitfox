-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

DROP INDEX gitspace_events_entity_id;
ALTER TABLE gitspace_events DROP COLUMN geven_entity_type;
ALTER TABLE gitspace_events DROP COLUMN geven_entity_uid;
ALTER TABLE gitspace_events DROP COLUMN geven_entity_id;
ALTER TABLE gitspace_events
  ADD COLUMN geven_gitspace_config_id INTEGER NOT NULL,
  ADD CONSTRAINT fk_geven_gitspace_config_id FOREIGN KEY (geven_gitspace_config_id)
    REFERENCES gitspace_configs (gconf_id) ON UPDATE NO ACTION ON DELETE CASCADE;

ALTER TABLE gitspaces
  ADD COLUMN gits_infra_provisioned_id INTEGER,
  ADD CONSTRAINT fk_gits_infra_provisioned_id FOREIGN KEY (gits_infra_provisioned_id)
    REFERENCES infra_provisioned (iprov_id) ON UPDATE NO ACTION ON DELETE NO ACTION;
DROP INDEX gitspaces_uid_space_id;
ALTER TABLE gitspaces DROP COLUMN gits_uid;
CREATE UNIQUE INDEX gitspaces_gitspace_config_id_space_id ON gitspaces (gits_gitspace_config_id, gits_space_id);
