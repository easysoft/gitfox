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

package auth

import (
	"github.com/easysoft/gitfox/types"
)

// Session contains information of the authenticated principal and auth related metadata.
type Session struct {
	// Principal is the authenticated principal.
	Principal types.Principal

	// Metadata contains auth related information (access grants, tokenId, sshKeyId, ...)
	Metadata Metadata

	// SudoUser is the external authenticated principal
	SudoUser *types.Principal

	// User is the end principal. default to Principal, override if SudoUser exist
	User *types.Principal
}
