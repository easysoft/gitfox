-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE repositories ADD COLUMN repo_is_public BOOLEAN NOT NULL DEFAULT FALSE;

-- update repo public access
UPDATE repositories
SET repo_is_public = TRUE
WHERE repo_id IN (
    SELECT public_access_repo_id
    FROM public_access_repo
); 

ALTER TABLE spaces ADD COLUMN space_is_public BOOLEAN NOT NULL DEFAULT FALSE;

-- update space public access
UPDATE spaces
SET space_is_public = TRUE
WHERE space_id IN (
    SELECT public_access_space_id
    FROM public_access_space
); 

-- clear public_access
DROP TABLE public_access_repo;
DROP TABLE public_access_space