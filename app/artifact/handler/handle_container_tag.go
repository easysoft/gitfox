// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package handler

import (
	"net/http"

	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	artctl "github.com/easysoft/gitfox/app/artifact/controller"
)

func HandleListTag(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req, err := artCtrl.LoadContainerRequest(ctx, r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		data, err := artCtrl.ListTags(ctx, req)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		render.JSON(w, http.StatusOK, data)
	}
}
