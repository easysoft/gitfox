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

package authn

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/jwt"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	gojwt "github.com/golang-jwt/jwt"
	"github.com/rs/zerolog/log"
)

var _ Authenticator = (*SimpleJWTAuthenticator)(nil)

// SimpleJWTAuthenticator uses the provided JWT to authenticate the caller.
type SimpleJWTAuthenticator struct {
	cookieName     string
	principalStore store.PrincipalStore
	tokenStore     store.TokenStore
}

func NewSimpleTokenAuthenticator(
	principalStore store.PrincipalStore,
	cookieName string,
) *SimpleJWTAuthenticator {
	return &SimpleJWTAuthenticator{
		cookieName:     cookieName,
		principalStore: principalStore,
	}
}

func (a *SimpleJWTAuthenticator) Authenticate(r *http.Request) (*auth.Session, error) {
	ctx := r.Context()
	str := extractToken(r, a.cookieName)

	if len(str) == 0 {
		return nil, ErrNoAuthData
	}

	var principal *types.Principal
	var err error
	claims := &jwt.Claims{}
	parsed, err := gojwt.ParseWithClaims(str, claims, func(_ *gojwt.Token) (interface{}, error) {
		if claims.PrincipalID == auth.AnonymousPrincipal.ID {
			principal = &auth.AnonymousPrincipal
		} else {
			principal, err = a.principalStore.Find(ctx, claims.PrincipalID)
			if err != nil {
				return nil, fmt.Errorf("failed to get principal for token: %w", err)
			}
			if principal.Blocked {
				return nil, fmt.Errorf("user is blocked")
			}
		}
		return []byte(principal.Salt), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing of JWT claims failed: %w", err)
	}

	if !parsed.Valid {
		return nil, errors.New("parsed JWT token is invalid")
	}

	if _, ok := parsed.Method.(*gojwt.SigningMethodHMAC); !ok {
		return nil, errors.New("invalid HMAC signature for JWT")
	}

	var metadata auth.Metadata
	switch {
	case claims.Token != nil:
		if claims.Token.Type != enum.TokenTypeArtifact {
			return nil, errors.New("invalid token type")
		}
		metadata = &auth.TokenMetadata{
			TokenType: claims.Token.Type,
			TokenID:   claims.Token.ID, // always zero
		}
	default:
		return nil, fmt.Errorf("jwt is missing sub-claims")
	}

	log.Ctx(ctx).Debug().Msg("matched simple jwt auth")
	return &auth.Session{
		Principal: *principal,
		Metadata:  metadata,
		User:      principal,
	}, nil
}
