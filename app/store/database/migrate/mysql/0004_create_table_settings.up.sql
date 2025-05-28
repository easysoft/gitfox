-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE settings (
  setting_id       INT AUTO_INCREMENT PRIMARY KEY,
  setting_space_id INT,
  setting_repo_id  INT,
  setting_key      VARCHAR(100) NOT NULL,
  setting_value    TEXT,
  CONSTRAINT fk_settings_space_id FOREIGN KEY (setting_space_id)
    REFERENCES spaces (space_id)
    ON DELETE CASCADE,
  CONSTRAINT fk_settings_repo_id FOREIGN KEY (setting_repo_id)
    REFERENCES repositories (repo_id)
    ON DELETE CASCADE
);

create INDEX idx_setting_key ON settings (setting_key);
