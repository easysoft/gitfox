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

package keywordsearch

import (
	"encoding/json"
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/keywordsearch"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/types"
)

// HandleSearch returns keyword search results on repositories.
func HandleSearch(ctrl *keywordsearch.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)

		searchInput := types.SearchInput{}
		err := json.NewDecoder(r.Body).Decode(&searchInput)
		if err != nil {
			render.BadRequestf(ctx, w, "invalid Request Body: %s.", err)
			return
		}

		result, err := ctrl.Search(ctx, session, searchInput)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, result)
	}
}
