// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"context"
	"errors"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ store.ArtifactTreeNodeInterface = (*treeNode)(nil)

type treeNode struct {
	db *gorm.DB
}

func (c *treeNode) Create(ctx context.Context, newObj *types.ArtifactTreeNode) error {
	if err := dbtx.GetOrmAccessor(ctx, c.db).Model(new(types.ArtifactTreeNode)).Create(newObj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "failed to create artifact tree node")
	}
	return nil
}

func (c *treeNode) RecurseCreate(ctx context.Context, obj *types.ArtifactTreeNode) error {
	var err error

	if err = c.getNode(ctx, obj); err == nil {
		return nil
	}

	if !errors.Is(err, gitfox_store.ErrResourceNotFound) {
		return err
	}

	if obj.IsRoot() {
		return c.Create(ctx, obj)
	}

	// check node parent existing
	p := obj.Parent()
	if err = c.RecurseCreate(ctx, p); err != nil {
		return err
	}

	obj.ParentID = null.IntFrom(p.ID)
	return c.Create(ctx, obj)
}

func (c *treeNode) GetById(ctx context.Context, id int64) (*types.ArtifactTreeNode, error) {
	var obj types.ArtifactTreeNode
	if err := dbtx.GetOrmAccessor(ctx, c.db).Take(&obj, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "")
	}
	return &obj, nil
}

func (c *treeNode) GetByPath(ctx context.Context, ownerId int64, path string, format types.ArtifactFormat) (*types.ArtifactTreeNode, error) {
	result := types.ArtifactTreeNode{
		OwnerID: ownerId,
		Path:    path,
		Format:  format,
	}
	if err := c.getNode(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *treeNode) getNode(ctx context.Context, obj *types.ArtifactTreeNode) error {
	q := types.ArtifactTreeNode{
		OwnerID: obj.OwnerID,
		Path:    obj.Path,
		Format:  obj.Format,
	}
	if err := dbtx.GetOrmAccessor(ctx, c.db).Where(&q).Take(obj).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "")
	}
	return nil
}

func (c *treeNode) ListFormats(ctx context.Context, ownerId int64) ([]*types.ArtifactTreeNode, error) {
	stmt := dbtx.GetOrmAccessor(ctx, c.db)
	q := types.ArtifactTreeNode{
		OwnerID: ownerId,
		Path:    "/",
	}

	var data []*types.ArtifactTreeNode
	if err := stmt.Where(&q).Order("node_format").Find(&data).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to list artifact tree nodes")
	}
	return data, nil
}

func (c *treeNode) ListByParentId(ctx context.Context, parentId int64) ([]*types.ArtifactTreeNode, error) {
	stmt := dbtx.GetOrmAccessor(ctx, c.db)

	if parentId == 0 {
		return nil, errors.New("parentId can not be zero")
	} else {
		stmt = stmt.Where("node_parent_id = ?", parentId)
	}

	var data []*types.ArtifactTreeNode
	if err := stmt.Order("node_name").Find(&data).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to list artifact packages")
	}
	return data, nil
}

func (c *treeNode) DeleteById(ctx context.Context, nodeId int64) error {
	result := dbtx.GetOrmAccessor(ctx, c.db).Delete(&types.ArtifactTreeNode{}, nodeId)
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "exec tree node delete failed")
	}

	if result.RowsAffected == 0 {
		return types.ErrPkgNoItemDeleted
	}
	return nil
}

func (c *treeNode) RecurseDeleteById(ctx context.Context, nodeId int64) error {
	subNodes, err := c.ListByParentId(ctx, nodeId)
	if err != nil {
		return err
	}
	if len(subNodes) > 0 {
		for _, subNode := range subNodes {
			err = c.RecurseDeleteById(ctx, subNode.ID)
			if err != nil {
				return err
			}
		}
	}

	err = c.DeleteById(ctx, nodeId)
	if err != nil {
		return err
	}
	return nil
}
