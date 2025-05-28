-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE pullreq_activities
    ADD COLUMN pullreq_activity_outdated BOOLEAN,
    ADD COLUMN pullreq_activity_code_comment_merge_base_sha TEXT,
    ADD COLUMN pullreq_activity_code_comment_source_sha TEXT,
    ADD COLUMN pullreq_activity_code_comment_path TEXT,
    ADD COLUMN pullreq_activity_code_comment_line_new INTEGER,
    ADD COLUMN pullreq_activity_code_comment_span_new INTEGER,
    ADD COLUMN pullreq_activity_code_comment_line_old INTEGER,
    ADD COLUMN pullreq_activity_code_comment_span_old INTEGER;