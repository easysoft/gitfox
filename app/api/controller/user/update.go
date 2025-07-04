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

package user

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/check"
	"github.com/easysoft/gitfox/types/enum"

	"golang.org/x/crypto/bcrypt"
)

// UpdateInput store infos to update an existing user.
type UpdateInput struct {
	Email        *string `json:"email"`
	Password     *string `json:"password"`
	DisplayName  *string `json:"display_name"`
	Source       *string `json:"source,omitempty"`
	PasswordHash *string `json:"password_hash,omitempty"`
}

// Update updates the provided user.
func (c *Controller) Update(ctx context.Context, session *auth.Session,
	userUID string, in *UpdateInput) (*types.User, error) {
	user, err := findUserFromUID(ctx, c.principalStore, userUID)
	if err != nil {
		return nil, err
	}

	// Ensure principal has required permissions on parent.
	if err = apiauth.CheckUser(ctx, c.authorizer, session, user, enum.PermissionUserEdit); err != nil {
		return nil, err
	}

	if err = c.sanitizeUpdateInput(in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if in.DisplayName != nil {
		user.DisplayName = *in.DisplayName
	}
	if in.Email != nil {
		user.Email = *in.Email
	}
	if in.Source != nil && *in.Source == string(enum.PrincipalSourceZentao) {
		if in.PasswordHash != nil {
			user.Password = string(*in.PasswordHash)
		} else if in.Password != nil {
			hashPass := md5.Sum([]byte(*in.Password))
			hash := hex.EncodeToString(hashPass[:])
			user.Password = string(hash)
		}
		user.Source = enum.PrincipalSourceZentao
	} else {
		if in.Password != nil {
			var hash []byte
			hash, err = hashPassword([]byte(*in.Password), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("failed to hash password: %w", err)
			}
			user.Password = string(hash)
		}
	}
	user.Updated = time.Now().UnixMilli()

	err = c.principalStore.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *Controller) sanitizeUpdateInput(in *UpdateInput) error {
	if in.Email != nil {
		*in.Email = strings.TrimSpace(*in.Email)
		if err := check.Email(*in.Email); err != nil {
			return err
		}
	}

	if in.DisplayName != nil {
		*in.DisplayName = strings.TrimSpace(*in.DisplayName)
		if err := check.DisplayName(*in.DisplayName); err != nil {
			return err
		}
	}

	if in.Source != nil &&
		*in.Source == string(enum.PrincipalSourceZentao) &&
		in.PasswordHash != nil {
		if err := check.ValidHashPassword(*in.PasswordHash); err != nil {
			return err
		}
	} else {
		if in.Password != nil {
			if err := check.Password(*in.Password); err != nil {
				return err
			}
		}
	}
	return nil
}
