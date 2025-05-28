-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE checks (
 check_id INTEGER PRIMARY KEY AUTOINCREMENT
,check_created_by INTEGER NOT NULL
,check_created BIGINT NOT NULL
,check_updated BIGINT NOT NULL
,check_repo_id INTEGER NOT NULL
,check_commit_sha TEXT NOT NULL
,check_type TEXT NOT NULL
,check_uid TEXT NOT NULL
,check_status TEXT NOT NULL
,check_summary TEXT NOT NULL
,check_link TEXT NOT NULL
,check_payload TEXT NOT NULL
,check_metadata TEXT NOT NULL
,CONSTRAINT fk_check_created_by FOREIGN KEY (check_created_by)
    REFERENCES principals (principal_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
,CONSTRAINT fk_check_repo_id FOREIGN KEY (check_repo_id)
    REFERENCES repositories (repo_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);

CREATE UNIQUE INDEX checks_repo_id_commit_sha_uid
    ON checks(check_repo_id, check_commit_sha, check_uid);

CREATE INDEX checks_repo_id_created
    ON checks(check_repo_id, check_created);

CREATE TABLE reqchecks (
 reqcheck_id INTEGER PRIMARY KEY AUTOINCREMENT
,reqcheck_created_by INTEGER NOT NULL
,reqcheck_created BIGINT NOT NULL
,reqcheck_repo_id INTEGER NOT NULL
,reqcheck_branch_pattern TEXT NOT NULL
,reqcheck_check_uid TEXT NOT NULL
,CONSTRAINT fk_check_created_by FOREIGN KEY (reqcheck_created_by)
    REFERENCES principals (principal_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
,CONSTRAINT fk_check_repo_id FOREIGN KEY (reqcheck_repo_id)
    REFERENCES repositories (repo_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);

CREATE INDEX reqchecks_repo_id
    ON reqchecks(reqcheck_repo_id);
