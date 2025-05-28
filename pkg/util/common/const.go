// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package common

const (
	// DefaultPipelineExecutionCreatedUnix is the default unix timestamp for pipeline execution creation time. 2024-03-01 00:00:00
	DefaultPipelineExecutionCreatedUnix int64 = 1709222400000

	// DefaultGitImage is the default git image
	DefaultGitImage = "hub.zentao.net/ci/gitfox-plugin-git:2.0"
	// DefaultPlaceholderImage is the default placeholder image
	DefaultPlaceholderImage = "hub.zentao.net/ci/gitfox-plugin-placeholder:1"
)
