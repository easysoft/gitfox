// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
)

type BaseReq struct {
	spaceName string

	session *auth.Session
	view    *adapter.ViewDescriptor
}

func (b *BaseReq) Session() *auth.Session {
	return b.session
}

func (b *BaseReq) Space() *types.Space {
	return b.view.Space
}

func (c *Controller) LoadBaseRequest(ctx context.Context, r *http.Request) (*BaseReq, error) {
	spaceName, err := request.GetSpaceRefFromPath(r)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("could not get space name")
		return nil, adapter.ErrMissPathField.WithDetail(fmt.Sprintf("require path name: %s", request.PathParamSpaceRef))
	}

	reqView, err := c.ParseSpaceView(ctx, spaceName)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("error parsing view")
		return nil, container.ErrUnknown.WithDetail(err.Error())
	}

	session, _ := request.AuthSessionFrom(ctx)
	return &BaseReq{
		spaceName: spaceName,
		session:   session,
		view:      reqView,
	}, nil
}
