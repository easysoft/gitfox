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

package space

import (
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/space"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
)

func HandleListPipelines(spaceCtrl *space.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)
		spaceRef, err := request.GetSpaceRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		filter, err := request.ParseListPipelinesFilterFromRequest(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
		}

		pipelines, totalCount, err := spaceCtrl.ListPipelines(ctx, session, spaceRef, filter)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.Pagination(r, w, filter.Page, filter.Size, int(totalCount))
		render.JSON(w, http.StatusOK, pipelines)
	}
}
