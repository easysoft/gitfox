-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE plugins (
                       plugin_uid         VARCHAR(255) NOT NULL UNIQUE,
                       plugin_description TEXT NOT NULL,
                       plugin_logo        TEXT NOT NULL,
                       plugin_spec        LONGBLOB NOT NULL,
                       plugin_type        VARCHAR(255) NOT NULL,
                       plugin_version     VARCHAR(255) NOT NULL
);
