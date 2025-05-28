-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE gitspace_events_temp
(
    geven_id          INTEGER PRIMARY KEY AUTOINCREMENT,
    geven_event       TEXT    NOT NULL,
    geven_created     BIGINT  NOT NULL,
    geven_entity_type TEXT    NOT NULL,
    geven_query_key   TEXT,
    geven_entity_id   INTEGER NOT NULL,
    geven_timestamp   BIGINT  NOT NULL
);

INSERT INTO gitspace_events_temp (geven_id, geven_event, geven_created, geven_entity_type, geven_query_key,
                                  geven_entity_id, geven_timestamp)
SELECT geven_id,
       geven_event,
       geven_created,
       geven_entity_type,
       geven_entity_uid,
       geven_entity_id,
       geven_created * 1000000
FROM gitspace_events;

DROP TABLE gitspace_events;

ALTER TABLE gitspace_events_temp RENAME TO gitspace_events;

CREATE INDEX gitspace_events_entity_id ON gitspace_events (geven_entity_id);