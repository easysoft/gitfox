// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"

	"github.com/easysoft/gitfox/app/services/artifactgc"
	"github.com/easysoft/gitfox/types"
)

func (c *Controller) GarbageCollectContainer(ctx context.Context) (*types.JobUIDResponse, error) {
	in := artifactgc.Input{
		Kind:          artifactgc.KindContainer,
		RetentionTime: 0,
	}
	if err := c.gcSvc.Trigger(ctx, &in); err != nil {
		return nil, err
	}
	return &types.JobUIDResponse{UID: in.JobUID}, nil
}

func (c *Controller) GarbageCollectSoftRemove(ctx context.Context) (*types.JobUIDResponse, error) {
	in := artifactgc.Input{
		Kind:          artifactgc.KindSoftRemove,
		RetentionTime: 0,
	}
	if err := c.gcSvc.Trigger(ctx, &in); err != nil {
		return nil, err
	}
	return &types.JobUIDResponse{UID: in.JobUID}, nil
}
