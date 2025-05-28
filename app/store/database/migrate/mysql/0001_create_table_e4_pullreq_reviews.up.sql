-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE pullreq_reviews (
                               pullreq_review_id         INT PRIMARY KEY AUTO_INCREMENT,
                               pullreq_review_created_by INT NOT NULL,
                               pullreq_review_created    BIGINT NOT NULL,
                               pullreq_review_updated    BIGINT NOT NULL,
                               pullreq_review_pullreq_id INT NOT NULL,
                               pullreq_review_decision   VARCHAR(255) NOT NULL,
                               pullreq_review_sha        VARCHAR(255) NOT NULL,
                               CONSTRAINT fk_pullreq_review_created_by FOREIGN KEY (pullreq_review_created_by) REFERENCES principals(principal_id),
                               CONSTRAINT fk_pullreq_review_pullreq_id FOREIGN KEY (pullreq_review_pullreq_id) REFERENCES pullreqs(pullreq_id) ON DELETE CASCADE
);

CREATE INDEX index_pullreq_review_pullreq_id
  ON pullreq_reviews (pullreq_review_pullreq_id);
