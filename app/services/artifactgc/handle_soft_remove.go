// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifactgc

import (
	"context"
	"fmt"
	"time"

	"github.com/easysoft/gitfox/job"

	"github.com/rs/zerolog/log"
)

const (
	JobTypeArtifactHardRemove         = "gitfox:cleanup:artifact-soft-remove"
	JobMaxDurationDArtifactSoftRemove = time.Minute * 5
	jobDefaultRetentionSoftRemove     = 0
)

type softRemoveJob struct {
	svc *Service
}

func NewSoftRemoveJob(svc *Service) *softRemoveJob {
	return &softRemoveJob{
		svc: svc,
	}
}

func (j *softRemoveJob) Handle(ctx context.Context, data string, _ job.ProgressReporter) (string, error) {
	log.Ctx(ctx).Info().Msg("start artifact recycle soft-removed job")

	var input Input
	var err error

	if data == "" {
		input = Input{
			RetentionTime: jobDefaultRetentionSoftRemove,
		}
	} else {
		input, err = getJobInput(data)
		if err != nil {
			return "", err
		}
	}

	res, err := j.svc.GarbageCollectSoftRemove(ctx, input.RetentionTime)
	if err != nil {
		return "", err
	}

	var result string
	if res.Count == 0 {
		result = "no soft-removed asset found"
	} else {
		result = fmt.Sprintf("recycle soft-removed asset count: %d, size: %d", res.Count, res.Size)
	}

	log.Ctx(ctx).Info().Msg(result)
	return result, nil
}
