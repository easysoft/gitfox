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
	"compress/gzip"
	"errors"
	"fmt"
	"net/http"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/api/controller/repo"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/url"
	"github.com/easysoft/gitfox/git/api"
	gitfox_request "github.com/easysoft/gitfox/pkg/context/request"
	"github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
)

// HandleGitServicePack handles the service pack part of git's smart http protocol (receive-/upload-pack).
func HandleGitServicePack(
	service enum.GitServiceType,
	repoCtrl *repo.Controller,
	urlProvider url.Provider,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = gitfox_request.WithContext(ctx, r)
		session, _ := request.AuthSessionFrom(ctx)
		repoRef, err := request.GetRepoRefFromPath(r)
		log.Ctx(ctx).Info().Msg("repoRef: " + repoRef)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("request.GetRepoRefFromPath")
			render.TranslatedUserError(ctx, w, err)
			return
		}

		contentEncoding := request.GetContentEncodingFromHeadersOrDefault(r, "")
		gitProtocol := request.GetGitProtocolFromHeadersOrDefault(r, "")

		// Handle GZIP.
		dataReader := r.Body
		if contentEncoding == "gzip" {
			gzipReader, err := gzip.NewReader(dataReader)
			if err != nil {
				render.TranslatedUserError(ctx, w, fmt.Errorf("failed to create new gzip reader: %w", err))
				return
			}
			defer func() {
				if cErr := gzipReader.Close(); cErr != nil {
					log.Ctx(ctx).Warn().Err(cErr).Msg("failed to close the gzip reader")
				}
			}()
			dataReader = gzipReader
		}

		// Clients MUST NOT reuse or revalidate a cached response.
		// Servers MUST include sufficient Cache-Control headers to prevent caching of the response.
		// https://git-scm.com/docs/http-protocol
		render.NoCache(w)
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", service))

		err = repoCtrl.GitServicePack(ctx, session, repoRef, api.ServicePackOptions{
			Service:      service,
			StatelessRPC: true,
			Stdout:       w,
			Stdin:        dataReader,
			Protocol:     gitProtocol,
		})
		if errors.Is(err, apiauth.ErrNotAuthorized) && auth.IsAnonymousSession(session) {
			renderBasicAuth(ctx, w, urlProvider)
			return
		}
		if errors.Is(err, store.ErrRequireFitClient) {
			w.Header().Set("Content-Type", "text/plain;charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("repoCtrl.GitServicePack")
			render.TranslatedUserError(ctx, w, err)
			return
		}
	}
}
