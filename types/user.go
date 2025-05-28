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

// Package types defines common data structures.
package types

import (
	"github.com/easysoft/gitfox/types/enum"
)

type (
	// User is a principal representing an end user.
	User struct {
		// Fields from Principal
		ID          int64  `db:"principal_id"             gorm:"column:principal_id;primaryKey" json:"id"`
		UID         string `db:"principal_uid"            gorm:"column:principal_uid"           json:"uid"`
		Email       string `db:"principal_email"          gorm:"column:principal_email"         json:"email"`
		DisplayName string `db:"principal_display_name"   gorm:"column:principal_display_name"  json:"display_name"`
		Admin       bool   `db:"principal_admin"          gorm:"column:principal_admin"         json:"admin"`
		Blocked     bool   `db:"principal_blocked"        gorm:"column:principal_blocked"       json:"blocked"`
		Salt        string `db:"principal_salt"           gorm:"column:principal_salt"          json:"-"`
		Created     int64  `db:"principal_created"        gorm:"column:principal_created"       json:"created"`
		Updated     int64  `db:"principal_updated"        gorm:"column:principal_updated"       json:"updated"`

		// User specific fields
		Password string `db:"principal_user_password"  gorm:"column:principal_user_password"  json:"-"`

		// Source is the source of the user account.
		Source enum.PrincipalSource `db:"principal_user_source" gorm:"column:principal_user_source" json:"-"`
	}

	// UserInput store user account details used to
	// create or update a user.
	UserInput struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
		Name     *string `json:"name"`
		Admin    *bool   `json:"admin"`
	}

	// UserFilter stores user query parameters.
	UserFilter struct {
		Page  int           `json:"page"`
		Size  int           `json:"size"`
		Sort  enum.UserAttr `json:"sort"`
		Order enum.Order    `json:"order"`
		Admin bool          `json:"admin"`
	}
)

func (u *User) ToPrincipal() *Principal {
	return &Principal{
		ID:          u.ID,
		UID:         u.UID,
		Email:       u.Email,
		Type:        enum.PrincipalTypeUser,
		DisplayName: u.DisplayName,
		Admin:       u.Admin,
		Blocked:     u.Blocked,
		Salt:        u.Salt,
		Created:     u.Created,
		Updated:     u.Updated,
	}
}

func (u *User) ToPrincipalInfo() *PrincipalInfo {
	return u.ToPrincipal().ToPrincipalInfo()
}
