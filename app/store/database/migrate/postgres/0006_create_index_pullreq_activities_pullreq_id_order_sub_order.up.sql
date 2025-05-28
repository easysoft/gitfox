-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE UNIQUE INDEX pullreq_activities_pullreq_id_order_sub_order
ON pullreq_activities(pullreq_activity_pullreq_id, pullreq_activity_order, pullreq_activity_sub_order);
