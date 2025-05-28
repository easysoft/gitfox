-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE tokens (
                      token_id           INT PRIMARY KEY AUTO_INCREMENT,
                      token_type         VARCHAR(20),
                      token_uid          VARCHAR(255),
                      token_principal_id INT,
                      token_expires_at   BIGINT,
                      token_issued_at    BIGINT,
                      token_created_by   INTEGER,
                      CONSTRAINT fk_token_principal_id FOREIGN KEY (token_principal_id) REFERENCES principals(principal_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX tokens_principal_id_uid ON tokens (token_principal_id, token_uid);
CREATE INDEX tokens_type_expires_at ON tokens (token_type, token_expires_at);
