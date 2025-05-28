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

package database

import (
	"context"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/jmoiron/sqlx"
)

var _ store.AIStore = AIStore{}

// NewAIStore returns a new AIStore.
func NewAIStore(db *sqlx.DB) AIStore {
	return AIStore{
		db: db,
	}
}

// AIStore implements a store.AIStore backed by a relational database.
type AIStore struct {
	db *sqlx.DB
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
}

// Find fetches a job by its unique space_id.
func (s AIStore) Find(ctx context.Context, spaceID int64) (*types.AIConfig, error) {
	return nil, nil
}

// Create creates a new ai config.
func (s AIStore) Create(ctx context.Context, cfg *types.AIConfig) error {
	return nil
}

// Delete deletes an ai config.
func (s AIStore) Delete(ctx context.Context, id int64) error {
	return nil
}

func (s AIStore) Count(
	ctx context.Context,
	spaceID int64,
) (int64, error) {
	return 0, nil
}

// List returns the public keys for the principal.
func (s AIStore) List(
	ctx context.Context,
	spaceID int64,
) ([]types.AIConfig, error) {
	cfgs := make([]aiConfig, 0)
	return mapToAIConfigs(cfgs), nil
}

// Default returns the default ai config for a space.
func (s AIStore) Default(
	ctx context.Context,
	spaceID int64,
) (*types.AIConfig, error) {
	return nil, nil
}

func (s AIStore) Record(ctx context.Context, aiReq *types.AIRequest) error {
	return nil
}

func (s AIStore) LastRecord(ctx context.Context, prID int64) (*types.AIRequest, error) {
	return nil, nil
}

// Update updates an ai config.
func (s AIStore) Update(ctx context.Context, cfg *types.AIConfig) error {
	return nil
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

func mapToAIConfigs(
	keys []aiConfig,
) []types.AIConfig {
	res := make([]types.AIConfig, len(keys))
	for i := 0; i < len(keys); i++ {
		res[i] = mapToAIConfig(&keys[i])
	}
	return res
}
