// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
)

var (
	Version       string
	BuildDate     string
	GitCommitHash string
	GitBranch     string
)

// GetVersionWithHash returns the version string with the git commit hash and build date.
func GetVersionWithHash() string {
	return fmt.Sprintf("%s-%s-%s", Version, GitCommitHash, BuildDate)
}
