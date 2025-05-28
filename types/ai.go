// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"time"

	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

type AIConfig struct {
	ID        int64 `json:"id"`
	SpaceID   int64 `json:"space_id"`
	Created   int64 `json:"created"`
	Updated   int64 `json:"updated"`
	CreatedBy int64 `json:"created_by"`
	UpdatedBy int64 `json:"updated_by,omitempty"`
	IsDefault bool  `json:"is_default"`

	Provider enum.Provider `json:"provider"`
	Model    string        `json:"model"`
	Endpoint string        `json:"endpoint"`
	Token    string        `json:"-"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	AIExtConfig
	// 最近使用消息
	AILastUsedMsg
}

type AIExtConfig struct {
	OrgID            string        `json:"org_id,omitempty"`
	Proxy            string        `json:"proxy,omitempty"`
	Socks            string        `json:"socks,omitempty"`
	Timeout          time.Duration `json:"timeout,omitempty"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	Temperature      float64       `json:"temperature,omitempty"`
	TopP             float64       `json:"top_p,omitempty"`
	FrequencyPenalty float64       `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64       `json:"presence_penalty,omitempty"`
	SkipVerify       bool          `json:"skip_verify,omitempty"`
	Headers          []string      `json:"headers,omitempty"`
	APIVersion       string        `json:"api_version,omitempty"`
}

type AILastUsedMsg struct {
	Status      enum.AIReqStatus `json:"status"`
	Error       string           `json:"error,omitempty"`
	RequestTime int64            `json:"request_time"`
}

type AIRequest struct {
	Created    int64            `json:"created"`
	Updated    int64            `json:"updated"`
	RepoID     int64            `json:"repo_id"`
	PullReqID  int64            `json:"pull_req_id"`
	ConfigID   int64            `json:"-"`
	Token      int64            `json:"token"`
	Duration   int64            `json:"duration"`
	Status     enum.AIReqStatus `json:"status"`
	Error      string           `json:"error,omitempty"`
	ClientMode bool             `json:"-"`

	ReviewMessage string                     `json:"review_message,omitempty"`
	ReviewStatus  enum.AIRequestReviewStatus `json:"review_status,omitempty"`
	ReviewSHA     string                     `json:"sha,omitempty"`
}
