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

package template

import (
	"encoding/json"
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/template"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/paths"
	"github.com/easysoft/gitfox/types/enum"
)

func HandleUpdate(templateCtrl *template.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)

		in := new(template.UpdateInput)
		err := json.NewDecoder(r.Body).Decode(in)
		if err != nil {
			render.BadRequestf(ctx, w, "Invalid Request Body: %s.", err)
			return
		}

		templateRef, err := request.GetTemplateRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		spaceRef, templateIdentifier, err := paths.DisectLeaf(templateRef)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		resolverType, err := request.GetTemplateTypeFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
		}

		resolverTypeEnum, err := enum.ParseResolverType(resolverType)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		template, err := templateCtrl.Update(ctx, session, spaceRef, templateIdentifier,
			resolverTypeEnum, in)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, template)
	}
}
