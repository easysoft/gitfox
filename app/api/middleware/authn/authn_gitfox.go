// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package authn

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/auth/authn"
	"github.com/easysoft/gitfox/app/url"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func Attempts(authenticators ...authn.Authenticator) func(http.Handler) http.Handler {
	return performAuthentications(authenticators, false)
}

func performAuthentications(
	authenticators []authn.Authenticator,
	required bool,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := hlog.FromRequest(r)

			var session *auth.Session
			var err error

			for _, a := range authenticators {
				session, err = a.Authenticate(r)
				if err == nil {
					break
				}
			}

			if err != nil {
				if !errors.Is(err, authn.ErrNoAuthData) {
					// log error to help with investigating any auth related errors
					log.Warn().Err(err).Msg("authentication failed")
				}

				if required {
					render.Unauthorized(ctx, w)
					return
				}
			}

			if session == nil {
				log.Info().Msg("No authentication data found, continue as anonymous")
				session = &auth.Session{
					Principal: auth.AnonymousPrincipal,
				}
			}

			// Update the logging context and inject principal in context
			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.
					Str("principal_uid", session.Principal.UID).
					Str("principal_type", string(session.Principal.Type)).
					Bool("principal_admin", session.Principal.Admin)
			})

			next.ServeHTTP(w, r.WithContext(
				request.WithAuthSession(ctx, session),
			))
		})
	}
}

// RequireContainerAccess tell client the next token-generate endpoint
func RequireContainerAccess(urlProvider url.Provider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			//if strings.Contains(r.UserAgent(), "conformance-tests") {
			//	next.ServeHTTP(w, r)
			//	return
			//}

			session, _ := request.AuthSessionFrom(ctx)
			if auth.IsAnonymousSession(session) {
				authAddress := fmt.Sprintf(`Bearer realm="%s",service="gitfox_registry",scope="*"`, urlProvider.GenerateRegistryURL("token").String())
				w.Header().Add("WWW-Authenticate", authAddress)
				container.RenderError(ctx, w, container.ErrUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
