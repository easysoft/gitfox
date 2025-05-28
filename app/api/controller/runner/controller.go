// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/easysoft/gitfox/app/auth/authz"
	"github.com/easysoft/gitfox/app/pipeline/manager"
	urlprovider "github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/livelog"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/client"
)

type Controller struct {
	tx          dbtx.Transactor
	authorizer  authz.Authorizer
	manager     manager.ExecutionManager
	urlProvider urlprovider.Provider
}

func NewController(
	tx dbtx.Transactor,
	authorizer authz.Authorizer,
	manager manager.ExecutionManager,
	urlProvider urlprovider.Provider,
) *Controller {
	return &Controller{
		tx:          tx,
		authorizer:  authorizer,
		manager:     manager,
		urlProvider: urlProvider,
	}
}

func (c *Controller) Request(ctx context.Context, args *manager.Request) (*drone.Stage, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	stage, err := c.manager.Request(ctxTimeout, args)
	if err != nil {
		return nil, err
	}
	return manager.ConvertToDroneStage(stage), nil
}

func (c *Controller) Accept(ctx context.Context, stageId int64, machine string) (*drone.Stage, error) {
	stage, err := c.manager.Accept(ctx, stageId, machine)
	if err != nil {
		return nil, err
	}
	return manager.ConvertToDroneStage(stage), nil
}

func (c *Controller) Details(ctx context.Context, stageID int64) (*client.Context, error) {
	details, err := c.manager.Details(ctx, stageID)
	if err != nil {
		return nil, err
	}

	return &client.Context{
		Build:   manager.ConvertToDroneBuild(details.Execution),
		Repo:    manager.ConvertToDroneRepo(details.Repo, false), // TODO: 需要确定是否需要传入true
		Stage:   manager.ConvertToDroneStage(details.Stage),
		Secrets: manager.ConvertToDroneSecrets(details.Secrets),
		Config:  manager.ConvertToDroneFile(details.Config),
		Netrc:   manager.ConvertToDroneNetrc(details.Netrc),
		System: &drone.System{
			Proto: c.urlProvider.GetAPIProto(ctx),
			Host:  c.urlProvider.GetAPIHostname(ctx),
		},
	}, nil
}

func (c *Controller) Update(ctx context.Context, stage *drone.Stage) error {
	var err error
	convertedStage := manager.ConvertFromDroneStage(stage)
	status := enum.ParseCIStatus(stage.Status)
	if status == enum.CIStatusPending || status == enum.CIStatusRunning {
		err = c.manager.BeforeStage(ctx, convertedStage)
	} else {
		err = c.manager.AfterStage(ctx, convertedStage)
	}
	*stage = *manager.ConvertToDroneStage(convertedStage)
	return err
}

func (c *Controller) UpdateStep(ctx context.Context, step *drone.Step) error {
	var err error
	convertedStep := manager.ConvertFromDroneStep(step)
	status := enum.ParseCIStatus(step.Status)
	if status == enum.CIStatusPending || status == enum.CIStatusRunning {
		err = c.manager.BeforeStep(ctx, convertedStep)
	} else {
		err = c.manager.AfterStep(ctx, convertedStep)
	}
	*step = *manager.ConvertToDroneStep(convertedStep)
	return err
}

// Watch watches for build cancellation requests.
func (c *Controller) Watch(ctx context.Context, executionID int64) (bool, error) {
	return c.manager.Watch(ctx, executionID)
}

func (c *Controller) Batch(ctx context.Context, step int64, lines []*drone.Line) error {
	for _, l := range lines {
		line := manager.ConvertFromDroneLine(l)
		err := c.manager.Write(ctx, step, line)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) Upload(ctx context.Context, step int64, l []*drone.Line) error {
	var buffer bytes.Buffer
	lines := []livelog.Line{}
	for _, line := range l {
		lines = append(lines, *manager.ConvertFromDroneLine(line))
	}
	out, err := json.Marshal(lines)
	if err != nil {
		return err
	}
	_, err = buffer.Write(out)
	if err != nil {
		return err
	}
	return c.manager.UploadLogs(ctx, step, &buffer)
}
