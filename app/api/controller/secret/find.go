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

package secret

import (
	"context"
	"fmt"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

func (c *Controller) Find(
	ctx context.Context,
	session *auth.Session,
	spaceRef string,
	identifier string,
) (*types.Secret, error) {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return nil, fmt.Errorf("failed to find space: %w", err)
	}
	err = apiauth.CheckSecret(ctx, c.authorizer, session, space.Path, identifier, enum.PermissionSecretView)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize: %w", err)
	}
	secret, err := c.secretStore.FindByIdentifier(ctx, space.ID, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to find secret: %w", err)
	}
	secret, err = Dec(c.encrypter, secret)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt secret: %w", err)
	}
	return secret, nil
}
