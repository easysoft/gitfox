-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE gitspace_events CHANGE COLUMN geven_entity_uid geven_query_key VARCHAR(255);

ALTER TABLE gitspace_events ADD COLUMN geven_timestamp BIGINT;
UPDATE gitspace_events SET geven_timestamp = geven_created * 1000000;
ALTER TABLE gitspace_events MODIFY COLUMN geven_timestamp BIGINT NOT NULL;
