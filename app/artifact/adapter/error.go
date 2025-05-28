// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package adapter

import (
	"github.com/easysoft/gitfox/app/api/usererror"
)

var (
	ErrMissFormField = usererror.BadRequest("miss required form field")
	ErrMissPathField = usererror.BadRequest("path field is missing")

	ErrInvalidPackageName       = usererror.BadRequest("invalid package name")
	ErrInvalidPackageVersion    = usererror.BadRequest("invalid package version")
	ErrInvalidGroupName         = usererror.BadRequest("invalid group name")
	ErrInvalidPackageContent    = usererror.BadRequest("invalid package content")
	ErrStorageFileAlreadyExists = usererror.BadRequest("storage file already exists")
	ErrStorageFileNotChanged    = usererror.BadRequest("storage file not changed")
)
