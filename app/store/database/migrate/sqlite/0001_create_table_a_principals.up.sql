-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE principals (
principal_id              INTEGER PRIMARY KEY AUTOINCREMENT
,principal_uid            TEXT
,principal_uid_unique     TEXT
,principal_email          TEXT
,principal_type           TEXT
,principal_display_name   TEXT
,principal_admin          BOOLEAN
,principal_blocked        BOOLEAN
,principal_salt           TEXT
,principal_created        BIGINT
,principal_updated        BIGINT

,principal_user_password  TEXT

,principal_sa_parent_type TEXT
,principal_sa_parent_id   INTEGER

,UNIQUE(principal_uid_unique)
);
