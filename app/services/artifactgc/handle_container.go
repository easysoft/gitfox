// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifactgc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/easysoft/gitfox/job"

	"github.com/rs/zerolog/log"
)

const (
	JobTypeArtifactContainerGC               = "gitfox:cleanup:artifact-container-garbage-collector"
	JobMaxDurationDArtifactContainerGC       = time.Minute * 15
	JobDefaultArtifactContainerRetentionTime = 0
)

type containerJob struct {
	svc *Service
}

func NewContainerJob(svc *Service) *containerJob {
	return &containerJob{
		svc: svc,
	}
}

func (j *containerJob) Handle(ctx context.Context, data string, _ job.ProgressReporter) (string, error) {
	log.Ctx(ctx).Info().Msg("start artifact container garbage collector job")

	var input Input
	var err error

	if data == "" {
		input = Input{
			RetentionTime: JobDefaultArtifactContainerRetentionTime,
		}
	} else {
		input, err = getJobInput(data)
		if err != nil {
			return "", err
		}
	}

	res, err := j.svc.GarbageCollectContainer(ctx, input.RetentionTime)
	if err != nil {
		return "", err
	}

	var result string
	if res.Count == 0 {
		result = "no container references need to be recycled"
	} else {
		result = fmt.Sprintf("remove container assets count: %d, size: %d", res.Count, res.Size)
	}

	log.Ctx(ctx).Info().Msg(result)
	return result, nil
}

func getJobInput(data string) (Input, error) {
	rawData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return Input{}, fmt.Errorf("failed to base64 decode job input: %w", err)
	}

	var input Input

	err = json.NewDecoder(bytes.NewReader(rawData)).Decode(&input)
	if err != nil {
		return Input{}, fmt.Errorf("failed to unmarshal job input json: %w", err)
	}

	return input, nil
}
