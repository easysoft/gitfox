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

package githook

import (
	"encoding/json"
	"net/http"

	controllergithook "github.com/easysoft/gitfox/app/api/controller/githook"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/types"
)

// HandleUpdate returns a handler function that handles update git hooks.
func HandleUpdate(
	githookCtrl *controllergithook.Controller,
	git git.Interface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)

		in := types.GithookUpdateInput{}
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			render.BadRequestf(ctx, w, "Invalid Request Body: %s.", err)
			return
		}

		// Harness doesn't require any custom git connector.
		out, err := githookCtrl.Update(ctx, git, session, in)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, out)
	}
}
