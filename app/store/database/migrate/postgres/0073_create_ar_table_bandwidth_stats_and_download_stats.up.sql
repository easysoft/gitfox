-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE download_stats
(
    download_stat_id                               SERIAL PRIMARY KEY,
    download_stat_artifact_id                      INTEGER NOT NULL
        CONSTRAINT fk_artifacts_artifact_id
            REFERENCES artifacts(artifact_id) ,
    download_stat_timestamp                        BIGINT NOT NULL,
    download_stat_created_at                       BIGINT NOT NULL,
    download_stat_updated_at                       BIGINT NOT NULL,
    download_stat_created_by                       INTEGER,
    download_stat_updated_by                       INTEGER
);

CREATE TABLE bandwidth_stats
(
    bandwidth_stat_id                               SERIAL PRIMARY KEY,
    bandwidth_stat_image_id                         INTEGER NOT NULL
        CONSTRAINT fk_images_image_id
            REFERENCES images(image_id) ,
    bandwidth_stat_timestamp                        BIGINT NOT NULL,
    bandwidth_stat_bytes                            BIGINT NOT NULL,
    bandwidth_stat_type                             TEXT NOT NULL,
    bandwidth_stat_created_at                       BIGINT NOT NULL,
    bandwidth_stat_updated_at                       BIGINT NOT NULL,
    bandwidth_stat_created_by                       INTEGER,
    bandwidth_stat_updated_by                       INTEGER
);