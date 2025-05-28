-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

DELETE
FROM gitspace_events;
DELETE
FROM SQLITE_SEQUENCE
WHERE name = 'gitspace_events';

DELETE
FROM gitspaces;
DELETE
FROM SQLITE_SEQUENCE
WHERE name = 'gitspaces';

DELETE
FROM gitspace_configs;
DELETE
FROM SQLITE_SEQUENCE
WHERE name = 'gitspace_configs';

DELETE
FROM infra_provider_resources;
DELETE
FROM SQLITE_SEQUENCE
WHERE name = 'infra_provider_resources';

DELETE
FROM infra_provider_configs;
DELETE
FROM SQLITE_SEQUENCE
WHERE name = 'infra_provider_configs';

ALTER TABLE gitspace_configs
    DROP COLUMN gconf_code_repo_ref;
