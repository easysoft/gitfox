-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE space_paths (
                           space_path_id         INT PRIMARY KEY AUTO_INCREMENT,
                           space_path_uid        VARCHAR(255) NOT NULL,
                           space_path_uid_unique VARCHAR(255) NOT NULL,
                           space_path_is_primary BOOLEAN DEFAULT NULL,
                           space_path_space_id   INT NOT NULL,
                           space_path_parent_id  INT,
                           space_path_created_by INT NOT NULL,
                           space_path_created    BIGINT NOT NULL,
                           space_path_updated    BIGINT NOT NULL,
                           CONSTRAINT fk_space_path_space_id FOREIGN KEY (space_path_space_id) REFERENCES spaces(space_id) ON DELETE CASCADE,
                           CONSTRAINT fk_space_path_parent_id FOREIGN KEY (space_path_parent_id) REFERENCES spaces(space_id) ON DELETE CASCADE,
                           CONSTRAINT fk_space_path_created_by FOREIGN KEY (space_path_created_by) REFERENCES principals(principal_id)
);

CREATE UNIQUE INDEX space_paths_space_id_is_primary ON space_paths (space_path_space_id, space_path_is_primary);

CREATE UNIQUE INDEX space_paths_uid_unique ON space_paths (space_path_parent_id, space_path_uid_unique);
