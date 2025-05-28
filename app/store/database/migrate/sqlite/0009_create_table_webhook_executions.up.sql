-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE webhook_executions (
webhook_execution_id INTEGER PRIMARY KEY AUTOINCREMENT
,webhook_execution_retrigger_of INTEGER
,webhook_execution_retriggerable BOOLEAN NOT NULL
,webhook_execution_webhook_id INTEGER NOT NULL
,webhook_execution_trigger_type TEXT NOT NULL
,webhook_execution_trigger_id TEXT NOT NULL
,webhook_execution_result TEXT NOT NULL
,webhook_execution_created BIGINT NOT NULL
,webhook_execution_duration BIGINT NOT NULL
,webhook_execution_error TEXT NOT NULL
,webhook_execution_request_url TEXT NOT NULL
,webhook_execution_request_headers TEXT NOT NULL
,webhook_execution_request_body TEXT NOT NULL
,webhook_execution_response_status_code INTEGER NOT NULL
,webhook_execution_response_status TEXT NOT NULL
,webhook_execution_response_headers TEXT NOT NULL
,webhook_execution_response_body TEXT NOT NULL
);
