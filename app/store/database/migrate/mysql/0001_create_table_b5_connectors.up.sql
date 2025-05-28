-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE connectors (
                          connector_id          INT PRIMARY KEY AUTO_INCREMENT,
                          connector_uid         VARCHAR(255) NOT NULL,
                          connector_description TEXT NOT NULL,
                          connector_type        VARCHAR(100) NOT NULL,
                          connector_space_id    INT NOT NULL,
                          connector_data        TEXT NOT NULL,
                          connector_created     BIGINT NOT NULL,
                          connector_updated     BIGINT NOT NULL,
                          connector_version     INT NOT NULL,
                          CONSTRAINT fk_connectors_space_id FOREIGN KEY (connector_space_id) REFERENCES spaces(space_id) ON DELETE CASCADE,
                          UNIQUE (connector_space_id, connector_uid)
);
