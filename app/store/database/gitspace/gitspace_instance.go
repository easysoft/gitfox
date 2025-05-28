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

package gitspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type gitspaceInstance struct {
	ID               int64                          `gorm:"column:gits_id"`
	GitSpaceConfigID int64                          `gorm:"column:gits_gitspace_config_id"`
	URL              null.String                    `gorm:"column:gits_url"`
	State            enum.GitspaceInstanceStateType `gorm:"column:gits_state"`
	// TODO: migrate to principal int64 id to use principal cache and consistent with gitness code.
	UserUID        string                  `gorm:"column:gits_user_uid"`
	ResourceUsage  null.String             `gorm:"column:gits_resource_usage"`
	SpaceID        int64                   `gorm:"column:gits_space_id"`
	LastUsed       int64                   `gorm:"column:gits_last_used"`
	TotalTimeUsed  int64                   `gorm:"column:gits_total_time_used"`
	TrackedChanges null.String             `gorm:"column:gits_tracked_changes"`
	AccessType     enum.GitspaceAccessType `gorm:"column:gits_access_type"`
	AccessKeyRef   null.String             `gorm:"column:gits_access_key_ref"`
	MachineUser    null.String             `gorm:"column:gits_machine_user"`
	Identifier     string                  `gorm:"column:gits_uid"`
	Created        int64                   `gorm:"column:gits_created"`
	Updated        int64                   `gorm:"column:gits_updated"`
}

const (
	gitspaceInstanceTable         = `gitspaces`
	gitspaceInstanceInsertColumns = `
        gits_gitspace_config_id,
        gits_url,
        gits_state,
        gits_user_uid,
        gits_resource_usage,
        gits_space_id,
        gits_created,
        gits_updated,
        gits_last_used,
        gits_total_time_used,
        gits_tracked_changes,
        gits_access_type,
        gits_machine_user,
        gits_uid,
        gits_access_key_ref`

	gitspaceInstanceSelectColumns = "gits_id," + gitspaceInstanceInsertColumns
)

var _ store.GitspaceInstanceStore = (*gitspaceInstanceStore)(nil)

// TODO Stubbed Impl
// NewGitspaceInstanceStore returns a new GitspaceInstanceStore.
func NewGitspaceInstanceStore(db *gorm.DB) store.GitspaceInstanceStore {
	return &gitspaceInstanceStore{
		db: db,
	}
}

type gitspaceInstanceStore struct {
	db *gorm.DB
}

func (g gitspaceInstanceStore) Find(ctx context.Context, id int64) (*types.GitspaceInstance, error) {
	gitspace := new(gitspaceInstance)
	if err := g.db.Table(gitspaceInstanceTable).Where("gits_id = ?", id).First(gitspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Failed to find gitspace")
	}
	return g.mapToGitspaceInstance(ctx, gitspace)
}

func (g gitspaceInstanceStore) FindByIdentifier(
	ctx context.Context,
	identifier string,
) (*types.GitspaceInstance, error) {
	gitspace := new(gitspaceInstance)
	if err := g.db.Table(gitspaceInstanceTable).
		Where("gits_uid = ?", identifier).
		First(gitspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Failed to find gitspace")
	}
	return g.mapToGitspaceInstance(ctx, gitspace)
}

func (g gitspaceInstanceStore) Create(ctx context.Context, gitspaceInstancePtr *types.GitspaceInstance) error {
	gitspace := &gitspaceInstance{
		GitSpaceConfigID: gitspaceInstancePtr.GitSpaceConfigID,
		URL:              null.NewString(*gitspaceInstancePtr.URL, *gitspaceInstancePtr.URL != ""),
		State:            gitspaceInstancePtr.State,
		UserUID:          gitspaceInstancePtr.UserID,
		ResourceUsage:    null.NewString(*gitspaceInstancePtr.ResourceUsage, *gitspaceInstancePtr.ResourceUsage != ""),
		SpaceID:          gitspaceInstancePtr.SpaceID,
		Created:          gitspaceInstancePtr.Created,
		Updated:          gitspaceInstancePtr.Updated,
		LastUsed:         gitspaceInstancePtr.LastUsed,
		TotalTimeUsed:    gitspaceInstancePtr.TotalTimeUsed,
		TrackedChanges:   null.NewString(*gitspaceInstancePtr.TrackedChanges, *gitspaceInstancePtr.TrackedChanges != ""),
		AccessType:       gitspaceInstancePtr.AccessType,
		AccessKeyRef:     null.NewString(*gitspaceInstancePtr.AccessKeyRef, *gitspaceInstancePtr.AccessKeyRef != ""),
		MachineUser:      null.NewString(*gitspaceInstancePtr.MachineUser, *gitspaceInstancePtr.MachineUser != ""),
		Identifier:       gitspaceInstancePtr.Identifier,
	}
	if err := g.db.Table(gitspaceInstanceTable).Create(gitspace).Error; err != nil {
		return errors.Wrap(err, "Failed to create gitspace")
	}
	gitspaceInstancePtr.ID = gitspace.ID
	return nil
}

func (g gitspaceInstanceStore) Update(
	ctx context.Context,
	gitspaceInstance *types.GitspaceInstance,
) error {
	if err := g.db.Table(gitspaceInstanceTable).
		Where("gits_id = ?", gitspaceInstance.ID).
		Updates(map[string]interface{}{
			"gits_state":     gitspaceInstance.State,
			"gits_last_used": gitspaceInstance.LastUsed,
			"gits_url":       gitspaceInstance.URL,
			"gits_updated":   gitspaceInstance.Updated,
		}).Error; err != nil {
		return errors.Wrap(err, "Failed to update gitspace")
	}
	return nil
}

func (g gitspaceInstanceStore) FindLatestByGitspaceConfigID(
	ctx context.Context,
	gitspaceConfigID int64,
) (*types.GitspaceInstance, error) {
	gitspace := new(gitspaceInstance)
	if err := g.db.Table(gitspaceInstanceTable).
		Where("gits_gitspace_config_id = ?", gitspaceConfigID).
		Order("gits_created DESC").
		First(gitspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Failed to find gitspace")
	}
	return g.mapToGitspaceInstance(ctx, gitspace)
}

func (g gitspaceInstanceStore) List(
	ctx context.Context,
	filter *types.GitspaceFilter,
) ([]*types.GitspaceInstance, error) {
	var dst []*gitspaceInstance
	query := g.db.Table(gitspaceInstanceTable).
		Where("gits_space_id = ?", filter.SpaceIDs).
		Where("gits_user_uid = ?", filter.UserID).
		Order("gits_created ASC")
	if err := query.Find(&dst).Error; err != nil {
		return nil, errors.Wrap(err, "Failed to execute gorm query")
	}
	return g.mapToGitspaceInstances(ctx, dst)
}

func (g gitspaceInstanceStore) FindAllLatestByGitspaceConfigID(
	ctx context.Context,
	gitspaceConfigIDs []int64,
) ([]*types.GitspaceInstance, error) {
	var whereClause = "(1=0)"
	if len(gitspaceConfigIDs) > 0 {
		whereClause = fmt.Sprintf("gits_gitspace_config_id IN (%s)",
			strings.Trim(strings.Join(strings.Split(fmt.Sprint(gitspaceConfigIDs), " "), ","), "[]"))
	}
	baseSelect := g.db.Table(gitspaceInstanceTable).
		Select("*").
		Select("ROW_NUMBER() OVER (PARTITION BY gits_gitspace_config_id ORDER BY gits_created DESC) AS rn").
		Where(whereClause)

	stmt := g.db.Table(gitspaceInstanceTable).
		Select(gitspaceInstanceSelectColumns).
		Joins(fmt.Sprintf("INNER JOIN (%s) AS RankedRows ON %s.gits_id = RankedRows.gits_id", baseSelect, gitspaceInstanceTable)).
		Where("rn = 1")

	var dst []*gitspaceInstance
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, errors.Wrap(err, "Failed executing custom list query")
	}
	return g.mapToGitspaceInstances(ctx, dst)
}

func (g gitspaceInstanceStore) mapToGitspaceInstance(
	_ context.Context,
	in *gitspaceInstance,
) (*types.GitspaceInstance, error) {
	var res = &types.GitspaceInstance{
		ID:               in.ID,
		Identifier:       in.Identifier,
		GitSpaceConfigID: in.GitSpaceConfigID,
		URL:              in.URL.Ptr(),
		State:            in.State,
		UserID:           in.UserUID,
		ResourceUsage:    in.ResourceUsage.Ptr(),
		LastUsed:         in.LastUsed,
		TotalTimeUsed:    in.TotalTimeUsed,
		TrackedChanges:   in.TrackedChanges.Ptr(),
		AccessType:       in.AccessType,
		AccessKeyRef:     in.AccessKeyRef.Ptr(),
		MachineUser:      in.MachineUser.Ptr(),
		SpaceID:          in.SpaceID,
		Created:          in.Created,
		Updated:          in.Updated,
	}
	return res, nil
}

func (g gitspaceInstanceStore) mapToGitspaceInstances(
	ctx context.Context,
	instances []*gitspaceInstance,
) ([]*types.GitspaceInstance, error) {
	var err error
	res := make([]*types.GitspaceInstance, len(instances))
	for i := range instances {
		res[i], err = g.mapToGitspaceInstance(ctx, instances[i])
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
