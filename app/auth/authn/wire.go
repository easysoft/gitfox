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
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	// ProvideAuthenticator by gitness is disabled
	//ProvideAuthenticator,
	ProvideAuthenticators,
)

func ProvideAuthenticator(
	config *types.Config,
	principalStore store.PrincipalStore,
	tokenStore store.TokenStore,
) Authenticator {
	return NewTokenAuthenticator(principalStore, tokenStore, config.Token.CookieName)
}

func ProvideAuthenticators(
	config *types.Config,
	principalStore store.PrincipalStore,
	tokenStore store.TokenStore,
) Authenticator {
	jwt := NewTokenAuthenticator(principalStore, tokenStore, config.Token.CookieName)
	_ = Register(AuthKindJWT, jwt)

	basic := NewBasicAuthenticator(principalStore, tokenStore)
	_ = Register(AuthKindBasic, basic)

	simpleJWT := NewSimpleTokenAuthenticator(principalStore, config.Token.CookieName)
	_ = Register(AuthKindSimpleJWT, simpleJWT)

	return jwt
}
