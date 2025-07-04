-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE pullreqs ADD COLUMN pullreq_merge_base_sha_not_nullable TEXT NOT NULL DEFAULT '';
UPDATE pullreqs SET pullreq_merge_base_sha_not_nullable = pullreq_merge_base_sha WHERE pullreq_merge_base_sha IS NOT NULL;
ALTER TABLE pullreqs DROP COLUMN pullreq_merge_base_sha;
ALTER TABLE pullreqs RENAME COLUMN pullreq_merge_base_sha_not_nullable TO pullreq_merge_base_sha;
