-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE webhook_executions (
                                  webhook_execution_id                   INT PRIMARY KEY AUTO_INCREMENT,
                                  webhook_execution_retrigger_of         INT,
                                  webhook_execution_retriggerable        BOOLEAN NOT NULL,
                                  webhook_execution_webhook_id           INT NOT NULL,
                                  webhook_execution_trigger_type         VARCHAR(255) NOT NULL,
                                  webhook_execution_trigger_id           VARCHAR(255) NOT NULL,
                                  webhook_execution_result               TEXT NOT NULL,
                                  webhook_execution_created              BIGINT NOT NULL,
                                  webhook_execution_duration             BIGINT NOT NULL,
                                  webhook_execution_error                TEXT NOT NULL,
                                  webhook_execution_request_url          VARCHAR(255) NOT NULL,
                                  webhook_execution_request_headers      TEXT NOT NULL,
                                  webhook_execution_request_body         TEXT NOT NULL,
                                  webhook_execution_response_status_code INT NOT NULL,
                                  webhook_execution_response_status      VARCHAR(255) NOT NULL,
                                  webhook_execution_response_headers     TEXT NOT NULL,
                                  webhook_execution_response_body        TEXT NOT NULL,
                                  INDEX webhook_executions_created (webhook_execution_created),
                                  INDEX webhook_executions_webhook_id (webhook_execution_webhook_id),
                                  CONSTRAINT fk_webhook_execution_webhook_id FOREIGN KEY (webhook_execution_webhook_id) REFERENCES webhooks(webhook_id)
);
