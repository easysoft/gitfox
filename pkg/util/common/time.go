// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package common

import (
	"time"

	"github.com/golang-module/carbon/v2"
)

// DayStartTime 获取一天的开始时间
func DayStartTime(t carbon.Carbon) time.Time {
	return t.StartOfDay().StdTime()
}

// DayEndTime 获取一天的结束时间
func DayEndTime(t carbon.Carbon) time.Time {
	return t.EndOfDay().StdTime()
}

// DayStartUnix 获取一天的开始时间 秒
func DayStartUnix(t carbon.Carbon) int64 {
	return DayStartTime(t).Unix()
}

// DayEndUnix 获取一天的结束时间 秒
func DayEndUnix(t carbon.Carbon) int64 {
	return DayEndTime(t).Unix()
}

// DayStartUnixMilli 获取一天的开始时间 毫秒
func DayStartUnixMilli(t carbon.Carbon) int64 {
	return DayStartTime(t).UnixMilli()
}

// DayEndUnixMilli 获取一天的结束时间 毫秒
func DayEndUnixMilli(t carbon.Carbon) int64 {
	return DayEndTime(t).UnixMilli()
}

// DayStartUnixRFC3339 获取一天的开始时间 格式化 2025-01-01T00:00:00Z
func DayStartUnixRFC3339(t carbon.Carbon) string {
	return DayStartTime(t).Format(time.RFC3339)
}

// DayEndUnixRFC3339 获取一天的结束时间 格式化
func DayEndUnixRFC3339(t carbon.Carbon) string {
	return DayEndTime(t).Format(time.RFC3339)
}
