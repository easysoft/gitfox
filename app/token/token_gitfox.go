// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package token

import (
	"fmt"
	"time"

	"github.com/easysoft/gitfox/app/jwt"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/gotidy/ptr"
)

// CreateART is same as CreateSAT,
// without Principal convert to User
func CreateART(
	createdFor *types.Principal,
	lifetime *time.Duration,
) (string, error) {
	issuedAt := time.Now()

	var expiresAt *int64
	if lifetime != nil {
		expiresAt = ptr.Int64(issuedAt.Add(*lifetime).UnixMilli())
	}

	// create db entry first so we get the id.
	token := types.Token{
		Type:        enum.TokenTypeArtifact,
		PrincipalID: createdFor.ID,
		IssuedAt:    issuedAt.UnixMilli(),
		ExpiresAt:   expiresAt,
		CreatedBy:   createdFor.ID,
	}

	// create jwt token.
	jwtToken, err := jwt.GenerateForToken(&token, createdFor.Salt)
	if err != nil {
		return "", fmt.Errorf("failed to create jwt token: %w", err)
	}

	return jwtToken, nil
}
