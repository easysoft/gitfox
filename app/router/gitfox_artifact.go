// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/easysoft/gitfox/app/api/middleware/address"
	middlewareauthn "github.com/easysoft/gitfox/app/api/middleware/authn"
	middlewarelogging "github.com/easysoft/gitfox/app/api/middleware/logging"
	"github.com/easysoft/gitfox/app/api/request"
	artctl "github.com/easysoft/gitfox/app/artifact/controller"
	"github.com/easysoft/gitfox/app/artifact/handler"
	"github.com/easysoft/gitfox/app/auth/authn"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/logging"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/hlog"
)

const (
	ArtifactMount          = "/_artifacts"
	ArtifactContainerMount = "/v2"
)

type ArtifactRouter struct {
	handler http.Handler
}

func NewArtifactRouter(handler http.Handler) *ArtifactRouter {
	return &ArtifactRouter{handler: handler}
}

func (r *ArtifactRouter) Handle(w http.ResponseWriter, req *http.Request) {
	req = req.WithContext(logging.NewContext(req.Context(), WithLoggingRouter("artifact")))
	r.handler.ServeHTTP(w, req)
}

func (r *ArtifactRouter) IsEligibleTraffic(req *http.Request) bool {
	p := req.URL.Path
	return strings.HasPrefix(p, ArtifactMount) || strings.HasPrefix(p, ArtifactContainerMount)
}

func (r *ArtifactRouter) Name() string {
	return "artifact"
}

// ArtifactHandler is an abstraction of a http handler that handles Artifact calls.
type ArtifactHandler interface {
	http.Handler
}

func NewArtifactHandler(
	appCtx context.Context,
	urlProvider url.Provider,
	config *types.Config,
	authenticator authn.Authenticator,
	artCtrl *artctl.Controller,
	artStore store.ArtifactStore,
	repoStore store.RepoStore,
	fileStore storage.ContentStorage,
) ArtifactHandler {
	// Use go-chi router for inner routing.
	r := chi.NewRouter()

	// Apply common api middleware.
	r.Use(middleware.NoCache)
	r.Use(middleware.Recoverer)

	// configure logging middleware.
	r.Use(hlog.URLHandler("http.url"))
	r.Use(hlog.MethodHandler("http.method"))
	r.Use(middlewarelogging.HLogRequestIDHandler())
	r.Use(middlewarelogging.HLogAccessLogHandler())
	r.Use(address.Handler("", ""))

	// for now always attempt auth - enforced per operation.
	r.Use(middlewareauthn.Attempts(
		authn.Get(authn.AuthKindJWT),       // allow web, serviceaccount in pipeline
		authn.Get(authn.AuthKindSimpleJWT), // allow temporary jwt token
		authn.Get(authn.AuthKindBasic),     // allow user/password
	))

	r.Route(ArtifactMount, func(r chi.Router) {
		r.Route(fmt.Sprintf("/{%s}", request.PathParamSpaceRef), func(r chi.Router) {
			setupArtifactHelm(r, appCtx, artStore, artCtrl)
			setupArtifactRaw(r, appCtx, artStore, artCtrl)
		})
	})

	r.Route(ArtifactContainerMount, func(r chi.Router) {
		// apiBase is default authorize endpoint
		r.With(middlewareauthn.RequireContainerAccess(urlProvider)).Get("/", handler.HandleAPIBase())
		r.Get("/token", handler.HandleToken())
		r.Route("/{space}", func(r chi.Router) {
			// write operations require a login session
			r.With(middlewareauthn.RequireContainerAccess(urlProvider)).Route("/", func(r chi.Router) {
				r.Delete("/{name}/blobs/{digest}", handler.HandleBlobDelete(artCtrl))

				r.Post("/{name}/blobs/uploads/", handler.HandleBlobUploadStart(artCtrl))
				r.Patch("/{name}/blobs/uploads/{uuid}", handler.HandleBlobUploadPatch(artCtrl))
				r.Put("/{name}/blobs/uploads/{uuid}", handler.HandleBlobUploadFinish(artCtrl))
				r.Delete("/{name}/blobs/uploads/{uuid}", handler.HandleBlobUploadCancel(artCtrl))

				r.Put("/{name}/manifests/{reference}", handler.HandleManifestPut(artCtrl))
				r.Delete("/{name}/manifests/{reference}", handler.HandleManifestDelete(artCtrl))
			})

			r.Head("/{name}/blobs/{digest}", handler.HandleBlobHead(artCtrl))
			r.Get("/{name}/blobs/{digest}", handler.HandleBlobGet(artCtrl))

			r.Get("/{name}/blobs/uploads/{uuid}", handler.HandleBlobUploadStatus(artCtrl))

			r.Head("/{name}/manifests/{reference}", handler.HandleManifestHead(artCtrl))
			r.Get("/{name}/manifests/{reference}", handler.HandleManifestGet(artCtrl))

			r.Get("/{name}/tags/list", handler.HandleListTag(artCtrl))
		})
	})

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.ServeHTTP(w, req)
	})
}

func setupArtifactHelm(r chi.Router, appCtx context.Context, artStore store.ArtifactStore, artCtrl *artctl.Controller) {
	r.Route("/helm", func(r chi.Router) {
		r.Post("/api/charts", handler.HandHelmUpload(artCtrl))
		r.Get("/*", handler.HandHelmDownload(artStore, artCtrl))
	})
}

func setupArtifactRaw(r chi.Router, appCtx context.Context, artStore store.ArtifactStore, artCtrl *artctl.Controller) {
	r.Route("/raw", func(r chi.Router) {
		r.Post("/upload", handler.HandRawUpload(artCtrl))
		r.Get("/*", handler.HandRawDownload(artStore, artCtrl))
	})
}
