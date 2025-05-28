// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"encoding/json"
	"net/http"

	"github.com/easysoft/gitfox/app/api/render"
	artctl "github.com/easysoft/gitfox/app/artifact/controller"
)

func HandleSoftRemoveByNode(artCtl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		baseReq, err := artCtl.LoadBaseRequest(ctx, r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		in := new(artctl.ListNodeInfoRequest)
		err = json.NewDecoder(r.Body).Decode(in)
		if err != nil {
			render.BadRequestf(ctx, w, "Invalid request body: %s.", err)
			return
		}

		data, err := artCtl.SoftRemove(ctx, baseReq, in)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, data)
	}
}
