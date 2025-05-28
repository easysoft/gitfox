-- Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
-- Use of this source code is covered by the following dual licenses:
-- (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
-- (2) Affero General Public License 3.0 (AGPL 3.0)
-- license that can be found in the LICENSE file.

CREATE TABLE jobs (
                    job_uid                  VARCHAR(255) NOT NULL,
                    job_created              BIGINT       NOT NULL,
                    job_updated              BIGINT       NOT NULL,
                    job_type                 VARCHAR(255) NOT NULL,
                    job_priority             INTEGER      NOT NULL,
                    job_data                 TEXT         NOT NULL,
                    job_result               TEXT         NOT NULL,
                    job_max_duration_seconds INTEGER      NOT NULL,
                    job_max_retries          INTEGER      NOT NULL,
                    job_state                VARCHAR(255) NOT NULL,
                    job_scheduled            BIGINT       NOT NULL,
                    job_total_executions     INTEGER,
                    job_run_by               VARCHAR(255) NOT NULL,
                    job_run_deadline         BIGINT,
                    job_run_progress         INTEGER      NOT NULL,
                    job_last_executed        BIGINT,
                    job_is_recurring         BOOLEAN      NOT NULL,
                    job_recurring_cron       VARCHAR(255) NOT NULL,
                    job_consecutive_failures INTEGER      NOT NULL,
                    job_last_failure_error   TEXT         NOT NULL,
                    job_group_id             VARCHAR(255) DEFAULT '' NOT NULL,
                    CONSTRAINT pk_jobs_uid PRIMARY KEY (job_uid)
);

CREATE INDEX job_group_id
  ON jobs (job_group_id);

CREATE INDEX jobs_last_executed
  ON jobs (job_last_executed);

CREATE INDEX jobs_run_deadline
  ON jobs (job_run_deadline);

CREATE INDEX jobs_scheduled
  ON jobs (job_scheduled);
