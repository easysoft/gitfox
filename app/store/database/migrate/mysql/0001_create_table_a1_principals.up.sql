-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE principals (
                          principal_id             INTEGER PRIMARY KEY AUTO_INCREMENT,
                          principal_uid            varchar(255),
                          principal_uid_unique     varchar(255) UNIQUE,
                          principal_email          varchar(60),
                          principal_type           varchar(25),
                          principal_display_name   varchar(100),
                          principal_admin          BOOLEAN,
                          principal_blocked        BOOLEAN,
                          principal_salt           varchar(25),
                          principal_created        BIGINT,
                          principal_updated        BIGINT,
                          principal_user_password  varchar(100),
                          principal_sa_parent_type varchar(10),
                          principal_sa_parent_id   INTEGER
);
CREATE UNIQUE INDEX principals_lower_email ON principals (principal_email);
CREATE INDEX principals_sa_parent_id_sa_parent_type ON principals (principal_sa_parent_id, principal_sa_parent_type);
