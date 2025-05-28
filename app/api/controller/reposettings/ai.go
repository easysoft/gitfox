// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package reposettings

import (
	"github.com/easysoft/gitfox/app/services/settings"

	"github.com/gotidy/ptr"
)

type AISettings struct {
	AIReviewEnabled *bool `json:"ai_review_enabled" yaml:"ai_review_enabled"`
	SpaceAIProvider int64 `json:"space_ai_provider" yaml:"space_ai_provider"`
}

func GetDefaultAISettings() *AISettings {
	return &AISettings{
		AIReviewEnabled: ptr.Bool(settings.DefaultAIReviewEnabled),
	}
}

func GetAISettingsMappings(s *AISettings) []settings.SettingHandler {
	return []settings.SettingHandler{
		settings.Mapping(settings.KeyAIReviewEnabled, s.AIReviewEnabled),
	}
}

func GetAISettingsAsKeyValues(s *AISettings) []settings.KeyValue {
	kvs := make([]settings.KeyValue, 0, 1)
	if s.AIReviewEnabled != nil {
		kvs = append(kvs, settings.KeyValue{Key: settings.KeyAIReviewEnabled, Value: *s.AIReviewEnabled})
	}
	return kvs
}
