-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE public_keys (
 public_key_id SERIAL PRIMARY KEY
,public_key_principal_id INTEGER NOT NULL
,public_key_created BIGINT NOT NULL
,public_key_verified BIGINT
,public_key_identifier TEXT NOT NULL
,public_key_usage TEXT NOT NULL
,public_key_fingerprint TEXT NOT NULL
,public_key_content TEXT NOT NULL
,public_key_comment TEXT NOT NULL
,public_key_type TEXT NOT NULL
,CONSTRAINT fk_public_key_principal_id FOREIGN KEY (public_key_principal_id)
    REFERENCES principals (principal_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);

CREATE INDEX public_keys_fingerprint
    ON public_keys(public_key_fingerprint);

CREATE UNIQUE INDEX public_keys_principal_id_identifier
    ON public_keys(public_key_principal_id, LOWER(public_key_identifier));
