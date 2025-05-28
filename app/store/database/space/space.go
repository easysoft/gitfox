// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package space

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/paths"
	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/guregu/null"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ store.SpaceStore = (*OrmStore)(nil)

// NewSpaceOrmStore returns a new OrmStore.
func NewSpaceOrmStore(
	db *gorm.DB,
	spacePathCache store.SpacePathCache,
	spacePathStore store.SpacePathStore,
) *OrmStore {
	return &OrmStore{
		db:             db,
		spacePathCache: spacePathCache,
		spacePathStore: spacePathStore,
	}
}

// OrmStore implements a OrmStore backed by a relational database.
type OrmStore struct {
	db             *gorm.DB
	spacePathCache store.SpacePathCache
	spacePathStore store.SpacePathStore
}

// space is an internal representation used to store space data in DB.
type space struct {
	ID      int64 `gorm:"column:space_id;primaryKey"`
	Version int64 `gorm:"column:space_version"`
	// IMPORTANT: We need to make parentID optional for spaces to allow it to be a foreign key.
	ParentID    null.Int `gorm:"column:space_parent_id"`
	Identifier  string   `gorm:"column:space_uid"`
	Description string   `gorm:"column:space_description"`
	CreatedBy   int64    `gorm:"column:space_created_by"`
	Created     int64    `gorm:"column:space_created"`
	Updated     int64    `gorm:"column:space_updated"`
	Deleted     null.Int `gorm:"column:space_deleted"`
}

const (
	spaceTable = "spaces"

	spaceColumns = `
		space_id
		,space_version
		,space_parent_id
		,space_uid
		,space_description
		,space_created_by
		,space_created
		,space_updated
		,space_deleted`

	spaceSelectBase = `
	SELECT` + spaceColumns + `
	FROM spaces`
)

// Find the space by id.
func (s *OrmStore) Find(ctx context.Context, id int64) (*types.Space, error) {
	return s.find(ctx, id, nil)
}

func (s *OrmStore) find(ctx context.Context, id int64, deletedAt *int64) (*types.Space, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(spaceTable).
		Where("space_id = ?", id)

	if deletedAt != nil {
		stmt = stmt.Where("space_deleted = ?", *deletedAt)
	} else {
		stmt = stmt.Where("space_deleted IS NULL")
	}

	dst := new(space)

	if err := stmt.Take(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find space")
	}

	return mapToSpace(ctx, s.db, s.spacePathStore, dst)
}

// FindByRef finds the space using the spaceRef as either the id or the space path.
func (s *OrmStore) FindByRef(ctx context.Context, spaceRef string) (*types.Space, error) {
	return s.findByRef(ctx, spaceRef, nil)
}

// FindByRefCaseInsensitive finds the space using the spaceRef.
func (s *OrmStore) FindByRefCaseInsensitive(ctx context.Context, spaceRef string) (*types.Space, error) {
	segments := paths.Segments(spaceRef)
	if len(segments) < 1 {
		return nil, fmt.Errorf("invalid space reference provided")
	}

	db := s.db
	switch {
	case len(segments) == 1:
		db = db.Model(&space{}).Select("space_id").Where("LOWER(space_uid) = LOWER(?)", segments[0])

	case len(segments) > 1:
		db = buildRecursiveSelectQueryUsingCaseInsensitivePath(segments, db)
	}
	var spaceID int64
	if err := db.Take(&spaceID).Error; err != nil {
		return nil, fmt.Errorf("failed to find space: %w", err)
	}

	return s.find(ctx, spaceID, nil)
}

// FindByRefAndDeletedAt finds the space using the spaceRef as either the id or the space path and deleted timestamp.
func (s *OrmStore) FindByRefAndDeletedAt(
	ctx context.Context,
	spaceRef string,
	deletedAt int64,
) (*types.Space, error) {
	// ASSUMPTION: digits only is not a valid space path
	id, err := strconv.ParseInt(spaceRef, 10, 64)
	if err != nil {
		return s.findByPathAndDeletedAt(ctx, spaceRef, deletedAt)
	}

	return s.find(ctx, id, &deletedAt)
}

func (s *OrmStore) findByRef(ctx context.Context, spaceRef string, deletedAt *int64) (*types.Space, error) {
	// ASSUMPTION: digits only is not a valid space path
	id, err := strconv.ParseInt(spaceRef, 10, 64)
	if err != nil {
		var path *types.SpacePath
		path, err = s.spacePathCache.Get(ctx, spaceRef)
		if err != nil {
			return nil, fmt.Errorf("failed to get path: %w", err)
		}

		id = path.SpaceID
	}
	return s.find(ctx, id, deletedAt)
}

func (s *OrmStore) findByPathAndDeletedAt(
	ctx context.Context,
	spaceRef string,
	deletedAt int64,
) (*types.Space, error) {
	segments := paths.Segments(spaceRef)
	if len(segments) < 1 {
		return nil, fmt.Errorf("invalid space reference provided")
	}

	stmt := dbtx.GetOrmAccessor(ctx, s.db)
	switch {
	case len(segments) == 1:
		stmt = stmt.Table(spaceTable).
			Select("space_id").
			Where("space_uid = ? AND space_deleted = ? AND space_parent_id IS NULL", segments[0], deletedAt)

	case len(segments) > 1:
		stmt = buildRecursiveSelectQueryUsingPath(stmt, segments, deletedAt)
	}

	var spaceID int64
	if err := stmt.Take(&spaceID).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom select query")
	}

	return s.find(ctx, spaceID, &deletedAt)
}

const spaceAncestorsQuery = `
WITH RECURSIVE space_ancestors(space_ancestor_id, space_ancestor_uid, space_ancestor_parent_id) AS (
	SELECT space_id, space_uid, space_parent_id
	FROM spaces
	WHERE space_id = ?

	UNION

	SELECT space_id, space_uid, space_parent_id
	FROM spaces
	JOIN space_ancestors ON space_id = space_ancestor_parent_id
)
`

const spaceDescendantsQuery = `
WITH RECURSIVE space_descendants(space_descendant_id, space_descendant_uid, space_descendant_parent_id) AS (
	SELECT space_id, space_uid, space_parent_id
	FROM spaces
	WHERE space_id = ?

	UNION

	SELECT space_id, space_uid, space_parent_id
	FROM spaces
	JOIN space_descendants ON space_descendant_id = space_parent_id
)
`

// GetRootSpace returns a space where space_parent_id is NULL.
func (s *OrmStore) GetRootSpace(ctx context.Context, spaceID int64) (*types.Space, error) {
	query := spaceAncestorsQuery + `
		SELECT space_ancestor_id
		FROM space_ancestors
		WHERE space_ancestor_parent_id IS NULL`

	var rootID int64
	if err := s.db.Raw(query, spaceID).Scan(&rootID).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to get root space_id")
	}

	return s.Find(ctx, rootID)
}

func (s *OrmStore) GetAncestors(
	ctx context.Context,
	spaceID int64,
) ([]*types.Space, error) {
	query := spaceAncestorsQuery + `
		SELECT ` + spaceColumns + `
		FROM spaces INNER JOIN space_ancestors ON space_id = space_ancestor_id`

	var dst []*space
	if err := s.db.Raw(query, spaceID).Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing get space ancestors query")
	}

	return s.mapToSpaces(ctx, dst)
}

// GetAncestorIDs returns a list of all space IDs along the recursive path to the root space.
func (s *OrmStore) GetAncestorIDs(ctx context.Context, spaceID int64) ([]int64, error) {
	query := spaceAncestorsQuery + `
		SELECT space_ancestor_id FROM space_ancestors`

	var spaceIDs []int64
	if err := s.db.Raw(query, spaceID).Scan(&spaceIDs).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to get space ancestors IDs")
	}

	return spaceIDs, nil
}

// GetAncestorsData returns a list of space parent data for spaces that are ancestors of the space.
func (s *OrmStore) GetAncestorsData(ctx context.Context, spaceID int64) ([]types.SpaceParentData, error) {
	query := spaceAncestorsQuery + `
		SELECT space_ancestor_id, space_ancestor_uid, space_ancestor_parent_id FROM space_ancestors`

	return s.readParentsData(ctx, query, spaceID)
}

// GetDescendantsData returns a list of space parent data for spaces that are descendants of the space.
func (s *OrmStore) GetDescendantsData(ctx context.Context, spaceID int64) ([]types.SpaceParentData, error) {
	query := spaceDescendantsQuery + `
		SELECT space_descendant_id, space_descendant_uid, space_descendant_parent_id FROM space_descendants`

	return s.readParentsData(ctx, query, spaceID)
}

func (s *OrmStore) readParentsData(
	ctx context.Context,
	query string,
	spaceID int64,
) ([]types.SpaceParentData, error) {
	var result []types.SpaceParentData

	err := s.db.Raw(query, spaceID).Scan(&result).Error
	if err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to run space parent data query")
	}

	return result, nil
}

// Create a new space.
func (s *OrmStore) Create(ctx context.Context, space *types.Space) error {
	if space == nil {
		return errors.New("space is nil")
	}

	dbSpace := mapToInternalSpace(space)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(spaceTable).Create(dbSpace).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	space.ID = dbSpace.ID
	return nil
}

// Update updates the space details.
func (s *OrmStore) Update(ctx context.Context, spaceObj *types.Space) error {
	if spaceObj == nil {
		return errors.New("space is nil")
	}

	dbSpace := mapToInternalSpace(spaceObj)

	// update Version (used for optimistic locking) and Updated time
	dbSpace.Version++
	dbSpace.Updated = time.Now().UnixMilli()

	updateFields := []string{"Version", "Updated", "ParentID", "Identifier", "Description", "IsPublic", "Deleted"}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(spaceTable).
		Where(&space{ID: spaceObj.ID, Version: dbSpace.Version - 1}).
		Select(updateFields).Updates(dbSpace)

	count, err := res.RowsAffected, res.Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Update query failed")
	}

	if count == 0 {
		return gitfox_store.ErrVersionConflict
	}

	spaceObj.Version = dbSpace.Version
	spaceObj.Updated = dbSpace.Updated

	// update path in case parent/identifier changed
	spaceObj.Path, err = getSpacePath(ctx, s.db, s.spacePathStore, spaceObj.ID)
	if err != nil {
		return err
	}

	return nil
}

// updateOptLock updates the space using the optimistic locking mechanism.
func (s *OrmStore) updateOptLock(
	ctx context.Context,
	space *types.Space,
	mutateFn func(space *types.Space) error,
) (*types.Space, error) {
	for {
		dup := *space

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

		space, err = s.find(ctx, space.ID, space.Deleted)
		if err != nil {
			return nil, err
		}
	}
}

// UpdateOptLock updates the space using the optimistic locking mechanism.
func (s *OrmStore) UpdateOptLock(ctx context.Context,
	space *types.Space,
	mutateFn func(space *types.Space) error,
) (*types.Space, error) {
	for {
		dup := *space

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

		space, err = s.Find(ctx, space.ID)
		if err != nil {
			return nil, err
		}
	}
}

// UpdateDeletedOptLock updates a soft deleted space using the optimistic locking mechanism.
func (s *OrmStore) updateDeletedOptLock(
	ctx context.Context,
	space *types.Space,
	mutateFn func(space *types.Space) error,
) (*types.Space, error) {
	return s.updateOptLock(
		ctx,
		space,
		func(r *types.Space) error {
			if space.Deleted == nil {
				return gitfox_store.ErrResourceNotFound
			}
			return mutateFn(r)
		},
	)
}

// FindForUpdate finds the space and locks it for an update (should be called in a tx).
func (s *OrmStore) FindForUpdate(ctx context.Context, id int64) (*types.Space, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(spaceTable).
		Where("space_id = ? AND space_deleted IS NULL", id).
		Clauses(clause.Locking{Strength: "UPDATE"})

	dst := new(space)
	if err := stmt.Take(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find space")
	}

	return mapToSpace(ctx, s.db, s.spacePathStore, dst)
}

// SoftDelete deletes a space softly.
func (s *OrmStore) SoftDelete(
	ctx context.Context,
	space *types.Space,
	deletedAt int64,
) error {
	_, err := s.UpdateOptLock(ctx, space, func(s *types.Space) error {
		s.Deleted = &deletedAt
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Purge deletes a space permanently.
func (s *OrmStore) Purge(ctx context.Context, id int64, deletedAt *int64) error {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(spaceTable).
		Where("space_id = ?", id)

	if deletedAt != nil {
		stmt = stmt.Where("space_deleted = ?", *deletedAt)
	} else {
		stmt = stmt.Where("space_deleted IS NULL")
	}

	if err := stmt.Delete(nil).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "the delete query failed")
	}

	return nil
}

// Restore restores a soft deleted space.
func (s *OrmStore) Restore(
	ctx context.Context,
	space *types.Space,
	newIdentifier *string,
	newParentID *int64,
) (*types.Space, error) {
	space, err := s.updateDeletedOptLock(ctx, space, func(s *types.Space) error {
		s.Deleted = nil
		if newParentID != nil {
			s.ParentID = *newParentID
		}

		if newIdentifier != nil {
			s.Identifier = *newIdentifier
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return space, nil
}

// Count the child spaces of a space.
func (s *OrmStore) Count(ctx context.Context, id int64, opts *types.SpaceFilter) (int64, error) {
	if opts.Recursive {
		return s.countAll(ctx, id, opts)
	}
	return s.count(ctx, id, opts)
}

func (s *OrmStore) count(
	ctx context.Context,
	id int64,
	opts *types.SpaceFilter,
) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(spaceTable)
	if id == 0 {
		// allow list root spaces for gitfox custom api
		stmt = stmt.Where("space_parent_id IS NULL")
	} else {
		stmt = stmt.Where("space_parent_id = ?", id)
	}

	if opts.Query != "" {
		stmt = stmt.Where("LOWER(space_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(opts.Query)))
	}

	stmt = s.applyQueryFilter(stmt, opts)

	var count int64
	if err := stmt.Count(&count).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}

	return count, nil
}

func (s *OrmStore) countAll(
	ctx context.Context,
	id int64,
	opts *types.SpaceFilter,
) (int64, error) {
	// todo: support space_id is null for gitfox ListAllSpaces api
	ctePrefix := `WITH RECURSIVE SpaceHierarchy AS (
		SELECT space_id, space_parent_id, space_deleted, space_uid
		FROM spaces
		WHERE space_id = ?

		UNION

		SELECT s.space_id, s.space_parent_id, s.space_deleted, s.space_uid
		FROM spaces s
		JOIN SpaceHierarchy h ON s.space_parent_id = h.space_id
	)`

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Clauses(
		clause.Clause{
			BeforeExpression: clause.Expr{SQL: ctePrefix, Vars: []interface{}{id}},
			Expression: clause.Expr{
				SQL:  "SELECT count(*) FROM SpaceHierarchy h1 WHERE h1.space_id <> ?",
				Vars: []interface{}{id},
			},
		},
	)

	stmt = s.applyQueryFilter(stmt, opts)

	var count int64
	if err := stmt.Raw(trimSQLForWithRecursive(stmt)).Scan(&count).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "failed to count sub spaces")
	}

	return count, nil
}

func trimSQLForWithRecursive(stmt *gorm.DB) string {
	sql := stmt.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Scan(nil)
	})

	point := strings.Index(sql, "WITH RECURSIVE")
	return sql[point:]
}

// List returns a list of spaces under the parent space.
func (s *OrmStore) List(ctx context.Context, id int64, opts *types.SpaceFilter) ([]*types.Space, error) {
	if opts.Recursive {
		return s.listAll(ctx, id, opts)
	}
	return s.list(ctx, id, opts)
}

func (s *OrmStore) list(
	ctx context.Context,
	id int64,
	opts *types.SpaceFilter,
) ([]*types.Space, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(spaceTable)
	if id == 0 {
		// allow list root spaces for gitfox custom api
		stmt = stmt.Where("space_parent_id IS NULL")
	} else {
		stmt = stmt.Where("space_parent_id = ?", id)
	}

	stmt = s.applyQueryFilter(stmt, opts)
	stmt = s.applySortFilter(stmt, opts)

	var dst []*space
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return s.mapToSpaces(ctx, dst)
}

func (s *OrmStore) listAll(ctx context.Context,
	id int64,
	opts *types.SpaceFilter,
) ([]*types.Space, error) {
	// todo: support space_id is null for gitfox ListAllSpaces api
	ctePrefix := `WITH RECURSIVE SpaceHierarchy AS (
		SELECT *
		FROM spaces
		WHERE space_id = ?

		UNION

		SELECT s.*
		FROM spaces s
		JOIN SpaceHierarchy h ON s.space_parent_id = h.space_id
	)`

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Clauses(
		clause.Clause{
			BeforeExpression: clause.Expr{SQL: ctePrefix, Vars: []interface{}{id}},
			Expression: clause.Expr{
				SQL:  "SELECT * FROM SpaceHierarchy h1 WHERE h1.space_id <> ?",
				Vars: []interface{}{id},
			},
		},
	)

	stmt = s.applyQueryFilter(stmt, opts)
	stmt = s.applySortFilter(stmt, opts)

	var dst []*space
	if err := stmt.Raw(trimSQLForWithRecursive(stmt)).Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return s.mapToSpaces(ctx, dst)
}

func (s *OrmStore) applyQueryFilter(
	stmt *gorm.DB,
	opts *types.SpaceFilter,
) *gorm.DB {
	if opts.Query != "" {
		stmt = stmt.Where("LOWER(space_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(opts.Query)))
	}

	if opts.DeletedBeforeOrAt != nil {
		stmt = stmt.Where("space_deleted <= ?", opts.DeletedBeforeOrAt)
	} else {
		stmt = stmt.Where("space_deleted IS NULL")
	}

	return stmt
}

func (s *OrmStore) applySortFilter(
	stmt *gorm.DB,
	opts *types.SpaceFilter,
) *gorm.DB {
	stmt = stmt.Limit(int(database.Limit(opts.Size)))
	stmt = stmt.Offset(int(database.Offset(opts.Page, opts.Size)))

	switch opts.Sort {
	case enum.SpaceAttrUID, enum.SpaceAttrIdentifier, enum.SpaceAttrNone:
		// NOTE: string concatenation is safe because the
		// order attribute is an enum and is not user-defined,
		// and is therefore not subject to injection attacks.
		stmt = stmt.Order("space_uid " + opts.Order.String())
		//TODO: Postgres does not support COLLATE NOCASE for UTF8
		// stmt = stmt.OrderBy("space_uid COLLATE NOCASE " + opts.Order.String())
	case enum.SpaceAttrCreated:
		stmt = stmt.Order("space_created " + opts.Order.String())
	case enum.SpaceAttrUpdated:
		stmt = stmt.Order("space_updated " + opts.Order.String())
	case enum.SpaceAttrDeleted:
		stmt = stmt.Order("space_deleted " + opts.Order.String())
	}
	return stmt
}

func mapToSpace(
	ctx context.Context,
	gormdb *gorm.DB,
	spacePathStore store.SpacePathStore,
	in *space,
) (*types.Space, error) {
	var err error
	res := &types.Space{
		ID:          in.ID,
		Version:     in.Version,
		Identifier:  in.Identifier,
		Description: in.Description,
		Created:     in.Created,
		CreatedBy:   in.CreatedBy,
		Updated:     in.Updated,
		Deleted:     in.Deleted.Ptr(),
	}

	// Only overwrite ParentID if it's not a root space
	if in.ParentID.Valid {
		res.ParentID = in.ParentID.Int64
	}

	// backfill path
	res.Path, err = getSpacePath(ctx, gormdb, spacePathStore, in.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get primary path for space %d: %w", in.ID, err)
	}

	return res, nil
}

func getSpacePath(
	ctx context.Context,
	gormdb *gorm.DB,
	spacePathStore store.SpacePathStore,
	spaceID int64,
) (string, error) {
	spacePath, err := spacePathStore.FindPrimaryBySpaceID(ctx, spaceID)
	// delete space will delete paths; generate the path if space is soft deleted.
	if errors.Is(err, gitfox_store.ErrResourceNotFound) {
		return getPathForDeletedSpace(ctx, gormdb, spaceID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get primary path for space %d: %w", spaceID, err)
	}

	return spacePath.Value, nil
}

func (s *OrmStore) mapToSpaces(
	ctx context.Context,
	spaces []*space,
) ([]*types.Space, error) {
	var err error
	res := make([]*types.Space, len(spaces))
	for i := range spaces {
		res[i], err = mapToSpace(ctx, s.db, s.spacePathStore, spaces[i])
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// List returns a list of spaces under the parent space.
func (s *OrmStore) ListAll(ctx context.Context, opts *types.SpaceFilter) ([]*types.Space, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(spaceTable)
	stmt = s.applyQueryFilter(stmt, opts)
	stmt = s.applySortFilter(stmt, opts)

	var dst []*space
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return s.mapToSpaces(ctx, dst)
}

func mapToInternalSpace(s *types.Space) *space {
	res := &space{
		ID:          s.ID,
		Version:     s.Version,
		Identifier:  s.Identifier,
		Description: s.Description,
		Created:     s.Created,
		CreatedBy:   s.CreatedBy,
		Updated:     s.Updated,
		Deleted:     null.IntFromPtr(s.Deleted),
	}

	// Only overwrite ParentID if it's not a root space
	// IMPORTANT: s.ParentID==0 has to be translated to nil as otherwise the foreign key fails
	if s.ParentID > 0 {
		res.ParentID = null.IntFrom(s.ParentID)
	}

	return res
}

func getPathForDeletedSpace(
	ctx context.Context,
	gormdb *gorm.DB,
	id int64,
) (string, error) {
	path := ""
	nextSpaceID := null.IntFrom(id)

	db := dbtx.GetOrmAccessor(ctx, gormdb).Table(spaceTable)
	dst := new(space)

	for nextSpaceID.Valid {
		err := db.Where("space_id = ?", nextSpaceID.Int64).Error
		if err != nil {
			return "", fmt.Errorf("failed to find the space %d: %w", id, err)
		}

		path = paths.Concatenate(dst.Identifier, path)
		nextSpaceID = dst.ParentID
	}

	return path, nil
}

// buildRecursiveSelectQueryUsingPath builds the recursive select query using path among active or soft deleted spaces.
func buildRecursiveSelectQueryUsingPath(stmt *gorm.DB, segments []string, deletedAt int64) *gorm.DB {
	leaf := "s" + strconv.Itoa(len(segments)-1)

	// add the current space (leaf)
	stmt = stmt.
		Select(leaf+".space_id").
		Table(spaceTable+" "+leaf).
		Where(leaf+".space_uid = ? AND "+leaf+".space_deleted = ?", segments[len(segments)-1], deletedAt)

	for i := len(segments) - 2; i >= 0; i-- {
		parentAlias := "s" + strconv.Itoa(i)
		alias := "s" + strconv.Itoa(i+1)

		stmt = stmt.InnerJoins(fmt.Sprintf("spaces %s ON %s.space_id = %s.space_parent_id", parentAlias, parentAlias, alias)).
			Where(parentAlias+".space_uid = ?", segments[i])
	}

	// add parent check for root
	stmt = stmt.Where("s0.space_parent_id IS NULL")

	return stmt
}

// buildRecursiveSelectQueryUsingCaseInsensitivePath builds the recursive select query using path among active or soft
// deleted spaces.
func buildRecursiveSelectQueryUsingCaseInsensitivePath(segments []string, db *gorm.DB) *gorm.DB {
	leaf := "s" + strconv.Itoa(len(segments)-1)

	// add the current space (leaf)
	stmt := db.Table("spaces "+leaf).
		Select(leaf+".space_id").
		Where("LOWER("+leaf+".space_uid) = LOWER(?)", segments[len(segments)-1])

	for i := len(segments) - 2; i >= 0; i-- {
		parentAlias := "s" + strconv.Itoa(i)
		alias := "s" + strconv.Itoa(i+1)

		stmt = stmt.Joins(fmt.Sprintf("INNER JOIN spaces %s ON %s.space_id = %s.space_parent_id", parentAlias, parentAlias,
			alias)).
			Where(parentAlias+".space_uid = ?", segments[i])
	}

	// add parent check for root
	stmt = stmt.Where("s0.space_parent_id IS NULL")

	return stmt
}
