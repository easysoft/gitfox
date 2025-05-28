// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"net/http"

	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	artctl "github.com/easysoft/gitfox/app/artifact/controller"
	"github.com/easysoft/gitfox/types"
)

func HandListVersions(artCtl *artctl.Controller, format types.ArtifactFormat) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		baseReq, err := artCtl.LoadBaseRequest(ctx, r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		filter, err := request.ParseArtifactVersionFilter(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		data, err := artCtl.ListVersions(ctx, baseReq, format, filter)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, data)
	}
}
