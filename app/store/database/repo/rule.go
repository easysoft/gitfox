// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo

import (
	"context"
	"encoding/json"
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
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var _ store.RuleStore = (*RuleStore)(nil)

// NewRuleOrmStore returns a new RuleStore.
func NewRuleOrmStore(
	db *gorm.DB,
	pCache store.PrincipalInfoCache,
) *RuleStore {
	return &RuleStore{
		pCache: pCache,
		db:     db,
	}
}

// RuleStore implements a store.RuleStore backed by a relational database.
type RuleStore struct {
	db     *gorm.DB
	pCache store.PrincipalInfoCache
}

type rule struct {
	ID      int64 `gorm:"column:rule_id;primaryKey"`
	Version int64 `gorm:"column:rule_version"`

	CreatedBy int64 `gorm:"column:rule_created_by"`
	Created   int64 `gorm:"column:rule_created"`
	Updated   int64 `gorm:"column:rule_updated"`

	SpaceID null.Int `gorm:"column:rule_space_id"`
	RepoID  null.Int `gorm:"column:rule_repo_id"`

	Identifier  string `gorm:"column:rule_uid"`
	Description string `gorm:"column:rule_description"`

	Type  types.RuleType `gorm:"column:rule_type"`
	State enum.RuleState `gorm:"column:rule_state"`

	Pattern    string `gorm:"column:rule_pattern"`
	Definition string `gorm:"column:rule_definition"`

	Buildin bool `gorm:"column:rule_buildin"`
}

const (
	ruleTable = "rules"
)

// FindBuildIn finds the buildin rule by identifier.
func (s *RuleStore) FindBuildIn(ctx context.Context, identifier string, repoID int64) (*types.Rule, error) {
	dst := &rule{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table("").First(dst, "rule_uid = ? and rule_buildin = true and rule_repo_id = ?", identifier, repoID).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find buildin rule")
	}

	r := s.mapToRule(ctx, dst)

	return &r, nil
}

// Find finds the rule by id.
func (s *RuleStore) Find(ctx context.Context, id int64) (*types.Rule, error) {
	dst := &rule{}
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table("").First(dst, id).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find rule")
	}

	r := s.mapToRule(ctx, dst)

	return &r, nil
}

func (s *RuleStore) FindByIdentifier(
	ctx context.Context,
	spaceID *int64,
	repoID *int64,
	identifier string,
) (*types.Rule, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(ruleTable).Where("LOWER(rule_uid) = ?", strings.ToLower(identifier))
	stmt = s.applyParentID(stmt, spaceID, repoID)

	dst := &rule{}
	if err := stmt.Take(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing find rule by identifier query")
	}

	r := s.mapToRule(ctx, dst)

	return &r, nil
}

// Create creates a new protection rule.
func (s *RuleStore) Create(ctx context.Context, rule *types.Rule) error {
	dbRule := mapToInternalRule(rule)

	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(ruleTable).Create(&dbRule).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert rule query failed")
	}

	r := s.mapToRule(ctx, &dbRule)

	*rule = r

	return nil
}

// Update updates the protection rule details.
func (s *RuleStore) Update(ctx context.Context, ruleObj *types.Rule) error {
	dbRule := mapToInternalRule(ruleObj)
	dbRule.Version++
	dbRule.Updated = time.Now().UnixMilli()

	updateFields := []string{"Version", "Updated", "Identifier", "Description", "State", "Pattern", "Definition"}
	res := dbtx.GetOrmAccessor(ctx, s.db).Table(ruleTable).
		Where(&rule{ID: ruleObj.ID, Version: dbRule.Version - 1}).
		Select(updateFields).Updates(dbRule)

	if res.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, res.Error, "Failed to update rule")
	}

	count := res.RowsAffected

	if count == 0 {
		return gitfox_store.ErrVersionConflict
	}

	ruleObj.Version = dbRule.Version
	ruleObj.Updated = dbRule.Updated

	return nil
}

// Delete the protection rule.
func (s *RuleStore) Delete(ctx context.Context, id int64) error {
	if err := dbtx.GetOrmAccessor(ctx, s.db).Table(ruleTable).Delete(&rule{ID: id}).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "the delete rule query failed")
	}

	return nil
}

func (s *RuleStore) DeleteByIdentifier(ctx context.Context, spaceID, repoID *int64, identifier string) error {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(ruleTable).Where("LOWER(rule_uid) = ?", strings.ToLower(identifier))

	if spaceID != nil {
		stmt = stmt.Where("rule_space_id = ?", *spaceID)
	}

	if repoID != nil {
		stmt = stmt.Where("rule_repo_id = ?", *repoID)
	}

	if err := stmt.Delete(&rule{}).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Failed executing delete rule by identifier query")
	}

	return nil
}

// Count returns count of protection rules matching the provided criteria.
func (s *RuleStore) Count(ctx context.Context, spaceID, repoID *int64, filter *types.RuleFilter) (int64, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(ruleTable)

	stmt = s.applyParentID(stmt, spaceID, repoID)
	stmt = s.applyFilter(stmt, filter)

	var count int64

	err := stmt.Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count rules query")
	}

	return count, nil
}

// List returns a list of protection rules of a repository or a space.
func (s *RuleStore) List(ctx context.Context, spaceID, repoID *int64, filter *types.RuleFilter) ([]types.Rule, error) {
	stmt := dbtx.GetOrmAccessor(ctx, s.db).Table(ruleTable)

	stmt = s.applyParentID(stmt, spaceID, repoID)
	stmt = s.applyFilter(stmt, filter)

	stmt = stmt.Limit(int(database.Limit(filter.Size)))
	stmt = stmt.Offset(int(database.Offset(filter.Page, filter.Size)))

	order := filter.Order
	if order == enum.OrderDefault {
		order = enum.OrderAsc
	}

	switch filter.Sort {
	case enum.RuleSortCreated:
		stmt = stmt.Order("rule_created " + order.String())
	case enum.RuleSortUpdated:
		stmt = stmt.Order("rule_updated " + order.String())
		// TODO [CODE-1363]: remove after identifier migration.
	case enum.RuleSortUID, enum.RuleSortIdentifier:
		stmt = stmt.Order("LOWER(rule_uid) " + order.String())
	}

	dst := make([]rule, 0)
	if err := stmt.Scan(&dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return s.mapToRules(ctx, dst), nil
}

type ruleInfo struct {
	SpacePath  string         `gorm:"column:space_path"`
	RepoPath   string         `gorm:"column:repo_path"`
	ID         int64          `gorm:"column:rule_id"`
	Identifier string         `gorm:"column:rule_uid"`
	Type       types.RuleType `gorm:"column:rule_type"`
	State      enum.RuleState `gorm:"column:rule_state"`
	Pattern    string         `gorm:"column:rule_pattern"`
	Definition string         `gorm:"column:rule_definition"`
	Buildin    bool           `gorm:"column:rule_buildin"`
}

// ListAllRepoRules returns a list of all protection rules that can be applied on a repository.
// This includes the rules defined directly on the repository and all those defined on the parent spaces.
func (s *RuleStore) ListAllRepoRules(ctx context.Context, repoID int64) ([]types.RuleInfoInternal, error) {
	const query = `
		WITH RECURSIVE
			repo_info(repo_id, repo_uid, repo_space_id) AS (
				SELECT repo_id, repo_uid, repo_parent_id
				FROM repositories
				WHERE repo_id = ?
			),
			space_parents(space_id, space_uid, space_parent_id) AS (
				SELECT space_id, space_uid, space_parent_id
				FROM spaces
				INNER JOIN repo_info ON repo_info.repo_space_id = spaces.space_id
				UNION ALL
				SELECT spaces.space_id, spaces.space_uid, spaces.space_parent_id
				FROM spaces
				INNER JOIN space_parents ON space_parents.space_parent_id = spaces.space_id
			),
			spaces_with_path(space_id, space_parent_id, space_uid, space_full_path) AS (
				SELECT space_id, space_parent_id, space_uid, space_uid
				FROM space_parents
				WHERE space_parent_id IS NULL
				UNION ALL
				SELECT
					space_parents.space_id,
					space_parents.space_parent_id,
					space_parents.space_uid,
					spaces_with_path.space_full_path || '/' || space_parents.space_uid
				FROM space_parents
				INNER JOIN spaces_with_path ON spaces_with_path.space_id = space_parents.space_parent_id
			)
		SELECT
			 space_full_path AS "space_path"
			,'' as "repo_path"
			,rule_id
			,rule_uid
			,rule_type
			,rule_state
			,rule_pattern
			,rule_definition
		FROM spaces_with_path
		INNER JOIN rules ON rules.rule_space_id = spaces_with_path.space_id
		WHERE rule_state IN ('active', 'monitor')
		UNION ALL
		SELECT
			 '' as "space_path"
			,space_full_path || '/' || repo_info.repo_uid AS "repo_path"
			,rule_id
			,rule_uid
			,rule_type
			,rule_state
			,rule_pattern
			,rule_definition
		FROM rules
		INNER JOIN repo_info ON repo_info.repo_id = rules.rule_repo_id
		INNER JOIN spaces_with_path ON spaces_with_path.space_id = repo_info.repo_space_id
		WHERE rule_state IN ('active', 'monitor')`

	result := make([]ruleInfo, 0)
	if err := dbtx.GetOrmAccessor(ctx, s.db).Raw(query, repoID).Scan(&result).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed executing custom list query")
	}

	return s.mapToRuleInfos(result), nil
}

func (*RuleStore) applyParentID(
	stmt *gorm.DB,
	spaceID, repoID *int64,
) *gorm.DB {
	if spaceID != nil {
		stmt = stmt.Where("rule_space_id = ?", *spaceID)
	}

	if repoID != nil {
		stmt = stmt.Where("rule_repo_id = ?", *repoID)
	}

	return stmt
}

func (*RuleStore) applyFilter(
	stmt *gorm.DB,
	filter *types.RuleFilter,
) *gorm.DB {
	if len(filter.States) == 1 {
		stmt = stmt.Where("rule_state = ?", filter.States[0])
	} else if len(filter.States) > 1 {
		stmt = stmt.Where("rule_state in ?", filter.States)
	}

	if filter.Query != "" {
		stmt = stmt.Where("LOWER(rule_uid) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query)))
	}

	return stmt
}

func (s *RuleStore) mapToRule(
	ctx context.Context,
	in *rule,
) types.Rule {
	r := types.Rule{
		ID:          in.ID,
		Version:     in.Version,
		CreatedBy:   in.CreatedBy,
		Created:     in.Created,
		Updated:     in.Updated,
		SpaceID:     in.SpaceID.Ptr(),
		RepoID:      in.RepoID.Ptr(),
		Identifier:  in.Identifier,
		Description: in.Description,
		Type:        in.Type,
		State:       in.State,
		Pattern:     json.RawMessage(in.Pattern),
		Definition:  json.RawMessage(in.Definition),
		Buildin:     in.Buildin,
	}

	createdBy, err := s.pCache.Get(ctx, in.CreatedBy)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to load rule creator")
	}

	if createdBy != nil {
		r.CreatedByInfo = *createdBy
	}

	return r
}

func (s *RuleStore) mapToRules(
	ctx context.Context,
	rules []rule,
) []types.Rule {
	res := make([]types.Rule, len(rules))
	for i := 0; i < len(rules); i++ {
		res[i] = s.mapToRule(ctx, &rules[i])
	}
	return res
}

func mapToInternalRule(in *types.Rule) rule {
	return rule{
		ID:          in.ID,
		Version:     in.Version,
		CreatedBy:   in.CreatedBy,
		Created:     in.Created,
		Updated:     in.Updated,
		SpaceID:     null.IntFromPtr(in.SpaceID),
		RepoID:      null.IntFromPtr(in.RepoID),
		Identifier:  in.Identifier,
		Description: in.Description,
		Type:        in.Type,
		State:       in.State,
		Pattern:     string(in.Pattern),
		Definition:  string(in.Definition),
		Buildin:     in.Buildin,
	}
}

func (*RuleStore) mapToRuleInfo(in *ruleInfo) types.RuleInfoInternal {
	return types.RuleInfoInternal{
		RuleInfo: types.RuleInfo{
			SpacePath:  in.SpacePath,
			RepoPath:   in.RepoPath,
			ID:         in.ID,
			Identifier: in.Identifier,
			Type:       in.Type,
			State:      in.State,
			Buildin:    in.Buildin,
		},
		Pattern:    json.RawMessage(in.Pattern),
		Definition: json.RawMessage(in.Definition),
	}
}

func (s *RuleStore) mapToRuleInfos(
	ruleInfos []ruleInfo,
) []types.RuleInfoInternal {
	res := make([]types.RuleInfoInternal, len(ruleInfos))
	for i := 0; i < len(ruleInfos); i++ {
		res[i] = s.mapToRuleInfo(&ruleInfos[i])
	}
	return res
}
