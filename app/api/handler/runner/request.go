// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package runner

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/runner"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/pipeline/manager"
)

func HandleRequest(runnerCtrl *runner.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		//session, _ := request.AuthSessionFrom(ctx)
		in := new(manager.Request)
		err := json.NewDecoder(r.Body).Decode(in)
		if err != nil {
			render.BadRequestf(ctx, w, "Invalid Request Body: %s.", err)
			return
		}

		stage, err := runnerCtrl.Request(ctx, in)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				w.WriteHeader(http.StatusNoContent)
				return
			} else if errors.Is(err, context.Canceled) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, stage)
	}
}
