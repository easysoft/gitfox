-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

ALTER TABLE repositories
    ADD COLUMN repo_importing BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN repo_importing_job_uid TEXT,
    ADD CONSTRAINT fk_repo_importing_job_uid
        FOREIGN KEY (repo_importing_job_uid)
        REFERENCES jobs(job_uid)
        ON DELETE SET NULL
        ON UPDATE NO ACTION;
