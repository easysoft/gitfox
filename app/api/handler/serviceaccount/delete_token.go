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

package serviceaccount

import (
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/serviceaccount"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
)

// HandleDeleteToken returns an http.HandlerFunc that
// deletes a SAT token of a service account.
func HandleDeleteToken(saCrl *serviceaccount.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)
		saUID, err := request.GetServiceAccountUIDFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		tokentokenIdentifier, err := request.GetTokenIdentifierFromPath(r)
		if err != nil {
			render.BadRequest(ctx, w)
			return
		}

		err = saCrl.DeleteToken(ctx, session, saUID, tokentokenIdentifier)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.DeleteSuccessful(w)
	}
}
