-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

create table upstream_proxy_configs_dg_tmp
(
    upstream_proxy_config_id                INTEGER
        primary key autoincrement,
    upstream_proxy_config_registry_id       INTEGER not null
        constraint fk_upstream_proxy_config_registry_id
            references registries
            on delete cascade,
    upstream_proxy_config_source            text,
    upstream_proxy_config_url               text,
    upstream_proxy_config_auth_type         text    not null,
    upstream_proxy_config_user_name         text,
    upstream_proxy_config_secret_identifier text,
    upstream_proxy_config_secret_space_id   int,
    upstream_proxy_config_token             text,
    upstream_proxy_config_created_at        INTEGER,
    upstream_proxy_config_updated_at        INTEGER,
    upstream_proxy_config_created_by        INTEGER,
    upstream_proxy_config_updated_by        INTEGER
);

insert into upstream_proxy_configs_dg_tmp(upstream_proxy_config_id, upstream_proxy_config_registry_id,
                                          upstream_proxy_config_source, upstream_proxy_config_url,
                                          upstream_proxy_config_auth_type, upstream_proxy_config_user_name,
                                          upstream_proxy_config_secret_identifier,
                                          upstream_proxy_config_secret_space_id, upstream_proxy_config_token,
                                          upstream_proxy_config_created_at, upstream_proxy_config_updated_at,
                                          upstream_proxy_config_created_by, upstream_proxy_config_updated_by)
select upstream_proxy_config_id,
       upstream_proxy_config_registry_id,
       upstream_proxy_config_source,
       upstream_proxy_config_url,
       upstream_proxy_config_auth_type,
       upstream_proxy_config_user_name,
       upstream_proxy_config_secret_identifier,
       upstream_proxy_config_secret_space_id,
       upstream_proxy_config_token,
       upstream_proxy_config_created_at,
       upstream_proxy_config_updated_at,
       upstream_proxy_config_created_by,
       upstream_proxy_config_updated_by
from upstream_proxy_configs;

drop table upstream_proxy_configs;

alter table upstream_proxy_configs_dg_tmp
    rename to upstream_proxy_configs;

create index index_upstream_proxy_config_on_registry_id
    on upstream_proxy_configs (upstream_proxy_config_registry_id);

