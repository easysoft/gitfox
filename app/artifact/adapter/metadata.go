// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package adapter

import (
	"encoding/json"

	"github.com/easysoft/gitfox/app/services/protection"
)

type MetadataInterface interface {
	ToJSON() (json.RawMessage, error)
}

type VersionMetadata struct {
	CreatorName string `json:"creator_name"`
}

func (m *VersionMetadata) ToJSON() (json.RawMessage, error) {
	return protection.ToJSON(m)
}
