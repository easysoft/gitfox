-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE pullreq_reviews (
pullreq_review_id SERIAL PRIMARY KEY
,pullreq_review_created_by INTEGER NOT NULL
,pullreq_review_created BIGINT NOT NULL
,pullreq_review_updated BIGINT NOT NULL
,pullreq_review_pullreq_id INTEGER NOT NULL
,pullreq_review_decision TEXT NOT NULL
,pullreq_review_sha TEXT NOT NULL
,CONSTRAINT fk_pullreq_review_created_by FOREIGN KEY (pullreq_review_created_by)
    REFERENCES principals (principal_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
,CONSTRAINT fk_pullreq_review_pullreq_id FOREIGN KEY (pullreq_review_pullreq_id)
    REFERENCES pullreqs (pullreq_id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
);