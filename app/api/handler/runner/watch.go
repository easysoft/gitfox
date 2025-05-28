// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package runner

import (
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/runner"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
)

func HandleWatch(runnerCtrl *runner.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		//session, _ := request.AuthSessionFrom(ctx)
		execId, err := request.PathParamAsPositiveInt64(r, request.PathParamExecutionNumber)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		ok, err := runnerCtrl.Watch(ctx, execId)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		if !ok {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
