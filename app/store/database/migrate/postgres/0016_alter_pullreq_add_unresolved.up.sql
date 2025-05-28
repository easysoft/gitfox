-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE pullreqs ADD COLUMN pullreq_unresolved_count INTEGER NOT NULL DEFAULT 0;

WITH unresolved_counts AS (
    SELECT
        pullreq_activity_pullreq_id AS "unresolved_pullreq_id",
        COUNT(*) AS "unresolved_count"
    FROM pullreq_activities
    WHERE
        pullreq_activity_sub_order = 0 AND
        pullreq_activity_resolved IS NULL AND
        pullreq_activity_deleted IS NULL AND
        pullreq_activity_kind <> 'system'
    GROUP BY pullreq_activity_pullreq_id
)
UPDATE pullreqs
SET pullreq_unresolved_count = unresolved_counts.unresolved_count
FROM unresolved_counts
WHERE pullreq_id = unresolved_pullreq_id;
