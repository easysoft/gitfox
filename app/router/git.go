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

package router

import (
	"fmt"
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/repo"
	handlerrepo "github.com/easysoft/gitfox/app/api/handler/repo"
	middlewareauthn "github.com/easysoft/gitfox/app/api/middleware/authn"
	middlewareauthz "github.com/easysoft/gitfox/app/api/middleware/authz"
	"github.com/easysoft/gitfox/app/api/middleware/encode"
	"github.com/easysoft/gitfox/app/api/middleware/goget"
	"github.com/easysoft/gitfox/app/api/middleware/logging"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/auth/authn"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/check"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/hlog"
)

// NewGitHandler returns a new GitHandler.
func NewGitHandler(
	config *types.Config,
	urlProvider url.Provider,
	authenticator authn.Authenticator,
	repoCtrl *repo.Controller,
) http.Handler {
	// maxRepoDepth depends on config
	maxRepoDepth := check.MaxRepoPathDepth
	if !config.NestedSpacesEnabled {
		maxRepoDepth = 2
	}

	// Use go-chi router for inner routing.
	r := chi.NewRouter()

	// Apply common api middleware.
	r.Use(middleware.NoCache)
	r.Use(middleware.Recoverer)

	// configure logging middleware.
	r.Use(logging.URLHandler("http.url"))
	r.Use(hlog.MethodHandler("http.method"))
	r.Use(hlog.UserAgentHandler("http.user-agent"))
	r.Use(logging.HLogRequestIDHandler())
	r.Use(logging.HLogAccessLogHandler())

	// for now always attempt auth - enforced per operation.
	// use Attempts instead for multi authenticators support
	//r.Use(middlewareauthn.Attempt(authenticator))
	r.Use(middlewareauthn.Attempts(authenticator, authn.Get(authn.AuthKindBasic)))

	r.Route(fmt.Sprintf("/{%s}", request.PathParamRepoRef), func(r chi.Router) {
		r.Use(goget.Middleware(maxRepoDepth, repoCtrl, urlProvider))
		// routes that aren't coming from git
		r.Group(func(r chi.Router) {
			// redirect to repo (meant for UI, in case user navigates to clone url in browser)
			r.Get("/", handlerrepo.HandleGitRedirect(urlProvider))
		})

		// routes that are coming from git (where we block the usage of session tokens)
		r.Group(func(r chi.Router) {
			r.Use(middlewareauthz.BlockSessionToken)

			// smart protocol
			r.Post("/git-upload-pack", handlerrepo.HandleGitServicePack(
				enum.GitServiceTypeUploadPack, repoCtrl, urlProvider))
			r.Post("/git-receive-pack", handlerrepo.HandleGitServicePack(
				enum.GitServiceTypeReceivePack, repoCtrl, urlProvider))
			r.Get("/info/refs", handlerrepo.HandleGitInfoRefs(repoCtrl, urlProvider))

			// dumb protocol
			r.Get("/HEAD", stubGitHandler())
			r.Get("/objects/info/alternates", stubGitHandler())
			r.Get("/objects/info/http-alternates", stubGitHandler())
			r.Get("/objects/info/packs", stubGitHandler())
			r.Get("/objects/info/{file:[^/]*}", stubGitHandler())
			r.Get("/objects/{head:[0-9a-f]{2}}/{hash:[0-9a-f]{38}}", stubGitHandler())
			r.Get("/objects/pack/pack-{file:[0-9a-f]{40}}.pack", stubGitHandler())
			r.Get("/objects/pack/pack-{file:[0-9a-f]{40}}.idx", stubGitHandler())
		})
	})

	// wrap router in git path encoder.
	return encode.GitPathBefore(r)
}

func stubGitHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Seems like an asteroid destroyed the ancient git protocol"))
		w.WriteHeader(http.StatusBadGateway)
	}
}
