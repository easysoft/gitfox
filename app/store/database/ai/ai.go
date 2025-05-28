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

package ai

import (
	"context"
	"errors"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.AIStore = AIStore{}

// NewAIStore returns a new AIStore.
func NewAIStore(db *gorm.DB) AIStore {
	return AIStore{
		db: db,
	}
}

// AIStore implements a store.AIStore backed by a relational database.
type AIStore struct {
	db *gorm.DB
}

type aiConfig struct {
	ID        int64 `gorm:"column:ai_id"`
	SpaceID   int64 `gorm:"column:ai_space_id"`
	Created   int64 `gorm:"column:ai_created"`
	Updated   int64 `gorm:"column:ai_updated"`
	CreatedBy int64 `gorm:"column:ai_created_by"`
	UpdatedBy int64 `gorm:"column:ai_updated_by"`
	IsDefault bool  `gorm:"column:ai_default"`

	Provider enum.Provider `gorm:"column:ai_provider"`
	Model    string        `gorm:"column:ai_model"`
	Endpoint string        `gorm:"column:ai_endpoint"`
	Token    string        `gorm:"column:ai_token"`

	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type aiRequest struct {
	ID         int64            `gorm:"column:ai_id"`
	Created    int64            `gorm:"column:ai_created"`
	Updated    int64            `gorm:"column:ai_updated"`
	RepoID     int64            `gorm:"column:ai_repo_id"`
	PullReqID  int64            `gorm:"column:ai_pullreq_id"`
	ConfigID   int64            `gorm:"column:ai_config_id"`
	Token      int64            `gorm:"column:ai_token"`
	Duration   int64            `gorm:"column:ai_duration"`
	Status     enum.AIReqStatus `gorm:"column:ai_status"`
	Error      string           `gorm:"column:ai_error"`
	ClientMode bool             `gorm:"column:ai_client_mode"`
	// 新增
	ReviewMessage string                     `gorm:"column:ai_review_message"` // 评审意见
	ReviewStatus  enum.AIRequestReviewStatus `gorm:"column:ai_review_status"`  // 评审状态
	ReviewSHA     string                     `gorm:"column:ai_review_sha"`     // 评审的SHA
}

// Find finds an ai config by id.
func (s AIStore) Find(ctx context.Context, id int64) (*types.AIConfig, error) {
	result := &aiConfig{}
	if err := s.db.Where("ai_id = ?", id).Last(result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find ai config by id %v", id)
	}
	key := mapToAIConfig(result)
	return &key, nil
}

// Create creates a new ai config.
func (s AIStore) Create(ctx context.Context, cfg *types.AIConfig) error {
	dbCfg := mapToInternalAIConfig(cfg)
	if err := s.db.Create(&dbCfg).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to create ai config")
	}
	cfg.ID = dbCfg.ID
	return nil
}

// Delete deletes an ai config.
func (s AIStore) Delete(ctx context.Context, id int64) error {
	result := s.db.Where("ai_id = ?", id).Delete(&aiConfig{})
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to delete ai config")
	}
	return nil
}

func (s AIStore) Count(
	ctx context.Context,
	spaceID int64,
) (int64, error) {
	var count int64
	if err := s.db.Model(&aiConfig{}).Where("ai_space_id = ?", spaceID).Count(&count).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed to count space %v ai configs", spaceID)
	}
	return count, nil
}

// List returns the public keys for the principal.
func (s AIStore) List(
	ctx context.Context,
	spaceID int64,
) ([]types.AIConfig, error) {
	stmt := s.db.Model(&aiConfig{}).Where("ai_space_id = ?", spaceID)

	cfgs := make([]aiConfig, 0)
	if err := stmt.Find(&cfgs).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to list space %v ai configs", spaceID)
	}
	aicfgs := mapToAIConfigs(cfgs)
	for i := range aicfgs {
		lastUsedMsg := &aiRequest{}
		s.db.Model(&aiRequest{}).Where("ai_config_id = ?", aicfgs[i].ID).Last(lastUsedMsg)
		if lastUsedMsg.ID != 0 {
			aicfgs[i].Status = lastUsedMsg.Status
			aicfgs[i].Error = lastUsedMsg.Error
			aicfgs[i].RequestTime = lastUsedMsg.Created
		} else {
			aicfgs[i].Status = enum.AIRequestStatusOther
			aicfgs[i].Error = ""
			aicfgs[i].RequestTime = 0
		}
	}
	return aicfgs, nil
}

// Default returns the default ai config for a space.
func (s AIStore) Default(
	ctx context.Context,
	spaceID int64,
) (*types.AIConfig, error) {
	cfg := &aiConfig{}
	if err := s.db.Where("ai_space_id = ? AND ai_default = ?", spaceID, true).First(cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.db.Where("ai_space_id = ?", spaceID).First(cfg).Error; err != nil {
				return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find default ai config for space %v", spaceID)
			}
			return mapToAIConfigPtr(cfg), nil
		}
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find default ai config for space %v", spaceID)
	}
	return mapToAIConfigPtr(cfg), nil
}

func (s AIStore) Record(ctx context.Context, aiReq *types.AIRequest) error {
	dbReq := mapToInternalAIRequest(aiReq)
	if err := s.db.Create(&dbReq).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to record ai request")
	}
	return nil
}

func (s AIStore) LastRecord(ctx context.Context, prID int64) (*types.AIRequest, error) {
	lastRecord := &aiRequest{}
	if err := s.db.Model(&aiRequest{}).Where("ai_pullreq_id = ?", prID).Last(lastRecord).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find last ai request")
	}
	return mapToAIRequestPtr(lastRecord), nil
}

func (s AIStore) Update(ctx context.Context, cfg *types.AIConfig) error {
	dbCfg := mapToInternalAIConfig(cfg)
	if err := s.db.Save(&dbCfg).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed to update ai config")
	}
	return nil
}

func mapToInternalAIRequest(in *types.AIRequest) aiRequest {
	return aiRequest{
		Created:       in.Created,
		Updated:       in.Updated,
		RepoID:        in.RepoID,
		PullReqID:     in.PullReqID,
		ConfigID:      in.ConfigID,
		Token:         in.Token,
		Duration:      in.Duration,
		Status:        in.Status,
		Error:         in.Error,
		ClientMode:    in.ClientMode,
		ReviewMessage: in.ReviewMessage,
		ReviewStatus:  in.ReviewStatus,
		ReviewSHA:     in.ReviewSHA,
	}
}

func mapToInternalAIConfig(in *types.AIConfig) aiConfig {
	return aiConfig{
		ID:        in.ID,
		SpaceID:   in.SpaceID,
		Created:   in.Created,
		Updated:   in.Updated,
		CreatedBy: in.CreatedBy,
		UpdatedBy: in.UpdatedBy,
		IsDefault: in.IsDefault,
		Provider:  in.Provider,
		Model:     in.Model,
		Endpoint:  in.Endpoint,
		Token:     in.Token,
	}
}

func mapToAIConfig(in *aiConfig) types.AIConfig {
	return types.AIConfig{
		ID:        in.ID,
		SpaceID:   in.SpaceID,
		Created:   in.Created,
		Updated:   in.Updated,
		CreatedBy: in.CreatedBy,
		UpdatedBy: in.UpdatedBy,
		IsDefault: in.IsDefault,
		Provider:  in.Provider,
		Model:     in.Model,
		Endpoint:  in.Endpoint,
		Token:     in.Token,
	}
}

func mapToAIConfigPtr(in *aiConfig) *types.AIConfig {
	return &types.AIConfig{
		ID:        in.ID,
		SpaceID:   in.SpaceID,
		Created:   in.Created,
		Updated:   in.Updated,
		CreatedBy: in.CreatedBy,
		UpdatedBy: in.UpdatedBy,
		IsDefault: in.IsDefault,
		Provider:  in.Provider,
		Model:     in.Model,
		Endpoint:  in.Endpoint,
		Token:     in.Token,
	}
}

func mapToAIConfigs(
	keys []aiConfig,
) []types.AIConfig {
	res := make([]types.AIConfig, len(keys))
	for i := 0; i < len(keys); i++ {
		res[i] = mapToAIConfig(&keys[i])
	}
	return res
}

func mapToAIRequestPtr(in *aiRequest) *types.AIRequest {
	return &types.AIRequest{
		Created:       in.Created,
		Updated:       in.Updated,
		RepoID:        in.RepoID,
		PullReqID:     in.PullReqID,
		ConfigID:      in.ConfigID,
		Token:         in.Token,
		Duration:      in.Duration,
		Status:        in.Status,
		Error:         in.Error,
		ClientMode:    in.ClientMode,
		ReviewMessage: in.ReviewMessage,
		ReviewStatus:  in.ReviewStatus,
		ReviewSHA:     in.ReviewSHA,
	}
}
