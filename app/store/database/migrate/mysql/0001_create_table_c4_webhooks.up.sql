-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE webhooks (
                        webhook_id                      INT PRIMARY KEY AUTO_INCREMENT,
                        webhook_version                 INT DEFAULT 0 NOT NULL,
                        webhook_created_by              INT NOT NULL,
                        webhook_created                 BIGINT NOT NULL,
                        webhook_updated                 BIGINT NOT NULL,
                        webhook_space_id                INT,
                        webhook_repo_id                 INT,
                        webhook_display_name            VARCHAR(255) NOT NULL,
                        webhook_description             VARCHAR(255) NOT NULL,
                        webhook_url                     VARCHAR(255) NOT NULL,
                        webhook_secret                  VARCHAR(255) NOT NULL,
                        webhook_enabled                 BOOLEAN NOT NULL,
                        webhook_insecure                BOOLEAN NOT NULL,
                        webhook_triggers                TEXT NOT NULL,
                        webhook_latest_execution_result VARCHAR(255),
                        webhook_internal                BOOLEAN DEFAULT FALSE NOT NULL,
                        webhook_uid                     VARCHAR(255),
                        CONSTRAINT fk_webhook_created_by FOREIGN KEY (webhook_created_by) REFERENCES principals(principal_id),
                        CONSTRAINT fk_webhook_space_id FOREIGN KEY (webhook_space_id) REFERENCES spaces(space_id) ON DELETE CASCADE,
                        CONSTRAINT fk_webhook_repo_id FOREIGN KEY (webhook_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX webhooks_repo_id_uid ON webhooks (webhook_repo_id, webhook_uid);
CREATE UNIQUE INDEX webhooks_space_id_uid ON webhooks (webhook_space_id, webhook_uid);
