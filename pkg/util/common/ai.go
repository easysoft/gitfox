// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package common

import (
	"errors"

	"github.com/easysoft/gitfox/internal/ai"
	"github.com/easysoft/gitfox/internal/ai/openai"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

// GetClient returns the generative client based on the platform
func GetClient(cfg *types.AIConfig) (ai.Generative, error) {
	switch cfg.Provider {
	case enum.OpenAI, enum.Azure, enum.Anthropic:
		return NewOpenAI(cfg)
	case enum.DeepSeek:
		return NewDeepSeek(cfg)
	}
	return nil, errors.New("invalid provider")
}

func NewOpenAI(cfg *types.AIConfig) (*openai.Client, error) {
	return openai.New(
		openai.WithToken(cfg.Token),
		openai.WithModel(cfg.Model),
		openai.WithOrgID(cfg.OrgID),
		openai.WithProxyURL(cfg.Proxy),
		openai.WithSocksURL(cfg.Socks),
		openai.WithBaseURL(cfg.Endpoint),
		openai.WithTimeout(cfg.Timeout),
		openai.WithMaxTokens(cfg.MaxTokens),
		openai.WithTemperature(float32(cfg.Temperature)),
		openai.WithTopP(float32(cfg.TopP)),
		openai.WithFrequencyPenalty(float32(cfg.FrequencyPenalty)),
		openai.WithPresencePenalty(float32(cfg.PresencePenalty)),
		openai.WithProvider(cfg.Provider),
		openai.WithSkipVerify(true),
		openai.WithHeaders(cfg.Headers),
		openai.WithAPIVersion(cfg.APIVersion),
	)
}

func NewDeepSeek(cfg *types.AIConfig) (*openai.Client, error) {
	return openai.New(
		openai.WithToken(cfg.Token),
		openai.WithModel("deepseek-coder"),
		openai.WithProxyURL(cfg.Proxy),
		openai.WithSocksURL(cfg.Socks),
		openai.WithBaseURL(cfg.Endpoint),
		openai.WithTemperature(1.0),
		openai.WithProvider(enum.DeepSeek),
		openai.WithSkipVerify(true),
	)
}
