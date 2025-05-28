-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE infra_provider_configs ADD CONSTRAINT infra_provider_configs_ipconf_uid_ipconf_space_id_key UNIQUE (ipconf_uid, ipconf_space_id);
ALTER TABLE infra_provider_templates ADD CONSTRAINT infra_provider_templates_iptemp_uid_iptemp_space_id_key UNIQUE (iptemp_uid, iptemp_space_id);
ALTER TABLE infra_provider_resources ADD CONSTRAINT infra_provider_resources_ipreso_uid_ipreso_space_id_key UNIQUE (ipreso_uid, ipreso_space_id);
ALTER TABLE gitspace_configs ADD CONSTRAINT gitspace_configs_gconf_uid_gconf_space_id_key UNIQUE (gconf_uid, gconf_space_id);
ALTER TABLE gitspaces ADD CONSTRAINT gitspaces_gits_gitspace_config_id_gits_space_id_key UNIQUE (gits_gitspace_config_id, gits_space_id);
DROP INDEX infra_provider_configs_uid_space_id ;
DROP INDEX infra_provider_templates_uid_space_id;
DROP INDEX infra_provider_resources_uid_space_id;
DROP INDEX gitspace_configs_uid_space_id;
DROP INDEX gitspaces_gitspace_config_id_space_id;

ALTER TABLE infra_provider_configs CHANGE COLUMN ipconf_display_name ipconf_name VARCHAR(255);
ALTER TABLE infra_provider_resources CHANGE COLUMN ipreso_display_name ipreso_name VARCHAR(255);
ALTER TABLE gitspace_configs CHANGE COLUMN gconf_display_name gconf_name VARCHAR(255);
