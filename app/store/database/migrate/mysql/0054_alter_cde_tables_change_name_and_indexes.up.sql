-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE infra_provider_configs RENAME COLUMN ipconf_name TO ipconf_display_name;
ALTER TABLE infra_provider_resources RENAME COLUMN ipreso_name TO ipreso_display_name;
ALTER TABLE gitspace_configs RENAME COLUMN gconf_name TO gconf_display_name;

--- TODO 忽略这个字段
--- ALTER TABLE infra_provider_configs DROP CONSTRAINT infra_provider_configs_ipconf_uid_ipconf_space_id_key;
--- ALTER TABLE infra_provider_templates DROP CONSTRAINT infra_provider_templates_iptemp_uid_iptemp_space_id_key;
--- ALTER TABLE infra_provider_resources DROP CONSTRAINT infra_provider_resources_ipreso_uid_ipreso_space_id_key;
--- ALTER TABLE gitspace_configs DROP CONSTRAINT gitspace_configs_gconf_uid_gconf_space_id_key;
--- ALTER TABLE gitspaces DROP CONSTRAINT gitspaces_gits_gitspace_config_id_gits_space_id_key;
CREATE UNIQUE INDEX infra_provider_configs_uid_space_id ON infra_provider_configs (ipconf_uid, ipconf_space_id);
CREATE UNIQUE INDEX infra_provider_templates_uid_space_id ON infra_provider_templates (iptemp_uid, iptemp_space_id);
CREATE UNIQUE INDEX infra_provider_resources_uid_space_id ON infra_provider_resources (ipreso_uid, ipreso_space_id);
CREATE UNIQUE INDEX gitspace_configs_uid_space_id ON gitspace_configs (gconf_uid, gconf_space_id);
CREATE UNIQUE INDEX gitspaces_gitspace_config_id_space_id ON gitspaces (gits_gitspace_config_id, gits_space_id);
