-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE webhooks ADD COLUMN webhook_uid TEXT;

CREATE UNIQUE INDEX webhooks_repo_id_uid
    ON webhooks(webhook_repo_id, LOWER(webhook_uid))
    WHERE webhook_space_id IS NULL;

CREATE UNIQUE INDEX webhooks_space_id_uid
    ON webhooks(webhook_space_id, LOWER(webhook_uid))
    WHERE webhook_repo_id IS NULL;


DROP INDEX webhooks_repo_id;
DROP INDEX webhooks_space_id;

-- code migration will backfill uids