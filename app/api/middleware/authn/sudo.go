// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package authn

import (
	"net/http"
	"strconv"

	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/errors"
	store2 "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func Sudo(p store.PrincipalStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := hlog.FromRequest(r)

			sudoUser, found := parseSudoUser(r)
			if !found {
				next.ServeHTTP(w, r)
				return
			}

			session, ok := request.AuthSessionFrom(ctx)
			if !ok || session == nil {
				render.UserError(ctx, w, usererror.BadRequest("auth session is nil, use auth before sudo"))
				return
			}

			if !session.Principal.Admin || session.Principal.Blocked {
				render.UserError(ctx, w, usererror.BadRequest("only admin user can use sudo feature"))
				return
			}

			var sudoPrincipal *types.Principal

			sudoId, err := strconv.ParseInt(sudoUser, 10, 64)
			if err != nil {
				sudoPrincipal, err = p.FindByUID(ctx, sudoUser)
			} else {
				sudoPrincipal, err = p.Find(ctx, sudoId)
			}

			if err != nil {
				log.Warn().Err(err).Msg("find sudo user failed")

				if errors.Is(err, store2.ErrResourceNotFound) {
					render.UserError(ctx, w, usererror.BadRequestf("sudo user '%s' is not found", sudoUser))
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			//session.SudoUser = sudoPrincipal
			session.User = sudoPrincipal
			session.Principal = *sudoPrincipal

			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("sudo_uid", sudoPrincipal.UID)
			})

			next.ServeHTTP(w, r.WithContext(
				request.WithAuthSession(ctx, session),
			))
		})
	}
}

func parseSudoUser(r *http.Request) (sudoUid string, found bool) {
	sudoUid = r.URL.Query().Get(request.QueryParamSudo)
	if sudoUid != "" {
		found = true
		return
	}

	sudoUid = r.Header.Get(request.HeaderSudo)
	if sudoUid != "" {
		found = true
		return
	}
	return
}
