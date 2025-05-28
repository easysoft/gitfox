// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import (
	"errors"
	"net/http"

	"github.com/easysoft/gitfox/types"

	"github.com/golang-module/carbon/v2"
)

const (
	QueryParamStatisticsIncludeMergeRequest = "include_merge_request"
	QueryParamStatisticsIncludePushRequest  = "include_push_request"
	QueryParamStatisticsIncludeReview       = "include_review"
	QueryParamStatisticsIncludeCode         = "include_code"
	QueryParamStatisticsBranch              = "branch"
	QueryParamStatisticsPeriod              = "period"
	QueryParamStatisticsBeginTime           = "begin"
	QueryParamStatisticsEndTime             = "end"
	QueryParamStatisticsTop                 = "top"
	QueryParamStatisticsUser                = "user"
)

func ParseStatisticsIncludeMergeRequestFromQuery(r *http.Request) (bool, error) {
	return QueryParamAsBoolOrDefault(r, QueryParamStatisticsIncludeMergeRequest, false)
}

func ParseStatisticsIncludePushRequestFromQuery(r *http.Request) (bool, error) {
	return QueryParamAsBoolOrDefault(r, QueryParamStatisticsIncludePushRequest, false)
}

func ParseStatisticsIncludeReviewFromQuery(r *http.Request) (bool, error) {
	return QueryParamAsBoolOrDefault(r, QueryParamStatisticsIncludeReview, false)
}

func ParseStatisticsIncludeCodeFromQuery(r *http.Request) (bool, error) {
	return QueryParamAsBoolOrDefault(r, QueryParamStatisticsIncludeCode, true)
}

func ParseStatisticsBranchFromQuery(r *http.Request) string {
	return QueryParamOrDefault(r, QueryParamStatisticsBranch, "")
}

func ParseStatisticsPeriodFromQuery(r *http.Request) string {
	return QueryParamOrDefault(r, QueryParamStatisticsPeriod, "")
}

func ParseStatisticsBeginTimeFromQuery(r *http.Request) string {
	return QueryParamOrDefault(r, QueryParamStatisticsBeginTime, "")
}

func ParseStatisticsEndTimeFromQuery(r *http.Request) string {
	return QueryParamOrDefault(r, QueryParamStatisticsEndTime, "")
}

func ParseStatisticsTopFromQuery(r *http.Request) (int64, error) {
	return QueryParamAsPositiveInt64OrDefault(r, QueryParamStatisticsTop, 10)
}

func ParseStatisticsUserFromQuery(r *http.Request) string {
	return QueryParamOrDefault(r, QueryParamStatisticsUser, "")
}

// ParseRepositoryStatisticsFilter extracts the repository statistics filter from the url.
func ParseRepositoryStatisticsFilter(r *http.Request) (*types.RepositoryStatisticsFilter, error) {
	// recursive is optional to get all repos in a sapce and its subsapces recursively.
	// includeMergeRequest, err := ParseStatisticsIncludeMergeRequestFromQuery(r)
	// if err != nil {
	// 	return nil, err
	// }
	// includePushRequest, err := ParseStatisticsIncludePushRequestFromQuery(r)
	// if err != nil {
	// 	return nil, err
	// }
	// includeReview, err := ParseStatisticsIncludeReviewFromQuery(r)
	// if err != nil {
	// 	return nil, err
	// }
	// includeCode, err := ParseStatisticsIncludeCodeFromQuery(r)
	// if err != nil {
	// 	return nil, err
	// }
	user := ParseStatisticsUserFromQuery(r)
	// top, err := ParseStatisticsTopFromQuery(r)
	// if err != nil {
	// 	return nil, err
	// }

	// if len(user) == 0 && top < 10 {
	// 	top = 10
	// }

	st := ParseStatisticsBeginTimeFromQuery(r)
	et := ParseStatisticsEndTimeFromQuery(r)
	if len(et) == 0 {
		et = carbon.Now().ToDateString()
	}
	return &types.RepositoryStatisticsFilter{
		// IncludeMergeRequest: includeMergeRequest,
		// IncludePushRequest:  includePushRequest,
		// IncludeReview:       includeReview,
		// IncludeCode:         includeCode,
		Branch: ParseStatisticsBranchFromQuery(r),
		// Period:    enum.ParsePeriodSortOption(ParseStatisticsPeriodFromQuery(r)),
		BeginTime: carbon.Parse(st).ToDateString(),
		EndTime:   carbon.Parse(et).Tomorrow().ToDateString(),
		// Top:       top,
		Author: user,
	}, nil
}

// ParseAdminStatisticsFilter extracts the admin statistics filter from the url.
func ParseAdminStatisticsFilter(r *http.Request) (*types.RepositoryStatisticsFilter, error) {
	user := ParseStatisticsUserFromQuery(r)
	if len(user) == 0 {
		return nil, errors.New("user is required")
	}

	st := ParseStatisticsBeginTimeFromQuery(r)
	et := ParseStatisticsEndTimeFromQuery(r)
	if len(et) == 0 {
		et = carbon.Now().ToDateString()
	}
	return &types.RepositoryStatisticsFilter{
		BeginTime: carbon.Parse(st).ToDateString(),
		EndTime:   carbon.Parse(et).Tomorrow().ToDateString(),
		Author:    user,
	}, nil
}
