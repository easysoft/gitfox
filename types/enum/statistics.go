// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package enum

import "strings"

// PeriodSortOption specifies the available sort options for period.
type PeriodSortOption string

const (
	PeriodSortOptionDefault PeriodSortOption = "daily"
	PeriodSortOptionWeek    PeriodSortOption = "weekly"
	PeriodSortOptionMonth   PeriodSortOption = "monthly"
	PeriodSortOptionYear    PeriodSortOption = "yearly"
)

// String returns a string representation of the tag sort option.
func (o PeriodSortOption) String() string {
	switch o {
	case PeriodSortOptionDefault:
		return "daily"
	case PeriodSortOptionWeek:
		return "weekly"
	case PeriodSortOptionMonth:
		return "monthly"
	case PeriodSortOptionYear:
		return "yearly"
	default:
		return "daily"
	}
}

func ParsePeriodSortOption(s string) PeriodSortOption {
	switch strings.ToLower(s) {
	case "weekly":
		return PeriodSortOptionWeek
	case "monthly":
		return PeriodSortOptionMonth
	case "yearly":
		return PeriodSortOptionYear
	default:
		return PeriodSortOptionDefault
	}
}
