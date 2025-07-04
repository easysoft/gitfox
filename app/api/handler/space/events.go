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
	"context"
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/space"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"

	"github.com/rs/zerolog/log"
)

// HandleEvents returns a http.HandlerFunc that watches for events on a space.
func HandleEvents(appCtx context.Context, spaceCtrl *space.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)

		spaceRef, err := request.GetSpaceRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		chEvents, chErr, sseCancel, err := spaceCtrl.Events(ctx, session, spaceRef)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		defer func() {
			if err := sseCancel(ctx); err != nil {
				log.Ctx(ctx).Err(err).Msgf("failed to cancel sse stream for space '%s'", spaceRef)
			}
		}()

		render.StreamSSE(ctx, w, appCtx.Done(), chEvents, chErr)
	}
}
