-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE webhooks (
webhook_id SERIAL PRIMARY KEY
,webhook_version INTEGER NOT NULL DEFAULT 0
,webhook_created_by INTEGER NOT NULL
,webhook_created BIGINT NOT NULL
,webhook_updated BIGINT NOT NULL
,webhook_space_id INTEGER
,webhook_repo_id INTEGER
,webhook_display_name TEXT NOT NULL
,webhook_description TEXT NOT NULL
,webhook_url TEXT NOT NULL
,webhook_secret TEXT NOT NULL
,webhook_enabled BOOLEAN NOT NULL
,webhook_insecure BOOLEAN NOT NULL
,webhook_triggers TEXT NOT NULL
,webhook_latest_execution_result TEXT
,CONSTRAINT fk_webhook_created_by FOREIGN KEY (webhook_created_by)
    REFERENCES principals (principal_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
,CONSTRAINT fk_webhook_space_id FOREIGN KEY (webhook_space_id)
    REFERENCES spaces (space_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
,CONSTRAINT fk_webhook_repo_id FOREIGN KEY (webhook_repo_id)
    REFERENCES repositories (repo_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);
