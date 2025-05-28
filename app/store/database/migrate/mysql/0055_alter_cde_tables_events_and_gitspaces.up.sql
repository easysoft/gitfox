-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE gitspace_events DROP FOREIGN KEY fk_geven_gitspace_config_id;
ALTER TABLE gitspace_events DROP COLUMN geven_gitspace_config_id;
ALTER TABLE gitspace_events
  ADD COLUMN geven_entity_type TEXT NOT NULL;
ALTER TABLE gitspace_events
  ADD COLUMN geven_entity_uid TEXT;
ALTER TABLE gitspace_events
  ADD COLUMN geven_entity_id INT NOT NULL;
CREATE INDEX gitspace_events_entity_id ON gitspace_events (geven_entity_id);

ALTER TABLE gitspaces DROP FOREIGN KEY fk_gits_infra_provisioned_id;
ALTER TABLE gitspaces DROP COLUMN gits_infra_provisioned_id;
DROP INDEX gitspaces_gitspace_config_id_space_id on gitspaces;
ALTER TABLE gitspaces ADD COLUMN gits_uid VARCHAR(255) NOT NULL;
CREATE UNIQUE INDEX gitspaces_uid_space_id ON gitspaces (gits_uid, gits_space_id);
