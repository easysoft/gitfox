// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package enum

type Provider string

const (
	OpenAI    Provider = "openai"
	Azure     Provider = "azure"
	Gemini    Provider = "gemini"
	Anthropic Provider = "anthropic"
	DeepSeek  Provider = "deepseek"
	Custom    Provider = "custom" // 自定义, 通用OneHub
)

// String returns the string representation of the Provider.
func (p Provider) String() string {
	return string(p)
}

// IsValid returns true if the Platform is valid.
func (p Provider) IsValid() bool {
	switch p {
	case OpenAI, Azure, Gemini, Anthropic, DeepSeek, Custom:
		return true
	}
	return false
}

const (
	DefaultMaxTokens   = 1000
	DefaultModel       = "gpt-4o-mini"
	DefaultTemperature = 1.0
	DefaultTopP        = 1.0
)

type AIReqStatus string

const (
	AIRequestStatusSuccess AIReqStatus = "success"
	AIRequestStatusFailed  AIReqStatus = "failed"
	AIRequestStatusOther   AIReqStatus = "other"
)

type AIRequestReviewStatus string

const (
	AIRequestReviewApproved AIRequestReviewStatus = "approved"
	AIRequestReviewRejected AIRequestReviewStatus = "rejected"
	AIRequestReviewUnknown  AIRequestReviewStatus = "unknown"
	AIRequestReviewIgnore   AIRequestReviewStatus = "ignore"
)
