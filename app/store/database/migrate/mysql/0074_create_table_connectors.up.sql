-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

-- Connectors table is not being used so can be dropped and recreated without
-- worrying about a migration
DROP TABLE IF EXISTS connectors;

CREATE TABLE connectors (
  connector_id INT AUTO_INCREMENT PRIMARY KEY,
  connector_identifier TEXT NOT NULL,
  connector_description TEXT NOT NULL,
  connector_type TEXT NOT NULL,
  connector_auth_type TEXT NOT NULL,
  connector_created_by INT NOT NULL,
  connector_space_id INT NOT NULL,
  connector_last_test_attempt INT NOT NULL,
  connector_last_test_error_msg TEXT NOT NULL,
  connector_last_test_status TEXT NOT NULL,
  connector_created INT NOT NULL,
  connector_updated INT NOT NULL,
  connector_version INT NOT NULL,
  connector_address TEXT,
  connector_insecure BOOLEAN,
  connector_username TEXT,
  connector_github_app_installation_id TEXT,
  connector_github_app_application_id TEXT,
  connector_region TEXT,
  connector_password INT,
  connector_token INT,
  connector_aws_key INT,
  connector_aws_secret INT,
  connector_github_app_private_key INT,
  connector_token_refresh INT,

  CONSTRAINT fk_connectors_space_id FOREIGN KEY (connector_space_id)
    REFERENCES spaces (space_id) ON UPDATE NO ACTION
    ON DELETE CASCADE,
  CONSTRAINT fk_connectors_created_by FOREIGN KEY (connector_created_by)
    REFERENCES principals (principal_id) ON UPDATE NO ACTION
    ON DELETE NO ACTION,
  CONSTRAINT fk_connectors_password FOREIGN KEY (connector_password)
    REFERENCES secrets (secret_id) ON UPDATE NO ACTION
    ON DELETE RESTRICT,
  CONSTRAINT fk_connectors_token FOREIGN KEY (connector_token)
    REFERENCES secrets (secret_id) ON UPDATE NO ACTION
    ON DELETE RESTRICT,
  CONSTRAINT fk_connectors_aws_key FOREIGN KEY (connector_aws_key)
    REFERENCES secrets (secret_id) ON UPDATE NO ACTION
    ON DELETE RESTRICT,
  CONSTRAINT fk_connectors_aws_secret FOREIGN KEY (connector_aws_secret)
    REFERENCES secrets (secret_id) ON UPDATE NO ACTION
    ON DELETE RESTRICT,
  CONSTRAINT fk_connectors_github_app_private_key FOREIGN KEY (connector_github_app_private_key)
    REFERENCES secrets (secret_id) ON UPDATE NO ACTION
    ON DELETE RESTRICT,
  CONSTRAINT fk_connectors_token_refresh FOREIGN KEY (connector_token_refresh)
    REFERENCES secrets (secret_id) ON UPDATE NO ACTION
    ON DELETE RESTRICT
);

CREATE UNIQUE INDEX unique_connector_lowercase_identifier ON connectors(connector_space_id, connector_identifier(255));
