// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import "net/http"

const (
	// QueryParamGitArchiveFormat was provided for git archive
	QueryParamGitArchiveFormat = "format"

	// PathParamExecutionId was provided for read execution primary key
	PathParamExecutionId = "execution_id"

	// PathParamStageId was provided for read stage primary key
	PathParamStageId = "stage_id"
)

func GetExecutionIdFromPath(r *http.Request) (int64, error) {
	return PathParamAsPositiveInt64(r, PathParamExecutionId)
}

func GetStageIdFromPath(r *http.Request) (int64, error) {
	return PathParamAsPositiveInt64(r, PathParamStageId)
}
