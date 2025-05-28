// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package space

import (
	"context"
	"errors"
	"fmt"

	"github.com/easysoft/gitfox/app/paths"
	"github.com/easysoft/gitfox/app/store"
	store2 "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ store.SpacePathStore = (*SpacePathStore)(nil)

// NewSpacePathOrmStore returns a new PathStore.
func NewSpacePathOrmStore(db *gorm.DB, pathTransformation store.SpacePathTransformation) *SpacePathStore {
	return &SpacePathStore{
		db:                      db,
		spacePathTransformation: pathTransformation,
	}
}

// SpacePathStore implements a store.SpacePathStore backed by a relational database.
type SpacePathStore struct {
	db                      *gorm.DB
	spacePathTransformation store.SpacePathTransformation
}

// spacePathSegment is an internal representation of a segment of a space path.
type spacePathSegment struct {
	ID int64 `gorm:"column:space_path_id"`
	// Identifier is the original identifier that was provided
	Identifier string `gorm:"column:space_path_uid"`
	// IdentifierUnique is a transformed version of Identifier which is used to ensure uniqueness guarantees
	IdentifierUnique string `gorm:"column:space_path_uid_unique"`
	// IsPrimary indicates whether the path is the primary path of the space
	// IMPORTANT: to allow DB enforcement of at most one primary path per repo/space
	// we have a unique index on spaceID + IsPrimary and set IsPrimary to true
	// for primary paths and to nil for non-primary paths.
	IsPrimary null.Bool `gorm:"column:space_path_is_primary"`
	ParentID  null.Int  `gorm:"column:space_path_parent_id"`
	SpaceID   int64     `gorm:"column:space_path_space_id"`
	CreatedBy int64     `gorm:"column:space_path_created_by"`
	Created   int64     `gorm:"column:space_path_created"`
	Updated   int64     `gorm:"column:space_path_updated"`
}

const (
	spacePathTable = "space_paths"
)

// InsertSegment inserts a space path segment to the table - returns the full path.
func (s *SpacePathStore) InsertSegment(ctx context.Context, segment *types.SpacePathSegment) error {
	dbSegment := s.mapToInternalSpacePathSegment(segment)

	db := dbtx.GetOrmAccessor(ctx, s.db)

	existObj := &spacePathSegment{}
	var found bool
	stmt := db.WithContext(ctx).Table(spacePathTable).Where(&spacePathSegment{Identifier: dbSegment.Identifier})
	if dbSegment.ParentID.IsZero() {
		stmt = stmt.Where("space_path_parent_id IS NULL")
	} else {
		stmt = stmt.Where("space_path_parent_id = ?", dbSegment.ParentID.Int64)
	}

	if err := stmt.First(existObj).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return database.ProcessGormSQLErrorf(ctx, err, "Check space_path conflict failed")
		}
	} else {
		found = true
	}

	if found {
		return store2.ErrDuplicate
	}

	if err := db.WithContext(ctx).Table(spacePathTable).Create(dbSegment).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	segment.ID = dbSegment.ID
	return nil
}

func (s *SpacePathStore) FindPrimaryBySpaceID(ctx context.Context, spaceID int64) (*types.SpacePath, error) {
	path := ""
	nextSpaceID := null.IntFrom(spaceID)

	for nextSpaceID.Valid {
		dst := new(spacePathSegment)
		err := dbtx.GetOrmAccessor(ctx, s.db).Table(spacePathTable).
			Where(&spacePathSegment{IsPrimary: null.BoolFrom(true), SpaceID: nextSpaceID.Int64}).
			Take(&dst).Error
		if err != nil {
			return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find primary segment for %d", nextSpaceID.Int64)
		}
		path = paths.Concatenate(dst.Identifier, path)
		nextSpaceID = dst.ParentID
	}

	return &types.SpacePath{
		SpaceID:   spaceID,
		Value:     path,
		IsPrimary: true,
	}, nil
}
func (s *SpacePathStore) FindByPath(ctx context.Context, path string) (*types.SpacePath, error) {
	segmentIdentifiers := paths.Segments(path)
	if len(segmentIdentifiers) == 0 {
		return nil, fmt.Errorf("path with no segments was passed '%s'", path)
	}

	var err error
	var parentID int64
	originalPath := ""
	isPrimary := true
	for i, segmentIdentifier := range segmentIdentifiers {
		segment := new(spacePathSegment)
		uniqueSegmentIdentifier := s.spacePathTransformation(segmentIdentifier, i == 0)

		if parentID == 0 {
			err = dbtx.GetOrmAccessor(ctx, s.db).Table(spacePathTable).Where("space_path_parent_id IS NULL").
				Where(&spacePathSegment{IdentifierUnique: uniqueSegmentIdentifier}).First(segment).Error
		} else {
			err = dbtx.GetOrmAccessor(ctx, s.db).Table(spacePathTable).
				Where(&spacePathSegment{
					IdentifierUnique: uniqueSegmentIdentifier, ParentID: null.IntFrom(parentID)},
				).First(segment).Error
		}
		if err != nil {
			return nil, database.ProcessGormSQLErrorf(
				ctx,
				err,
				"Failed to find segment for '%s' in '%s'",
				uniqueSegmentIdentifier,
				path,
			)
		}

		originalPath = paths.Concatenate(originalPath, segment.Identifier)
		parentID = segment.SpaceID
		isPrimary = isPrimary && segment.IsPrimary.ValueOrZero()
	}

	return &types.SpacePath{
		Value:     originalPath,
		IsPrimary: isPrimary,
		SpaceID:   parentID,
	}, nil
}

// DeletePrimarySegment deletes the primary segment of the space.
func (s *SpacePathStore) DeletePrimarySegment(ctx context.Context, spaceID int64) error {
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(spacePathTable).
		Where(&spacePathSegment{IsPrimary: null.BoolFrom(true), SpaceID: spaceID}).
		Delete(&spacePathSegment{})

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "the delete query failed")
	}

	return nil
}

// DeletePathsAndDescendandPaths deletes all space paths reachable from spaceID including itself.
func (s *SpacePathStore) DeletePathsAndDescendandPaths(ctx context.Context, spaceID int64) error {
	const sqlQuery = `WITH RECURSIVE DescendantPaths AS (
		SELECT space_path_id, space_path_space_id, space_path_parent_id
		FROM space_paths
		WHERE space_path_space_id = ?

		UNION

		SELECT sp.space_path_id, sp.space_path_space_id, sp.space_path_parent_id
		FROM space_paths sp
		JOIN DescendantPaths dp ON sp.space_path_parent_id = dp.space_path_space_id
	  )
	  DELETE FROM space_paths
	  WHERE space_path_id IN (SELECT space_path_id FROM DescendantPaths);`

	db := dbtx.GetOrmAccessor(ctx, s.db)

	if err := db.Exec(sqlQuery, spaceID).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "the delete query failed")
	}

	return nil
}

func (s *SpacePathStore) mapToInternalSpacePathSegment(p *types.SpacePathSegment) *spacePathSegment {
	res := &spacePathSegment{
		ID:               p.ID,
		Identifier:       p.Identifier,
		IdentifierUnique: s.spacePathTransformation(p.Identifier, p.ParentID == 0),
		SpaceID:          p.SpaceID,
		Created:          p.Created,
		CreatedBy:        p.CreatedBy,
		Updated:          p.Updated,

		// ParentID:  is set below
		// IsPrimary: is set below
	}

	// only set IsPrimary to a value if it's true (Unique Index doesn't allow multiple false, hence keep it nil)
	if p.IsPrimary {
		res.IsPrimary = null.BoolFrom(true)
	}

	if p.ParentID > 0 {
		res.ParentID = null.IntFrom(p.ParentID)
	}

	return res
}
