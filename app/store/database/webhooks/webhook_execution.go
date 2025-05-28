// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package webhooks

import (
	"context"
	"time"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ store.WebhookExecutionStore = (*WebhookExecutionStore)(nil)

// NewWebhookExecutionOrmStore returns a new WebhookExecutionStore.
func NewWebhookExecutionOrmStore(db *gorm.DB) *WebhookExecutionStore {
	return &WebhookExecutionStore{
		db: db,
	}
}

// WebhookExecutionStore implements store.WebhookExecution backed by a relational database.
type WebhookExecutionStore struct {
	db *gorm.DB
}

// webhookExecution is used to store executions of webhooks
// The object should be later re-packed into a different struct to return it as an API response.
type webhookExecution struct {
	ID                 int64                       `gorm:"column:webhook_execution_id;primaryKey"`
	RetriggerOf        null.Int                    `gorm:"column:webhook_execution_retrigger_of"`
	Retriggerable      bool                        `gorm:"column:webhook_execution_retriggerable"`
	WebhookID          int64                       `gorm:"column:webhook_execution_webhook_id"`
	TriggerType        enum.WebhookTrigger         `gorm:"column:webhook_execution_trigger_type"`
	TriggerID          string                      `gorm:"column:webhook_execution_trigger_id"`
	Result             enum.WebhookExecutionResult `gorm:"column:webhook_execution_result"`
	Created            int64                       `gorm:"column:webhook_execution_created"`
	Duration           int64                       `gorm:"column:webhook_execution_duration"`
	Error              string                      `gorm:"column:webhook_execution_error"`
	RequestURL         string                      `gorm:"column:webhook_execution_request_url"`
	RequestHeaders     string                      `gorm:"column:webhook_execution_request_headers"`
	RequestBody        string                      `gorm:"column:webhook_execution_request_body"`
	ResponseStatusCode int                         `gorm:"column:webhook_execution_response_status_code"`
	ResponseStatus     string                      `gorm:"column:webhook_execution_response_status"`
	ResponseHeaders    string                      `gorm:"column:webhook_execution_response_headers"`
	ResponseBody       string                      `gorm:"column:webhook_execution_response_body"`
}

const (
	tableWhExec = "webhook_executions"
)

// Find finds the webhook execution by id.
func (s *WebhookExecutionStore) Find(ctx context.Context, id int64) (*types.WebhookExecution, error) {
	dst := &webhookExecution{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWhExec).First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
	}

	return mapToWebhookExecution(dst), nil
}

// Create creates a new webhook execution entry.
func (s *WebhookExecutionStore) Create(ctx context.Context, execution *types.WebhookExecution) error {
	dbObj := mapToInternalWebhookExecution(execution)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWhExec).Create(dbObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	execution.ID = dbObj.ID
	return nil
}

// DeleteOld removes all executions that are older than the provided time.
func (s *WebhookExecutionStore) DeleteOld(ctx context.Context, olderThan time.Time) (int64, error) {
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWhExec).
		Where("webhook_execution_created < ?", olderThan.UnixMilli()).
		Delete(nil)

	if res.Error != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, res.Error, "failed to execute delete executions query")
	}

	return res.RowsAffected, nil
}

// ListForWebhook lists the webhook executions for a given webhook id.
func (s *WebhookExecutionStore) ListForWebhook(ctx context.Context, webhookID int64,
	opts *types.WebhookExecutionFilter) ([]*types.WebhookExecution, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWhExec).
		Where("webhook_execution_webhook_id = ?", webhookID)

	stmt = stmt.Limit(int(database.Limit(opts.Size)))
	stmt = stmt.Offset(int(database.Offset(opts.Page, opts.Size)))

	// fixed ordering by desc id (new ones first) - add customized ordering if deemed necessary.
	stmt = stmt.Order("webhook_execution_id DESC")

	dst := []*webhookExecution{}
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
	}

	return mapToWebhookExecutions(dst), nil
}

// CountForWebhook counts the total number of webhook executions for a given webhook ID.
func (s *WebhookExecutionStore) CountForWebhook(
	ctx context.Context,
	webhookID int64,
) (int64, error) {
	var count int64
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWhExec).
		Where("webhook_execution_webhook_id = ?", webhookID).
		Count(&count).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Count query failed")
	}

	return count, nil
}

// ListForTrigger lists the webhook executions for a given trigger id.
func (s *WebhookExecutionStore) ListForTrigger(ctx context.Context,
	triggerID string) ([]*types.WebhookExecution, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableWhExec).
		Where("webhook_execution_trigger_id = ?", triggerID)

	dst := []*webhookExecution{}
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
	}

	return mapToWebhookExecutions(dst), nil
}

func mapToWebhookExecution(execution *webhookExecution) *types.WebhookExecution {
	return &types.WebhookExecution{
		ID:            execution.ID,
		RetriggerOf:   execution.RetriggerOf.Ptr(),
		Retriggerable: execution.Retriggerable,
		Created:       execution.Created,
		WebhookID:     execution.WebhookID,
		TriggerType:   execution.TriggerType,
		TriggerID:     execution.TriggerID,
		Result:        execution.Result,
		Error:         execution.Error,
		Duration:      execution.Duration,
		Request: types.WebhookExecutionRequest{
			URL:     execution.RequestURL,
			Headers: execution.RequestHeaders,
			Body:    execution.RequestBody,
		},
		Response: types.WebhookExecutionResponse{
			StatusCode: execution.ResponseStatusCode,
			Status:     execution.ResponseStatus,
			Headers:    execution.ResponseHeaders,
			Body:       execution.ResponseBody,
		},
	}
}

func mapToInternalWebhookExecution(execution *types.WebhookExecution) *webhookExecution {
	return &webhookExecution{
		ID:                 execution.ID,
		RetriggerOf:        null.IntFromPtr(execution.RetriggerOf),
		Retriggerable:      execution.Retriggerable,
		Created:            execution.Created,
		WebhookID:          execution.WebhookID,
		TriggerType:        execution.TriggerType,
		TriggerID:          execution.TriggerID,
		Result:             execution.Result,
		Error:              execution.Error,
		Duration:           execution.Duration,
		RequestURL:         execution.Request.URL,
		RequestHeaders:     execution.Request.Headers,
		RequestBody:        execution.Request.Body,
		ResponseStatusCode: execution.Response.StatusCode,
		ResponseStatus:     execution.Response.Status,
		ResponseHeaders:    execution.Response.Headers,
		ResponseBody:       execution.Response.Body,
	}
}

func mapToWebhookExecutions(executions []*webhookExecution) []*types.WebhookExecution {
	m := make([]*types.WebhookExecution, len(executions))
	for i, hook := range executions {
		m[i] = mapToWebhookExecution(hook)
	}

	return m
}
