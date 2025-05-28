-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE gitspace_events CHANGE COLUMN geven_state geven_event VARCHAR(255);

-- TODO: 忽略这个字段
--- ALTER TABLE gitspace_events DROP COLUMN geven_space_id;

ALTER TABLE gitspaces ADD COLUMN gits_access_key TEXT;

ALTER TABLE gitspaces ADD COLUMN gits_access_type TEXT;

ALTER TABLE gitspaces ADD COLUMN gits_machine_user TEXT;
