-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE secrets (
                       secret_id          INT PRIMARY KEY AUTO_INCREMENT,
                       secret_uid         VARCHAR(255) NOT NULL,
                       secret_space_id    INT NOT NULL,
                       secret_description TEXT NOT NULL,
                       secret_data        LONGBLOB NOT NULL,
                       secret_created     BIGINT NOT NULL,
                       secret_updated     BIGINT NOT NULL,
                       secret_version     INT NOT NULL,
                       secret_created_by  INT NOT NULL,
                       CONSTRAINT fk_secrets_space_id FOREIGN KEY (secret_space_id) REFERENCES spaces(space_id) ON DELETE CASCADE,
                       CONSTRAINT fk_secrets_created_by FOREIGN KEY (secret_created_by) REFERENCES principals(principal_id),
                       UNIQUE (secret_space_id, secret_uid)
);
