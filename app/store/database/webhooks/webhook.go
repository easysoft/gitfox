// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package webhooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ store.WebhookStore = (*WebhookStore)(nil)

// NewWebhookOrmStore returns a new WebhookStore.
func NewWebhookOrmStore(db *gorm.DB) *WebhookStore {
	return &WebhookStore{
		db: db,
	}
}

// WebhookStore implements store.Webhook backed by a relational database.
type WebhookStore struct {
	db *gorm.DB
}

// webhook is an internal representation used to store webhook data in the database.
type webhook struct {
	ID        int64    `db:"webhook_id"          gorm:"column:webhook_id;primaryKey"`
	Version   int64    `db:"webhook_version"     gorm:"column:webhook_version"`
	RepoID    null.Int `db:"webhook_repo_id"     gorm:"column:webhook_repo_id"`
	SpaceID   null.Int `db:"webhook_space_id"    gorm:"column:webhook_space_id"`
	CreatedBy int64    `db:"webhook_created_by"  gorm:"column:webhook_created_by"`
	Created   int64    `db:"webhook_created"     gorm:"column:webhook_created"`
	Updated   int64    `db:"webhook_updated"     gorm:"column:webhook_updated"`
	Internal  bool     `db:"webhook_internal"    gorm:"column:webhook_internal"`

	Identifier string `db:"webhook_uid" gorm:"column:webhook_uid"`
	// TODO [CODE-1364]: Remove once UID/Identifier migration is completed.
	DisplayName           string      `db:"webhook_display_name"             gorm:"column:webhook_display_name"`
	Description           string      `db:"webhook_description"              gorm:"column:webhook_description"`
	URL                   string      `db:"webhook_url"                      gorm:"column:webhook_url"`
	Secret                string      `db:"webhook_secret"                   gorm:"column:webhook_secret"`
	Enabled               bool        `db:"webhook_enabled"                  gorm:"column:webhook_enabled"`
	Insecure              bool        `db:"webhook_insecure"                 gorm:"column:webhook_insecure"`
	Triggers              string      `db:"webhook_triggers"                 gorm:"column:webhook_triggers"`
	LatestExecutionResult null.String `db:"webhook_latest_execution_result"  gorm:"column:webhook_latest_execution_result"`
}

const (
	tableWh = "webhooks"
)

// Find finds the webhook by id.
func (s *WebhookStore) Find(ctx context.Context, id int64) (*types.Webhook, error) {
	dst := &webhook{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWh).First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
	}

	res, err := mapToWebhook(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to map webhook to external type: %w", err)
	}

	return res, nil
}

// FindByIdentifier finds the webhook with the given Identifier for the given parent.
func (s *WebhookStore) FindByIdentifier(
	ctx context.Context,
	parentType enum.WebhookParent,
	parentID int64,
	identifier string,
) (*types.Webhook, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWh).
		Where("LOWER(webhook_uid) = ?", strings.ToLower(identifier))

	switch parentType {
	case enum.WebhookParentRepo:
		stmt = stmt.Where("webhook_repo_id = ?", parentID)
	case enum.WebhookParentSpace:
		stmt = stmt.Where("webhook_space_id = ?", parentID)
	default:
		return nil, fmt.Errorf("webhook parent type '%s' is not supported", parentType)
	}

	dst := &webhook{}
	if err := stmt.Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
	}

	res, err := mapToWebhook(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to map webhook to external type: %w", err)
	}

	return res, nil
}

// Create creates a new webhook.
func (s *WebhookStore) Create(ctx context.Context, hook *types.Webhook) error {
	dbHook, err := mapToInternalWebhook(hook)
	if err != nil {
		return fmt.Errorf("failed to map webhook to internal db type: %w", err)
	}

	if err = dbtx.GetOrmAccessor(ctx, s.db).Table(tableWh).Create(dbHook).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	hook.ID = dbHook.ID
	return nil
}

// Update updates an existing webhook.
func (s *WebhookStore) Update(ctx context.Context, hook *types.Webhook) error {
	dbHook, err := mapToInternalWebhook(hook)
	if err != nil {
		return fmt.Errorf("failed to map webhook to internal db type: %w", err)
	}

	// update Version (used for optimistic locking) and Updated time
	dbHook.Version++
	dbHook.Updated = time.Now().UnixMilli()

	updateFields := []string{"Version", "Updated", "Identifier", "DisplayName", "Description", "URL", "Secret",
		"Enabled", "Insecure", "Triggers", "LatestExecutionResult", "Internal",
	}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWh).
		Where(&webhook{ID: hook.ID, Version: dbHook.Version - 1}).
		Select(updateFields).Updates(dbHook)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update webhook")
	}

	count := res.RowsAffected

	if count == 0 {
		return gitfox_store.ErrVersionConflict
	}

	hook.Version = dbHook.Version
	hook.Updated = dbHook.Updated

	return nil
}

// UpdateOptLock updates the webhook using the optimistic locking mechanism.
func (s *WebhookStore) UpdateOptLock(ctx context.Context, hook *types.Webhook,
	mutateFn func(hook *types.Webhook) error) (*types.Webhook, error) {
	for {
		dup := *hook

		err := mutateFn(&dup)
		if err != nil {
			return nil, fmt.Errorf("failed to mutate the webhook: %w", err)
		}

		err = s.Update(ctx, &dup)
		if err == nil {
			return &dup, nil
		}
		if !errors.Is(err, gitfox_store.ErrVersionConflict) {
			return nil, fmt.Errorf("failed to update the webhook: %w", err)
		}

		hook, err = s.Find(ctx, hook.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to find the latst version of the webhook: %w", err)
		}
	}
}

// Delete deletes the webhook for the given id.
func (s *WebhookStore) Delete(ctx context.Context, id int64) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWh).Delete(webhook{}, id).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "The delete query failed")
	}

	return nil
}

// DeleteByIdentifier deletes the webhook with the given Identifier for the given parent.
func (s *WebhookStore) DeleteByIdentifier(
	ctx context.Context,
	parentType enum.WebhookParent,
	parentID int64,
	identifier string,
) error {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWh).
		Where("LOWER(webhook_uid) = ?", strings.ToLower(identifier))

	switch parentType {
	case enum.WebhookParentRepo:
		stmt = stmt.Where("webhook_repo_id = ?", parentID)
	case enum.WebhookParentSpace:
		stmt = stmt.Where("webhook_space_id = ?", parentID)
	default:
		return fmt.Errorf("webhook parent type '%s' is not supported", parentType)
	}

	if err := stmt.Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "The delete query failed")
	}

	return nil
}

// Count counts the webhooks for a given parent type and id.
func (s *WebhookStore) Count(
	ctx context.Context,
	parents []types.WebhookParentInfo,
	opts *types.WebhookFilter,
) (int64, error) {
	stmt := s.db.Table(tableWh)

	stmt = selectParents(parents, stmt)
	if stmt == nil {
		return 0, fmt.Errorf("failed to select parents")
	}

	stmt = applyWebhookFilter(opts, stmt)

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}

	return count, nil
}

func (s *WebhookStore) List(
	ctx context.Context,
	parents []types.WebhookParentInfo,
	opts *types.WebhookFilter,
) ([]*types.Webhook, error) {
	stmt := s.db.Table(tableWh)

	stmt = selectParents(parents, stmt)
	if stmt == nil {
		return nil, fmt.Errorf("failed to select parents")
	}

	stmt = applyWebhookFilter(opts, stmt)

	stmt = stmt.Limit(database.GormLimit(opts.Size))
	stmt = stmt.Offset(database.GormOffset(opts.Page, opts.Size))

	switch opts.Sort {
	// TODO [CODE-1364]: Remove once UID/Identifier migration is completed
	case enum.WebhookAttrID, enum.WebhookAttrNone:
		// NOTE: string concatenation is safe because the
		// order attribute is an enum and is not user-defined,
		// and is therefore not subject to injection attacks.
		stmt = stmt.Order("webhook_id " + opts.Order.String())

		// TODO [CODE-1363]: remove after identifier migration.
	case enum.WebhookAttrUID, enum.WebhookAttrIdentifier:
		stmt = stmt.Order("LOWER(webhook_uid) " + opts.Order.String())
		// TODO [CODE-1364]: Remove once UID/Identifier migration is completed
	case enum.WebhookAttrDisplayName:
		stmt = stmt.Order("webhook_display_name " + opts.Order.String())
		//TODO: Postgres does not support COLLATE NOCASE for UTF8
	case enum.WebhookAttrCreated:
		stmt = stmt.Order("webhook_created " + opts.Order.String())
	case enum.WebhookAttrUpdated:
		stmt = stmt.Order("webhook_updated " + opts.Order.String())
	}

	dst := []*webhook{}
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
	}

	res, err := mapToWebhooks(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to map webhooks to external type: %w", err)
	}

	return res, nil
}

func applyWebhookFilter(
	opts *types.WebhookFilter,
	stmt *gorm.DB,
) *gorm.DB {
	if opts.Query != "" {
		stmt = stmt.Where("LOWER(webhook_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(opts.Query)))
	}

	if opts.SkipInternal {
		stmt = stmt.Where("webhook_internal != ?", true)
	}

	return stmt
}

func selectParents(
	parents []types.WebhookParentInfo,
	stmt *gorm.DB,
) *gorm.DB {
	for _, parent := range parents {
		switch parent.Type {
		case enum.WebhookParentRepo:
			stmt = stmt.Where("webhook_repo_id = ?", parent.ID)
		case enum.WebhookParentSpace:
			stmt = stmt.Where("webhook_space_id = ? or webhook_space_id is null", parent.ID)
		default:
			// Handle unsupported parent type
			return nil
		}
	}

	return stmt
}

func mapToWebhook(hook *webhook) (*types.Webhook, error) {
	res := &types.Webhook{
		ID:         hook.ID,
		Version:    hook.Version,
		CreatedBy:  hook.CreatedBy,
		Created:    hook.Created,
		Updated:    hook.Updated,
		Identifier: hook.Identifier,
		// TODO [CODE-1364]: Remove once UID/Identifier migration is completed
		DisplayName:           hook.DisplayName,
		Description:           hook.Description,
		URL:                   hook.URL,
		Secret:                hook.Secret,
		Enabled:               hook.Enabled,
		Insecure:              hook.Insecure,
		Triggers:              triggersFromString(hook.Triggers),
		LatestExecutionResult: (*enum.WebhookExecutionResult)(hook.LatestExecutionResult.Ptr()),
		Internal:              hook.Internal,
	}

	switch {
	case hook.RepoID.Valid && hook.SpaceID.Valid:
		return nil, fmt.Errorf("both repoID and spaceID are set for hook %d", hook.ID)
	case hook.RepoID.Valid:
		res.ParentType = enum.WebhookParentRepo
		res.ParentID = hook.RepoID.Int64
	case hook.SpaceID.Valid:
		res.ParentType = enum.WebhookParentSpace
		res.ParentID = hook.SpaceID.Int64
	default:
		return nil, fmt.Errorf("neither repoID nor spaceID are set for hook %d", hook.ID)
	}

	return res, nil
}

func mapToInternalWebhook(hook *types.Webhook) (*webhook, error) {
	res := &webhook{
		ID:         hook.ID,
		Version:    hook.Version,
		CreatedBy:  hook.CreatedBy,
		Created:    hook.Created,
		Updated:    hook.Updated,
		Identifier: hook.Identifier,
		// TODO [CODE-1364]: Remove once UID/Identifier migration is completed
		DisplayName:           hook.DisplayName,
		Description:           hook.Description,
		URL:                   hook.URL,
		Secret:                hook.Secret,
		Enabled:               hook.Enabled,
		Insecure:              hook.Insecure,
		Triggers:              triggersToString(hook.Triggers),
		LatestExecutionResult: null.StringFromPtr((*string)(hook.LatestExecutionResult)),
		Internal:              hook.Internal,
	}
	switch hook.ParentType {
	case enum.WebhookParentRepo:
		res.RepoID = null.IntFrom(hook.ParentID)
	case enum.WebhookParentSpace:
		res.SpaceID = null.IntFrom(hook.ParentID)
	default:
		return nil, fmt.Errorf("webhook parent type '%s' is not supported", hook.ParentType)
	}
	return res, nil
}

func mapToWebhooks(hooks []*webhook) ([]*types.Webhook, error) {
	var err error
	m := make([]*types.Webhook, len(hooks))
	for i, hook := range hooks {
		m[i], err = mapToWebhook(hook)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// triggersSeparator defines the character that's used to join triggers for storing them in the DB
// ASSUMPTION: triggers are defined in an enum and don't contain ",".
const triggersSeparator = ","

func triggersFromString(triggersString string) []enum.WebhookTrigger {
	if triggersString == "" {
		return []enum.WebhookTrigger{}
	}

	rawTriggers := strings.Split(triggersString, triggersSeparator)

	triggers := make([]enum.WebhookTrigger, len(rawTriggers))
	for i, rawTrigger := range rawTriggers {
		// ASSUMPTION: trigger is valid value (as we wrote it to DB)
		triggers[i] = enum.WebhookTrigger(rawTrigger)
	}

	return triggers
}

func triggersToString(triggers []enum.WebhookTrigger) string {
	rawTriggers := make([]string, len(triggers))
	for i := range triggers {
		rawTriggers[i] = string(triggers[i])
	}

	return strings.Join(rawTriggers, triggersSeparator)
}
