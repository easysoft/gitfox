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

package repo

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/api/controller/repo"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/url"
	gitfox_request "github.com/easysoft/gitfox/pkg/context/request"

	"github.com/rs/zerolog/log"
)

// HandleGitInfoRefs handles the info refs part of git's smart http protocol.
func HandleGitInfoRefs(repoCtrl *repo.Controller, urlProvider url.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = gitfox_request.WithContext(ctx, r)
		session, _ := request.AuthSessionFrom(ctx)
		repoRef, err := request.GetRepoRefFromPath(r)
		log.Ctx(ctx).Debug().Any("repoRef", repoRef).Msg("HandleGitInfoRefs")
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		gitProtocol := request.GetGitProtocolFromHeadersOrDefault(r, "")
		service, err := request.GetGitServiceTypeFromQuery(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		log.Ctx(ctx).Debug().Any("service", service).Msg("HandleGitInfoRefs")
		// Clients MUST NOT reuse or revalidate a cached response.
		// Servers MUST include sufficient Cache-Control headers to prevent caching of the response.
		// https://git-scm.com/docs/http-protocol
		render.NoCache(w)
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", service))

		err = repoCtrl.GitInfoRefs(ctx, session, repoRef, service, gitProtocol, w)
		if errors.Is(err, apiauth.ErrNotAuthorized) && auth.IsAnonymousSession(session) {
			renderBasicAuth(ctx, w, urlProvider)
			return
		}
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		log.Ctx(ctx).Debug().Any("end", "done").Msg("HandleGitInfoRefs")
	}
}

// renderBasicAuth renders a response that indicates that the client (GIT) requires basic authentication.
// This is required in order to tell git CLI to query user credentials.
func renderBasicAuth(ctx context.Context, w http.ResponseWriter, urlProvider url.Provider) {
	// Git doesn't seem to handle "realm" - so it doesn't seem to matter for basic user CLI interactions.
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, urlProvider.GetAPIHostname(ctx)))
	w.WriteHeader(http.StatusUnauthorized)
}
