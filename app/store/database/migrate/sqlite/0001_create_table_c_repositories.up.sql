-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE repositories (
 repo_id                INTEGER PRIMARY KEY AUTOINCREMENT
,repo_version           INTEGER NOT NULL DEFAULT 0
,repo_parent_id         INTEGER NOT NULL
,repo_uid               TEXT NOT NULL
,repo_description       TEXT
,repo_is_public         BOOLEAN NOT NULL
,repo_created_by        INTEGER NOT NULL
,repo_created           BIGINT NOT NULL
,repo_updated           BIGINT NOT NULL
,repo_git_uid           TEXT NOT NULL
,repo_default_branch    TEXT NOT NULL
,repo_fork_id           INTEGER
,repo_pullreq_seq       INTEGER NOT NULL
,repo_num_forks         INTEGER NOT NULL
,repo_num_pulls         INTEGER NOT NULL
,repo_num_closed_pulls  INTEGER NOT NULL
,repo_num_open_pulls    INTEGER NOT NULL
,repo_num_merged_pulls  INTEGER NOT NULL

,UNIQUE(repo_git_uid)

,CONSTRAINT fk_repo_parent_id FOREIGN KEY (repo_parent_id)
    REFERENCES spaces (space_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);
