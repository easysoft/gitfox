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

package check

import (
	"fmt"
	"regexp"
)

const (
	minPasswordLength = 1
	maxPasswordLength = 128
)

var (
	// ErrPasswordLength is returned when the password
	// is outside of the allowed length.
	ErrPasswordLength = &ValidationError{
		fmt.Sprintf("Password has to be within %d and %d characters", minPasswordLength, maxPasswordLength),
	}
	ErrPasswordHashNotValid = &ValidationError{
		fmt.Sprintf("Password has to be a valid md5 hash"),
	}
	ErrPasswordHash = &ValidationError{
		fmt.Sprintf("Password or PasswordHash has to be provided"),
	}
)

// Password returns true if the Password is valid.
// TODO: add proper password checks.
func Password(pw string) error {
	// validate length
	l := len(pw)
	if l < minPasswordLength || l > maxPasswordLength {
		return ErrPasswordLength
	}

	return nil
}

func HashPassword(pw, pwhash string) error {
	if len(pw) == 0 && len(pwhash) == 0 {
		return ErrPasswordHash
	}
	if len(pwhash) > 0 {
		// validate length
		if len(pwhash) != 32 {
			return ErrPasswordHashNotValid
		}
		match, _ := regexp.MatchString("^[a-fA-F0-9]{32}$", pwhash)
		if match {
			return nil
		}
		return ErrPasswordHashNotValid
	}
	return nil
}

func ValidHashPassword(pwhash string) error {
	if len(pwhash) != 32 {
		return ErrPasswordHashNotValid
	}
	match, _ := regexp.MatchString("^[a-fA-F0-9]{32}$", pwhash)
	if match {
		return nil
	}
	return ErrPasswordHashNotValid
}
