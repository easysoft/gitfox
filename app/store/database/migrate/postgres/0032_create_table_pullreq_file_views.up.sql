-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE pullreq_file_views (
 pullreq_file_view_pullreq_id INTEGER NOT NULL
,pullreq_file_view_principal_id INTEGER NOT NULL
,pullreq_file_view_path TEXT NOT NULL
,pullreq_file_view_sha TEXT NOT NULL
,pullreq_file_view_obsolete BOOLEAN NOT NULL
,pullreq_file_view_created BIGINT NOT NULL
,pullreq_file_view_updated BIGINT NOT NULL

-- for every pr and user at most one entry per file (existing enries are overwritten)
-- this index is also used for quick lookup of viewed files of a user for a given pr
,CONSTRAINT pk_pullreq_file_views PRIMARY KEY (pullreq_file_view_pullreq_id, pullreq_file_view_principal_id, pullreq_file_view_path)

,CONSTRAINT fk_pullreq_file_view_pullreq_id FOREIGN KEY (pullreq_file_view_pullreq_id)
    REFERENCES pullreqs (pullreq_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
,CONSTRAINT fk_pullreq_file_view_principal_id FOREIGN KEY (pullreq_file_view_principal_id)
    REFERENCES principals (principal_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);

-- this index is used to mark entries obsolete on branch update
CREATE INDEX pullreq_file_views_pullreq_id_file_path
    ON pullreq_file_views(pullreq_file_view_pullreq_id, pullreq_file_view_path);