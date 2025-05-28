-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

-- recreate table without index
CREATE TABLE tokens_new (
 token_id             INTEGER PRIMARY KEY AUTOINCREMENT
,token_type           TEXT
,token_uid            TEXT
,token_principal_id   INTEGER
,token_expires_at     BIGINT
,token_issued_at      BIGINT
,token_created_by     INTEGER

,CONSTRAINT fk_token_principal_id FOREIGN KEY (token_principal_id)
    REFERENCES principals (principal_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);

-- copy over data
INSERT INTO tokens_new(
     token_id
    ,token_type
    ,token_uid
    ,token_principal_id
    ,token_expires_at
    ,token_issued_at
    ,token_created_by
)
SELECT 
     token_id
    ,token_type
    ,token_uid
    ,token_principal_id
    ,token_expires_at
    ,token_issued_at
    ,token_created_by
FROM tokens;

-- delete old table (also deletes all indices)
DROP TABLE tokens;

-- rename table
ALTER TABLE tokens_new RENAME TO tokens;

-- create explicit unique index with case insensitivity
CREATE UNIQUE INDEX tokens_principal_id_uid ON tokens(token_principal_id, LOWER(token_uid));

-- recreate old indices if needed (principal_id can be ignored since above index includes it)
CREATE INDEX tokens_type_expires_at ON tokens(token_type, token_expires_at);

