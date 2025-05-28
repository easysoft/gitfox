// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package authn

import "fmt"

type authKind string

const (
	AuthKindJWT       = "jwt"
	AuthKindBasic     = "basic"
	AuthKindSimpleJWT = "simple-jwt"
	AuthKindApiKey    = "apiKey"
)

var _authMethods = make(map[authKind]Authenticator)

func Register(kind authKind, authenticator Authenticator) error {
	if _, ok := _authMethods[kind]; ok {
		return fmt.Errorf("conflict authenticator '%s'", kind)
	}
	_authMethods[kind] = authenticator
	return nil
}

func Get(kind authKind) Authenticator {
	return _authMethods[kind]
}
