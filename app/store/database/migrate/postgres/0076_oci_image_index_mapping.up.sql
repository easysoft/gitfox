-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE IF NOT EXISTS oci_image_index_mappings
(
    oci_mapping_id              SERIAL PRIMARY KEY,
    oci_mapping_parent_manifest_id   BIGINT NOT NULL,
    oci_mapping_child_digest    bytea NOT NULL,
    oci_mapping_created_at      BIGINT NOT NULL,
    oci_mapping_updated_at      BIGINT NOT NULL,
    oci_mapping_created_by      INTEGER NOT NULL,
    oci_mapping_updated_by      INTEGER NOT NULL,
    CONSTRAINT unique_oci_mapping_digests
        UNIQUE (oci_mapping_parent_manifest_id, oci_mapping_child_digest),
    CONSTRAINT fk_oci_mapping_registry_id
        FOREIGN KEY (oci_mapping_parent_manifest_id)
            REFERENCES manifests(manifest_id)
            ON DELETE CASCADE
)