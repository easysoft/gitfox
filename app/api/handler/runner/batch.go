// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package runner

import (
	"encoding/json"
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/runner"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"

	"github.com/drone/drone-go/drone"
)

func HandleBatch(runnerCtrl *runner.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		//session, _ := request.AuthSessionFrom(ctx)
		var lines []*drone.Line
		err := json.NewDecoder(r.Body).Decode(&lines)
		if err != nil {
			render.BadRequestf(ctx, w, "Invalid Request Body: %s.", err)
			return
		}

		stepId, err := request.PathParamAsPositiveInt64(r, request.PathParamStepNumber)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		err = runnerCtrl.Batch(ctx, stepId, lines)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
