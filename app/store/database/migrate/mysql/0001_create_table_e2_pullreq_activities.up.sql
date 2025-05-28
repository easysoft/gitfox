-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE pullreq_activities (
                                  pullreq_activity_id                          INT PRIMARY KEY AUTO_INCREMENT,
                                  pullreq_activity_version                     BIGINT NOT NULL,
                                  pullreq_activity_created_by                  INT,
                                  pullreq_activity_created                     BIGINT NOT NULL,
                                  pullreq_activity_updated                     BIGINT NOT NULL,
                                  pullreq_activity_edited                      BIGINT NOT NULL,
                                  pullreq_activity_deleted                     BIGINT,
                                  pullreq_activity_parent_id                   INT,
                                  pullreq_activity_repo_id                     INT NOT NULL,
                                  pullreq_activity_pullreq_id                  INT NOT NULL,
                                  pullreq_activity_order                       INT NOT NULL,
                                  pullreq_activity_sub_order                   INT NOT NULL,
                                  pullreq_activity_reply_seq                   INT NOT NULL,
                                  pullreq_activity_type                        VARCHAR(255) NOT NULL,
                                  pullreq_activity_kind                        VARCHAR(255) NOT NULL,
                                  pullreq_activity_text                        TEXT NOT NULL,
                                  pullreq_activity_payload                     TEXT NOT NULL,
                                  pullreq_activity_metadata                    TEXT NOT NULL,
                                  pullreq_activity_resolved_by                 INT DEFAULT 0,
                                  pullreq_activity_resolved                    BIGINT,
                                  pullreq_activity_outdated                    BOOLEAN,
                                  pullreq_activity_code_comment_merge_base_sha VARCHAR(255),
                                  pullreq_activity_code_comment_source_sha     VARCHAR(255),
                                  pullreq_activity_code_comment_path           VARCHAR(255),
                                  pullreq_activity_code_comment_line_new       INT,
                                  pullreq_activity_code_comment_span_new       INT,
                                  pullreq_activity_code_comment_line_old       INT,
                                  pullreq_activity_code_comment_span_old       INT,
                                  CONSTRAINT fk_pullreq_activities_created_by FOREIGN KEY (pullreq_activity_created_by) REFERENCES principals(principal_id),
                                  CONSTRAINT fk_pullreq_activities_parent_id FOREIGN KEY (pullreq_activity_parent_id) REFERENCES pullreq_activities(pullreq_activity_id) ON DELETE CASCADE,
                                  CONSTRAINT fk_pullreq_activities_repo_id FOREIGN KEY (pullreq_activity_repo_id) REFERENCES repositories(repo_id) ON DELETE CASCADE,
                                  CONSTRAINT fk_pullreq_activities_pullreq_id FOREIGN KEY (pullreq_activity_pullreq_id) REFERENCES pullreqs(pullreq_id) ON DELETE CASCADE,
                                  CONSTRAINT fk_pullreq_activities_resolved_by FOREIGN KEY (pullreq_activity_resolved_by) REFERENCES principals(principal_id)
);

CREATE UNIQUE INDEX pullreq_activities_pullreq_id_order_sub_order
  ON pullreq_activities (pullreq_activity_pullreq_id, pullreq_activity_order, pullreq_activity_sub_order);
