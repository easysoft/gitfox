// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package handler

import (
	"net/http"
	"time"

	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/auth/authn"
	"github.com/easysoft/gitfox/app/token"

	"github.com/rs/zerolog/log"
)

func HandleToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)
		if auth.IsAnonymousSession(session) {
			// basic auth failed, render 401 to client
			if _, _, isBasic := r.BasicAuth(); isBasic {
				container.RenderError(ctx, w, container.ErrUnauthorized)
				return
			}
		}

		data := make(map[string]interface{})

		var respToken string
		var err error

		switch session.Metadata.(type) {
		case *auth.MembershipMetadata, *auth.TokenMetadata, *auth.AccessPermissionMetadata:
			// authenticate by jwt token
			log.Ctx(ctx).Debug().Msg("pass current token as request token")
			respToken = authn.ExtractToken(r)
		default:
			// authenticate by basic auth token, without metadata
			// anonymous user also need a token
			// this token is for temporary use
			// if a request (pull|push) is too slow, increase the expired time
			expired := time.Minute * 30
			respToken, err = token.CreateART(&session.Principal, &expired)
			if err != nil {
				container.RenderError(ctx, w, container.ErrUnknown)
				return
			}
		}

		data["token"] = respToken
		render.JSON(w, http.StatusOK, data)
	}
}
