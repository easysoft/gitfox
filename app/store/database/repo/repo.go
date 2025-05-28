// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo

import (
	"cmp"
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
)

var _ store.RepoStore = (*OrmStore)(nil)

// NewRepoOrmStore returns a new OrmStore.
func NewRepoOrmStore(
	db *gorm.DB,
	spacePathCache store.SpacePathCache,
	spacePathStore store.SpacePathStore,
	spaceStore store.SpaceStore,
) *OrmStore {
	return &OrmStore{
		db:             db,
		spacePathCache: spacePathCache,
		spacePathStore: spacePathStore,
		spaceStore:     spaceStore,
	}
}

// OrmStore implements a store.RepoStore backed by a relational database.
type OrmStore struct {
	db             *gorm.DB
	spacePathCache store.SpacePathCache
	spacePathStore store.SpacePathStore
	spaceStore     store.SpaceStore
}

type repository struct {
	// TODO: int64 ID doesn't match DB
	ID          int64    `gorm:"column:repo_id;primaryKey"`
	Version     int64    `gorm:"column:repo_version"`
	ParentID    int64    `gorm:"column:repo_parent_id"`
	Identifier  string   `gorm:"column:repo_uid"`
	Description string   `gorm:"column:repo_description"`
	CreatedBy   int64    `gorm:"column:repo_created_by"`
	Created     int64    `gorm:"column:repo_created"`
	Updated     int64    `gorm:"column:repo_updated"`
	Deleted     null.Int `gorm:"column:repo_deleted"`

	Size        int64 `gorm:"column:repo_size"`
	SizeUpdated int64 `gorm:"column:repo_size_updated"`

	GitUID        string `gorm:"column:repo_git_uid"`
	DefaultBranch string `gorm:"column:repo_default_branch"`
	ForkID        int64  `gorm:"column:repo_fork_id"`
	PullReqSeq    int64  `gorm:"column:repo_pullreq_seq"`

	NumForks       int `gorm:"column:repo_num_forks"`
	NumPulls       int `gorm:"column:repo_num_pulls"`
	NumClosedPulls int `gorm:"column:repo_num_closed_pulls"`
	NumOpenPulls   int `gorm:"column:repo_num_open_pulls"`
	NumMergedPulls int `gorm:"column:repo_num_merged_pulls"`

	Mirror  bool           `gorm:"column:repo_mirror"`
	State   enum.RepoState `gorm:"column:repo_state"`
	IsEmpty bool           `gorm:"column:repo_is_empty"`
}

const (
	repoTable = "repositories"
)

const spaceDescendantsQuery = `
WITH RECURSIVE space_descendants(space_descendant_id, space_descendant_uid, space_descendant_parent_id) AS (
	SELECT space_id, space_uid, space_parent_id
	FROM spaces
	WHERE space_id IN ?

	UNION

	SELECT space_id, space_uid, space_parent_id
	FROM spaces
	JOIN space_descendants ON space_descendant_id = space_parent_id
)
`

// Find finds the repo by id.
func (s *OrmStore) Find(ctx context.Context, id int64) (*types.Repository, error) {
	return s.find(ctx, id, nil)
}

// find is a wrapper to find a repo by id w/o deleted timestamp.
func (s *OrmStore) find(ctx context.Context, id int64, deletedAt *int64) (*types.Repository, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable)

	if deletedAt != nil {
		stmt = stmt.Where("repo_deleted = ?", *deletedAt)
	} else {
		stmt = stmt.Where("repo_deleted IS NULL")
	}

	dst := new(repository)

	if err := stmt.First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find repo")
	}

	return s.mapToRepo(ctx, dst)
}

func (s *OrmStore) findByIdentifier(
	ctx context.Context,
	spaceID int64,
	identifier string,
	deletedAt *int64,
) (*types.Repository, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).
		Where("repo_parent_id = ? AND LOWER(repo_uid) = ?", spaceID, strings.ToLower(identifier))

	if deletedAt != nil {
		stmt = stmt.Where("repo_deleted = ?", *deletedAt)
	} else {
		stmt = stmt.Where("repo_deleted IS NULL")
	}

	dst := new(repository)

	if err := stmt.Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find repo")
	}

	return s.mapToRepo(ctx, dst)
}

func (s *OrmStore) findByRef(ctx context.Context, repoRef string, deletedAt *int64) (*types.Repository, error) {
	// ASSUMPTION: digits only is not a valid repo path
	id, err := strconv.ParseInt(repoRef, 10, 64)
	if err != nil {
		spacePath, repoIdentifier, err := paths.DisectLeaf(repoRef)
		if err != nil {
			return nil, fmt.Errorf("failed to disect leaf for path '%s': %w", repoRef, err)
		}
		pathObject, err := s.spacePathCache.Get(ctx, spacePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get space path: %w", err)
		}

		return s.findByIdentifier(ctx, pathObject.SpaceID, repoIdentifier, deletedAt)
	}
	return s.find(ctx, id, deletedAt)
}

// FindByRef finds the repo using the repoRef as either the id or the repo path.
func (s *OrmStore) FindByRef(ctx context.Context, repoRef string) (*types.Repository, error) {
	return s.findByRef(ctx, repoRef, nil)
}

// FindByRefAndDeletedAt finds the repo using the repoRef and deleted timestamp.
func (s *OrmStore) FindByRefAndDeletedAt(
	ctx context.Context,
	repoRef string,
	deletedAt int64,
) (*types.Repository, error) {
	return s.findByRef(ctx, repoRef, &deletedAt)
}

func (s *OrmStore) exist(ctx context.Context, identifier string, parentId int64) (error, bool) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).Where(&repository{
		Identifier: identifier, ParentID: parentId,
	}).Where("repo_deleted IS NULL")
	var dst repository
	if err := stmt.Take(&dst).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false
		} else {
			return err, false
		}
	}
	return nil, true
}

// Create creates a new repository.
func (s *OrmStore) Create(ctx context.Context, repo *types.Repository) error {
	err, exist := s.exist(ctx, repo.Identifier, repo.ParentID)
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Check conflict failed")
	}
	if exist {
		return gitfox_store.ErrDuplicate
	}

	dbRepo := mapToInternalRepo(repo)

	if err = dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).Create(dbRepo).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	repo.Path, err = s.getRepoPath(ctx, repo.ParentID, repo.Identifier)
	if err != nil {
		return err
	}

	repo.ID = dbRepo.ID
	return nil
}

// Update updates the repo details.
func (s *OrmStore) Update(ctx context.Context, repo *types.Repository) error {
	var err error
	dbRepo := mapToInternalRepo(repo)

	// update Version (used for optimistic locking) and Updated time
	dbRepo.Version++
	dbRepo.Updated = time.Now().UnixMilli()

	updateFields := []string{"Version", "Updated", "Deleted", "ParentID",
		"Identifier", "GitUID", "Description", "DefaultBranch",
		"PullReqSeq", "NumForks", "NumPulls", "NumClosedPulls",
		"NumOpenPulls", "NumMergedPulls", "State", "Mirror", "IsEmpty"}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).
		Where(&repository{ID: repo.ID, Version: dbRepo.Version - 1}).
		Select(updateFields).Updates(dbRepo)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update repository")
	}

	count := res.RowsAffected

	if count == 0 {
		return gitfox_store.ErrVersionConflict
	}

	repo.Version = dbRepo.Version
	repo.Updated = dbRepo.Updated

	// update path in case parent/identifier changed (its most likely cached anyway)
	repo.Path, err = s.getRepoPath(ctx, repo.ParentID, repo.Identifier)
	if err != nil {
		return err
	}

	return nil
}

// UpdateSize updates the size of a specific repository in the database.
func (s *OrmStore) UpdateSize(ctx context.Context, id int64, size int64) error {
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).
		Where("repo_id = ? AND repo_deleted IS NULL", id).
		Updates(map[string]interface{}{"repo_size": size, "repo_size_updated": time.Now().UnixMilli()})

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update repo size")
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("repo %d size not updated: %w", id, gitfox_store.ErrResourceNotFound)
	}

	return nil
}

// GetSize returns the repo size.
func (s *OrmStore) GetSize(ctx context.Context, id int64) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).Select("repo_size").Where("repo_deleted IS NULL")

	dst := new(repository)
	if err := stmt.First(dst, id).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "failed to get repo size")
	}
	return dst.Size, nil
}

// UpdateOptLock updates the active repository using the optimistic locking mechanism.
func (s *OrmStore) UpdateOptLock(
	ctx context.Context,
	repo *types.Repository,
	mutateFn func(repository *types.Repository) error,
) (*types.Repository, error) {
	return s.updateOptLock(
		ctx,
		repo,
		func(r *types.Repository) error {
			if repo.Deleted != nil {
				return gitfox_store.ErrResourceNotFound
			}
			return mutateFn(r)
		},
	)
}

// UpdateDeletedOptLock updates a deleted repository using the optimistic locking mechanism.
func (s *OrmStore) updateDeletedOptLock(ctx context.Context,
	repo *types.Repository,
	mutateFn func(repository *types.Repository) error,
) (*types.Repository, error) {
	return s.updateOptLock(
		ctx,
		repo,
		func(r *types.Repository) error {
			if repo.Deleted == nil {
				return gitfox_store.ErrResourceNotFound
			}
			return mutateFn(r)
		},
	)
}

// updateOptLock updates the repository using the optimistic locking mechanism.
func (s *OrmStore) updateOptLock(
	ctx context.Context,
	repo *types.Repository,
	mutateFn func(repository *types.Repository) error,
) (*types.Repository, error) {
	for {
		dup := *repo

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

		repo, err = s.find(ctx, repo.ID, repo.Deleted)
		if err != nil {
			return nil, err
		}
	}
}

// SoftDelete deletes a repo softly by setting the deleted timestamp.
func (s *OrmStore) SoftDelete(ctx context.Context, repo *types.Repository, deletedAt int64) error {
	_, err := s.UpdateOptLock(ctx, repo, func(r *types.Repository) error {
		r.Deleted = &deletedAt
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to soft delete repo: %w", err)
	}
	return nil
}

// Purge deletes the repo permanently.
func (s *OrmStore) Purge(ctx context.Context, id int64, deletedAt *int64) error {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable)

	if deletedAt != nil {
		stmt = stmt.Where("repo_deleted = ?", *deletedAt)
	} else {
		stmt = stmt.Where("repo_deleted IS NULL")
	}

	err := stmt.Delete(&repository{}, id).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "the delete query failed")
	}

	return nil
}

// Restore restores a deleted repo.
func (s *OrmStore) Restore(
	ctx context.Context,
	repo *types.Repository,
	newIdentifier *string,
	newParentID *int64,
) (*types.Repository, error) {
	repo, err := s.updateDeletedOptLock(ctx, repo, func(r *types.Repository) error {
		r.Deleted = nil
		if newIdentifier != nil {
			r.Identifier = *newIdentifier
		}
		if newParentID != nil {
			r.ParentID = *newParentID
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// Count of active repos in a space. if parentID (space) is zero then it will count all repositories in the system.
// Count deleted repos requires opts.DeletedBeforeOrAt filter.
func (s *OrmStore) Count(
	ctx context.Context,
	parentID int64,
	filter *types.RepoFilter,
) (int64, error) {
	if filter.Recursive {
		// allow parentID == 0
		return s.countAll(ctx, []int64{parentID}, filter)
	}

	parentIDs := make([]int64, 0)
	if parentID > 0 {
		parentIDs = append(parentIDs, parentID)
	}
	return s.count(ctx, parentIDs, filter)
}

func (s *OrmStore) CountMulti(
	ctx context.Context,
	parentIDs []int64,
	filter *types.RepoFilter,
) (int64, error) {
	if filter.Recursive {
		// todo: check parentIDs length great than zero
		return s.countAll(ctx, parentIDs, filter)
	}

	return s.count(ctx, parentIDs, filter)
}

func (s *OrmStore) count(
	ctx context.Context,
	parentIDs []int64,
	filter *types.RepoFilter,
) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable)

	if len(parentIDs) > 0 {
		stmt = stmt.Where("repo_parent_id IN ?", parentIDs)
	}

	stmt = applyQueryFilter(stmt, filter)

	var count int64
	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}

func (s *OrmStore) countAll(
	ctx context.Context,
	parentIDs []int64,
	filter *types.RepoFilter,
) (int64, error) {
	query := spaceDescendantsQuery + `
		SELECT space_descendant_id
		FROM space_descendants`

	var spaceIDs []int64
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).Raw(query, parentIDs).Scan(&spaceIDs).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "failed to retrieve spaces")
	}

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).Where("repo_parent_id IN ?", spaceIDs)

	stmt = applyQueryFilter(stmt, filter)

	var numRepos int64
	if err := stmt.Count(&numRepos).Error; err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "failed to count repositories")
	}

	return numRepos, nil
}

// CountByRootSpaces counts total number of repositories grouped by root spaces.
func (s *OrmStore) CountByRootSpaces(
	ctx context.Context,
) ([]types.RepositoryCount, error) {
	var result []types.RepositoryCount

	err := s.db.Raw(`
WITH RECURSIVE
	SpaceHierarchy(root_id, space_id, space_parent_id, space_uid) AS (
		SELECT space_id, space_id, space_parent_id, space_uid
		FROM spaces
		WHERE space_parent_id is null

		UNION

		SELECT h.root_id, s.space_id, s.space_parent_id, h.space_uid
		FROM spaces s
				 JOIN SpaceHierarchy h ON s.space_parent_id = h.space_id
	)
SELECT
	COUNT(r.repo_id) AS total,
	s.root_id AS root_space_id,
	s.space_uid
FROM repositories r
JOIN SpaceHierarchy s ON s.space_id = r.repo_parent_id
GROUP BY root_space_id, s.space_uid
`).Scan(&result).Error

	if err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to count repositories")
	}

	return result, nil
}

// List returns a list of active repos in a space.
// With "DeletedBeforeOrAt" filter, lists deleted repos by opts.DeletedBeforeOrAt.
func (s *OrmStore) List(
	ctx context.Context,
	parentID int64,
	filter *types.RepoFilter,
) ([]*types.Repository, error) {
	if filter.Recursive {
		return s.listAll(ctx, []int64{parentID}, filter)
	}
	return s.list(ctx, []int64{parentID}, filter)
}

// ListMulti returns a list of active repos in multi space.
// With "DeletedBeforeOrAt" filter, lists deleted repos by opts.DeletedBeforeOrAt.
func (s *OrmStore) ListMulti(
	ctx context.Context,
	parentIDs []int64,
	filter *types.RepoFilter,
) ([]*types.Repository, error) {
	if filter.Recursive {
		return s.listAll(ctx, parentIDs, filter)
	}
	return s.list(ctx, parentIDs, filter)
}

func (s *OrmStore) list(
	ctx context.Context,
	parentIDs []int64,
	filter *types.RepoFilter,
) ([]*types.Repository, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable)
	if len(parentIDs) > 0 {
		stmt = stmt.Where("repo_parent_id IN ?", parentIDs)
	}

	stmt = applyQueryFilter(stmt, filter)
	stmt = applySortFilter(stmt, filter)

	dst := []*repository{}
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return s.mapToRepos(ctx, dst)
}

func (s *OrmStore) listAll(
	ctx context.Context,
	parentIDs []int64,
	filter *types.RepoFilter,
) ([]*types.Repository, error) {
	query := spaceDescendantsQuery + `
		SELECT space_descendant_id
		FROM space_descendants`

	var spaceIDs []int64
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).Raw(query, parentIDs).Scan(&spaceIDs).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to retrieve spaces")
	}

	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).Where("repo_parent_id IN ?", spaceIDs)

	stmt = applyQueryFilter(stmt, filter)
	stmt = applySortFilter(stmt, filter)

	repos := []*repository{}
	if err := stmt.Find(&repos).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "failed to count repositories")
	}

	return s.mapToRepos(ctx, repos)
}

func (s *OrmStore) CountAllWithoutParent(ctx context.Context, filter *types.RepoFilter) (int64, error) {
	return s.count(ctx, []int64{}, filter)
}

func (s *OrmStore) ListAllWithoutParent(ctx context.Context, filter *types.RepoFilter) ([]*types.Repository, error) {
	return s.list(ctx, []int64{}, filter)
}

func (s *OrmStore) GetPublicAccess(ctx context.Context, id int64) (bool, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).Select("repo_is_public").Where("repo_id = ?", id)
	var isPublic bool
	if err := stmt.Row().Scan(&isPublic); err != nil {
		return false, database.ProcessGormSQLErrorf(ctx, err, "Failed to get public access")
	}
	return isPublic, nil
}

func (s *OrmStore) SetPublicAccess(ctx context.Context, id int64, isPublic bool) error {
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).
		Where("repo_id = ? AND repo_deleted IS NULL", id).
		Update("repo_is_public", isPublic)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update public access")
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("repo %d public access not updated: %w", id, gitfox_store.ErrResourceNotFound)
	}
	return nil
}

type repoSize struct {
	ID          int64  `gorm:"column:repo_id"`
	GitUID      string `gorm:"column:repo_git_uid"`
	Size        int64  `gorm:"column:repo_size"`
	SizeUpdated int64  `gorm:"column:repo_size_updated"`
}

func (s *OrmStore) ListSizeInfos(ctx context.Context) ([]*types.RepositorySizeInfo, error) {
	dst := []*repoSize{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(repoTable).Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return s.mapToRepoSizes(dst), nil
}

func (s *OrmStore) mapToRepo(
	ctx context.Context,
	in *repository,
) (*types.Repository, error) {
	var err error
	res := &types.Repository{
		ID:             in.ID,
		Version:        in.Version,
		ParentID:       in.ParentID,
		Identifier:     in.Identifier,
		Description:    in.Description,
		Created:        in.Created,
		CreatedBy:      in.CreatedBy,
		Updated:        in.Updated,
		Deleted:        in.Deleted.Ptr(),
		Size:           in.Size,
		SizeUpdated:    in.SizeUpdated,
		GitUID:         in.GitUID,
		DefaultBranch:  in.DefaultBranch,
		ForkID:         in.ForkID,
		PullReqSeq:     in.PullReqSeq,
		NumForks:       in.NumForks,
		NumPulls:       in.NumPulls,
		NumClosedPulls: in.NumClosedPulls,
		NumOpenPulls:   in.NumOpenPulls,
		NumMergedPulls: in.NumMergedPulls,
		Mirror:         in.Mirror,
		State:          in.State,
		IsEmpty:        in.IsEmpty,
		// Path: is set below
	}

	res.Path, err = s.getRepoPath(ctx, in.ParentID, in.Identifier)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *OrmStore) getRepoPath(ctx context.Context, parentID int64, repoIdentifier string) (string, error) {
	spacePath, err := s.spacePathStore.FindPrimaryBySpaceID(ctx, parentID)
	if err != nil {
		return "", fmt.Errorf("failed to get primary path for space %d: %w", parentID, err)
	}
	return paths.Concatenate(spacePath.Value, repoIdentifier), nil
}

func (s *OrmStore) mapToRepos(
	ctx context.Context,
	repos []*repository,
) ([]*types.Repository, error) {
	var err error
	res := make([]*types.Repository, len(repos))
	for i := range repos {
		res[i], err = s.mapToRepo(ctx, repos[i])
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (s *OrmStore) mapToRepoSize(
	in *repoSize,
) *types.RepositorySizeInfo {
	return &types.RepositorySizeInfo{
		ID:          in.ID,
		GitUID:      in.GitUID,
		Size:        in.Size,
		SizeUpdated: in.SizeUpdated,
	}
}

func (s *OrmStore) mapToRepoSizes(
	repoSizes []*repoSize,
) []*types.RepositorySizeInfo {
	res := make([]*types.RepositorySizeInfo, len(repoSizes))
	for i := range repoSizes {
		res[i] = s.mapToRepoSize(repoSizes[i])
	}
	return res
}

func mapToInternalRepo(in *types.Repository) *repository {
	return &repository{
		ID:             in.ID,
		Version:        in.Version,
		ParentID:       in.ParentID,
		Identifier:     in.Identifier,
		Description:    in.Description,
		Created:        in.Created,
		CreatedBy:      in.CreatedBy,
		Updated:        in.Updated,
		Deleted:        null.IntFromPtr(in.Deleted),
		Size:           in.Size,
		SizeUpdated:    in.SizeUpdated,
		GitUID:         in.GitUID,
		DefaultBranch:  in.DefaultBranch,
		ForkID:         in.ForkID,
		PullReqSeq:     in.PullReqSeq,
		NumForks:       in.NumForks,
		NumPulls:       in.NumPulls,
		NumClosedPulls: in.NumClosedPulls,
		NumOpenPulls:   in.NumOpenPulls,
		NumMergedPulls: in.NumMergedPulls,
		Mirror:         in.Mirror,
		State:          in.State,
		IsEmpty:        in.IsEmpty,
	}
}

func applyQueryFilter(stmt *gorm.DB, filter *types.RepoFilter) *gorm.DB {
	if filter.Query != "" {
		stmt = stmt.Where("LOWER(repo_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}
	if filter.DeletedAt != nil {
		stmt = stmt.Where("repo_deleted = ?", filter.DeletedAt)
	} else if filter.DeletedBeforeOrAt != nil {
		stmt = stmt.Where("repo_deleted <= ?", filter.DeletedBeforeOrAt)
	} else {
		stmt = stmt.Where("repo_deleted IS NULL")
	}
	return stmt
}

func applySortFilter(stmt *gorm.DB, filter *types.RepoFilter) *gorm.DB {
	stmt = stmt.Limit(cmp.Or(filter.Size, 500))
	stmt = stmt.Offset(database.GormOffset(filter.Page, filter.Size))

	switch filter.Sort {
	// TODO [CODE-1363]: remove after identifier migration.
	case enum.RepoAttrUID, enum.RepoAttrIdentifier, enum.RepoAttrNone:
		// NOTE: string concatenation is safe because the
		// order attribute is an enum and is not user-defined,
		// and is therefore not subject to injection attacks.
		stmt = stmt.Order("repo_state desc, repo_uid " + filter.Order.String())
	case enum.RepoAttrCreated:
		stmt = stmt.Order("repo_created " + filter.Order.String())
	case enum.RepoAttrUpdated:
		stmt = stmt.Order("repo_updated " + filter.Order.String())
	case enum.RepoAttrDeleted:
		stmt = stmt.Order("repo_deleted " + filter.Order.String())
	}

	return stmt
}
