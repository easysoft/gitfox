-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE usergroup_reviewers (
  usergroup_reviewer_pullreq_id INT NOT NULL,
  usergroup_reviewer_usergroup_id INT NOT NULL,
  usergroup_reviewer_created_by INT NOT NULL,
  usergroup_reviewer_created BIGINT NOT NULL,
  usergroup_reviewer_updated BIGINT NOT NULL,
  usergroup_reviewer_repo_id INT NOT NULL,
  PRIMARY KEY (usergroup_reviewer_pullreq_id, usergroup_reviewer_usergroup_id),
  FOREIGN KEY (usergroup_reviewer_pullreq_id)
    REFERENCES pullreqs (pullreq_id) ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (usergroup_reviewer_usergroup_id)
    REFERENCES usergroups (usergroup_id) ON UPDATE NO ACTION ON DELETE CASCADE,
  FOREIGN KEY (usergroup_reviewer_created_by)
    REFERENCES principals (principal_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  FOREIGN KEY (usergroup_reviewer_repo_id)
    REFERENCES repositories (repo_id) ON UPDATE NO ACTION ON DELETE CASCADE
);
