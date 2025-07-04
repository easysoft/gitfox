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

	"github.com/dchest/uniuri"
	"golang.org/x/crypto/bcrypt"
)

// CreateInput is the input used for create operations.
// On purpose don't expose admin, has to be enabled explicitly.
type CreateInput struct {
	UID          string `json:"uid"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	Password     string `json:"password,omitempty"`
	PasswordHash string `json:"password_hash,omitempty"`
	Source       string `json:"source,omitempty"`
}

// Create creates a new user.
func (c *Controller) Create(ctx context.Context, session *auth.Session, in *CreateInput) (*types.User, error) {
	// Ensure principal has required permissions (user is global, no explicit resource)
	scope := &types.Scope{}
	resource := &types.Resource{
		Type: enum.ResourceTypeUser,
	}
	if err := apiauth.Check(ctx, c.authorizer, session, scope, resource, enum.PermissionUserEdit); err != nil {
		return nil, err
	}

	return c.CreateNoAuth(ctx, in, false)
}

/*
 * CreateNoAuth creates a new user without auth checks.
 * WARNING: Never call as part of user flow.
 *
 * Note: take admin separately to avoid potential vulnerabilities for user calls.
 */
func (c *Controller) CreateNoAuth(ctx context.Context, in *CreateInput, admin bool) (*types.User, error) {
	if err := c.sanitizeCreateInput(in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	var hash string
	var err error

	user := &types.User{
		UID:         in.UID,
		DisplayName: in.DisplayName,
		Email:       in.Email,
		Salt:        uniuri.NewLen(uniuri.UUIDLen),
		Created:     time.Now().UnixMilli(),
		Updated:     time.Now().UnixMilli(),
		Admin:       admin,
	}

	if in.Source == string(enum.PrincipalSourceZentao) {
		if len(in.PasswordHash) != 0 {
			hash = in.PasswordHash
		} else {
			hashPass := md5.Sum([]byte(in.Password))
			hash = hex.EncodeToString(hashPass[:])
		}
		user.Source = enum.PrincipalSourceZentao
	} else {
		hashPass, err := hashPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to create hash: %w", err)
		}
		hash = string(hashPass)
	}

	user.Password = hash

	err = c.principalStore.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	uCount, err := c.principalStore.CountUsers(ctx, &types.UserFilter{})
	if err != nil {
		return nil, err
	}

	// first 'user' principal will be admin by default.
	if uCount == 1 {
		user.Admin = true
		err = c.principalStore.UpdateUser(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (c *Controller) sanitizeCreateInput(in *CreateInput) error {
	if err := c.principalUIDCheck(in.UID); err != nil {
		return err
	}

	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if err := check.DisplayName(in.DisplayName); err != nil {
		return err
	}

	if in.Source == string(enum.PrincipalSourceZentao) {
		if err := check.HashPassword(in.Password, in.PasswordHash); err != nil {
			return err
		}
		in.Email = strings.TrimSpace(in.Email)
		if len(in.Email) == 0 {
			in.Email = fmt.Sprintf("%s@gitfox.io", in.UID)
		}
	} else {
		in.Email = strings.TrimSpace(in.Email)
		if err := check.Email(in.Email); err != nil {
			return err
		}
		//nolint:revive
		if err := check.Password(in.Password); err != nil {
			return err
		}
	}
	return nil
}
