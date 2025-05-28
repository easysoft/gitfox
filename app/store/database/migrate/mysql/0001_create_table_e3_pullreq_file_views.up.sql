-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE pullreq_file_views (
                                  pullreq_file_view_pullreq_id   INT NOT NULL,
                                  pullreq_file_view_principal_id INT NOT NULL,
                                  pullreq_file_view_path         VARCHAR(255) NOT NULL,
                                  pullreq_file_view_sha          VARCHAR(255) NOT NULL,
                                  pullreq_file_view_obsolete     BOOLEAN NOT NULL,
                                  pullreq_file_view_created      BIGINT NOT NULL,
                                  pullreq_file_view_updated      BIGINT NOT NULL,
                                  CONSTRAINT fk_pullreq_file_view_pullreq_id FOREIGN KEY (pullreq_file_view_pullreq_id) REFERENCES pullreqs(pullreq_id) ON DELETE CASCADE,
                                  CONSTRAINT fk_pullreq_file_view_principal_id FOREIGN KEY (pullreq_file_view_principal_id) REFERENCES principals(principal_id) ON DELETE CASCADE,
                                  CONSTRAINT pk_pullreq_file_views PRIMARY KEY (pullreq_file_view_pullreq_id, pullreq_file_view_principal_id, pullreq_file_view_path)
);

CREATE INDEX pullreq_file_views_pullreq_id_file_path
  ON pullreq_file_views (pullreq_file_view_pullreq_id, pullreq_file_view_path);
