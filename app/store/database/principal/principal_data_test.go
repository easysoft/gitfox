// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package principal_test

import (
	"time"

	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

var current = time.Now().UnixMilli()

var svcItems = []types.Service{
	{ID: 1, UID: "svc_1", Email: "svc_1" + emailSuffix, DisplayName: "Service 1", Updated: current},
	{ID: 2, UID: "svc_2", Email: "svc_2" + emailSuffix, DisplayName: "Service 2", Updated: current},
	{ID: 3, UID: "svc_3", Email: "svc_3" + emailSuffix, DisplayName: "Service 3", Updated: current},
	{ID: 4, UID: "svc_4", Email: "svc_4" + emailSuffix, DisplayName: "Service 4", Updated: current},
}

var svcAccountItems = []types.ServiceAccount{
	{ID: 11, UID: "sa_1", Email: "sa_1" + emailSuffix, DisplayName: "ServiceAccount 1", Updated: current,
		ParentType: enum.ParentResourceTypeSpace, ParentID: 1},
	{ID: 12, UID: "sa_2", Email: "sa_2" + emailSuffix, DisplayName: "ServiceAccount 2", Updated: current,
		ParentType: enum.ParentResourceTypeSpace, ParentID: 1},
	{ID: 13, UID: "sa_3", Email: "sa_3" + emailSuffix, DisplayName: "ServiceAccount 3", Updated: current,
		ParentType: enum.ParentResourceTypeSpace, ParentID: 2},
	{ID: 14, UID: "sa_4", Email: "sa_4" + emailSuffix, DisplayName: "ServiceAccount 4", Updated: current,
		ParentType: enum.ParentResourceTypeRepo, ParentID: 1},
	{ID: 15, UID: "sa_5", Email: "sa_5" + emailSuffix, DisplayName: "ServiceAccount 5", Updated: current,
		ParentType: enum.ParentResourceTypeRepo, ParentID: 2},
	{ID: 16, UID: "sa_6", Email: "sa_6" + emailSuffix, DisplayName: "ServiceAccount 6", Updated: current,
		ParentType: enum.ParentResourceTypeRepo, ParentID: 3},
}

var userItems = []types.User{
	{ID: 21, UID: "admin", Email: "admin" + emailSuffix, Admin: true},
	{ID: 22, UID: "user1", Email: "user1" + emailSuffix, Admin: false},
	{ID: 23, UID: "user2", Email: "user2" + emailSuffix, Admin: false},
	{ID: 24, UID: "admin2", Email: "admin2" + emailSuffix, Admin: true},
}
