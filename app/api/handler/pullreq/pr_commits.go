// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pullreq

import (
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/pullreq"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/types"
)

// HandleCommits returns commits for PR.
func HandleCommits(pullreqCtrl *pullreq.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)

		repoRef, err := request.GetRepoRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		pullreqNumber, err := request.GetPullReqNumberFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		filter := &types.PaginationFilter{
			Page:  request.ParsePage(r),
			Limit: request.ParseLimit(r),
		}

		// gitref is Head branch in this case
		commits, err := pullreqCtrl.Commits(ctx, session, repoRef, pullreqNumber, filter)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		// TODO: get last page indicator explicitly - current check is wrong in case len % limit == 0
		isLastPage := len(commits) < filter.Limit
		render.PaginationNoTotal(r, w, filter.Page, filter.Limit, isLastPage)
		render.JSON(w, http.StatusOK, commits)
	}
}
