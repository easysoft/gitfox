-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE public_access_repo (
    public_access_repo_id INTEGER PRIMARY KEY,
    CONSTRAINT fk_public_access_repo_id FOREIGN KEY (public_access_repo_id)
        REFERENCES repositories (repo_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

CREATE TABLE public_access_space (
    public_access_space_id INTEGER PRIMARY KEY,
    CONSTRAINT fk_public_access_space_id FOREIGN KEY (public_access_space_id)
        REFERENCES spaces (space_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

INSERT INTO public_access_repo (
     public_access_repo_id
)
SELECT
     repo_id
FROM repositories
WHERE repo_is_public = TRUE;

ALTER TABLE repositories DROP COLUMN repo_is_public;

INSERT INTO public_access_space (
     public_access_space_id
)
SELECT
     space_id
FROM spaces
WHERE space_is_public = TRUE;

ALTER TABLE spaces DROP COLUMN space_is_public;
