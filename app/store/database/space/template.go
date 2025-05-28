// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package space

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

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ store.TemplateStore = (*TemplateStore)(nil)

const (
	tableTpl = "templates"
)

// NewTemplateOrmStore returns a new TemplateStore.
func NewTemplateOrmStore(db *gorm.DB) *TemplateStore {
	return &TemplateStore{
		db: db,
	}
}

type TemplateStore struct {
	db *gorm.DB
}

// Find returns a template given a template ID.
func (s *TemplateStore) Find(ctx context.Context, id int64) (*types.Template, error) {
	dst := new(types.Template)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTpl).First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find template")
	}
	return dst, nil
}

// FindByIdentifierAndType returns a template in a space with a given identifier and a given type.
func (s *TemplateStore) FindByIdentifierAndType(
	ctx context.Context,
	spaceID int64,
	identifier string,
	resolverType enum.ResolverType) (*types.Template, error) {
	q := types.Template{SpaceID: spaceID, Identifier: identifier, Type: resolverType}

	dst := new(types.Template)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTpl).Where(&q).Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find template")
	}
	return dst, nil
}

// Create creates a template.
func (s *TemplateStore) Create(ctx context.Context, template *types.Template) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTpl).Create(template).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "template query failed")
	}

	return nil
}

func (s *TemplateStore) Update(ctx context.Context, p *types.Template) error {
	updatedAt := time.Now()
	template := *p

	template.Version++
	template.Updated = updatedAt.UnixMilli()

	updateFields := []string{"Description", "Identifier", "Data", "Type", "Updated", "Version"}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTpl).
		Where(&types.Template{ID: p.ID, Version: template.Version - 1}).
		Select(updateFields).Updates(&template)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update template")
	}

	count := res.RowsAffected

	if count == 0 {
		return gitfox_store.ErrVersionConflict
	}

	p.Version = template.Version
	p.Updated = template.Updated
	return nil
}

// UpdateOptLock updates the pipeline using the optimistic locking mechanism.
func (s *TemplateStore) UpdateOptLock(ctx context.Context,
	template *types.Template,
	mutateFn func(template *types.Template) error,
) (*types.Template, error) {
	for {
		dup := *template

		err := mutateFn(&dup)
		if err != nil {
			return nil, err
		}

		err = s.Update(ctx, &dup)
		if err == nil {
			return &dup, nil
		}
		if !errors.Is(err, gitfox_store.ErrVersionConflict) {
			return nil, err
		}

		template, err = s.Find(ctx, template.ID)
		if err != nil {
			return nil, err
		}
	}
}

// List lists all the templates present in a space.
func (s *TemplateStore) List(
	ctx context.Context,
	parentID int64,
	filter types.ListQueryFilter,
) ([]*types.Template, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTpl).
		Where("template_space_id = ?", parentID)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(template_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	stmt = stmt.Limit(int(database.Limit(filter.Size)))
	stmt = stmt.Offset(int(database.Offset(filter.Page, filter.Size)))

	dst := []*types.Template{}
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return dst, nil
}

// Delete deletes a template given a template ID.
func (s *TemplateStore) Delete(ctx context.Context, id int64) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTpl).Where(&types.Template{ID: id}).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete template")
	}

	return nil
}

// DeleteByIdentifierAndType deletes a template with a given identifier in a space.
func (s *TemplateStore) DeleteByIdentifierAndType(
	ctx context.Context,
	spaceID int64,
	identifier string,
	resolverType enum.ResolverType,
) error {
	q := types.Template{SpaceID: spaceID, Identifier: identifier, Type: resolverType}

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTpl).Where(&q).Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete template")
	}

	return nil
}

// Count of templates in a space.
func (s *TemplateStore) Count(ctx context.Context, parentID int64, filter types.ListQueryFilter) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(tableTpl).
		Where("template_space_id = ?", parentID)

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(template_uid) LIKE ?", fmt.Sprintf("%%%s%%", filter.Query))
	}

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}
