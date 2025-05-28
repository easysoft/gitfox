// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package authn

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/hlog"
	"golang.org/x/crypto/bcrypt"
)

var _ Authenticator = (*BasicAuthenticator)(nil)

// BasicAuthenticator uses the provided JWT to authenticate the caller.
type BasicAuthenticator struct {
	principalStore store.PrincipalStore
	tokenStore     store.TokenStore
}

func NewBasicAuthenticator(
	principalStore store.PrincipalStore,
	tokenStore store.TokenStore,
) *BasicAuthenticator {
	return &BasicAuthenticator{
		principalStore: principalStore,
		tokenStore:     tokenStore,
	}
}

func (a *BasicAuthenticator) Authenticate(r *http.Request) (*auth.Session, error) {
	ctx := r.Context()
	log := hlog.FromRequest(r)
	log.Debug().Msg("trigger BasicAuthenticator")

	userName, passwd, ok := r.BasicAuth()
	if !ok {
		return nil, ErrNoAuthData
	}
	user, err := a.principalStore.FindUserByUID(ctx, userName)
	if errors.Is(err, gitfox_store.ErrResourceNotFound) {
		return nil, usererror.ErrUnauthorized
	}

	if user.Blocked {
		log.Debug().Str("user_uid", user.UID).Msg("user is blocked")
		return nil, usererror.ErrUnauthorized
	}

	if user.Source == enum.PrincipalSourceZentao {
		hash := md5.Sum([]byte(passwd))
		hashedPassword := hex.EncodeToString(hash[:])
		if hashedPassword != user.Password {
			log.Debug().Err(fmt.Errorf("password hash no invalid")).
				Str("user_uid", user.UID).
				Msg("invalid zentao password")
			return nil, usererror.ErrUnauthorized
		}
	} else {
		err := bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(passwd),
		)
		if err != nil {
			log.Debug().Err(err).
				Str("user_uid", user.UID).
				Msg("invalid password")
			return nil, usererror.ErrUnauthorized
		}
	}

	log.Debug().Msg("matched basic auth")
	return &auth.Session{
		Principal: *user.ToPrincipal(),
		Metadata:  &auth.EmptyMetadata{},
		User:      user.ToPrincipal(),
	}, nil
}
