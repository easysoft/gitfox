// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package codeowners

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/git/hook"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/rs/zerolog/log"
)

func (s *Service) getApplicableCodeOwnersForPush(
	ctx context.Context,
	repo *types.Repository,
	ref hook.ReferenceUpdate,
	objDir []string,
) (*CodeOwners, error) {
	codeOwners, err := s.get(ctx, repo, ref.Ref)
	if err != nil {
		return nil, err
	}

	diffFileStats, err := s.git.ReceiveDiffFileNames(ctx, &git.DiffParams{
		ReadParams: git.ReadParams{
			RepoUID:             repo.GetGitUID(),
			AlternateObjectDirs: objDir,
		},
		BaseRef: ref.New.String(),
		HeadRef: ref.Old.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get diff file stat: %w", err)
	}

	entryIDs := map[int]struct{}{}
	for _, file := range diffFileStats.Files {
		// last rule that matches wins (hence simply go in reverse order)
		for i := len(codeOwners.Entries) - 1; i >= 0; i-- {
			pattern := codeOwners.Entries[i].Pattern
			if ok, err := match(pattern, file); err != nil {
				return nil, fmt.Errorf("failed to match pattern %q for file %q: %w", pattern, file, err)
			} else if ok {
				entryIDs[i] = struct{}{}
				break
			}
		}
	}

	filteredEntries := make([]Entry, 0, len(entryIDs))
	for i := range entryIDs {
		if !codeOwners.Entries[i].IsOwnershipReset() {
			filteredEntries = append(filteredEntries, codeOwners.Entries[i])
		}
	}

	// sort output to match order of occurrence in source CODEOWNERS file
	sort.Slice(
		filteredEntries,
		func(i, j int) bool { return filteredEntries[i].LineNumber <= filteredEntries[j].LineNumber },
	)

	return &CodeOwners{
		FileSHA: codeOwners.FileSHA,
		Entries: filteredEntries,
	}, err
}

func (s *Service) ValidateCodeOwnerByPush(
	ctx context.Context,
	repo *types.Repository,
	uid int64,
	ref hook.ReferenceUpdate,
	objDir []string,
) ([]Entry, error) {
	denyEntries := make([]Entry, 0)

	owners, err := s.getApplicableCodeOwnersForPush(ctx, repo, ref, objDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get codeOwners: %w", err)
	}

	if owners == nil || len(owners.Entries) == 0 {
		return denyEntries, nil
	}

	for _, entry := range owners.Entries {
		ownerMatched := false
		for _, owner := range entry.Owners {
			// check for usrgrp
			if strings.HasPrefix(owner, userGroupPrefixMarker) {
				// user group not supported, skip
				continue
			}
			// user email based codeowner
			dbUser, err := s.principalStore.FindByEmail(ctx, owner)
			if errors.Is(err, gitfox_store.ErrResourceNotFound) {
				log.Ctx(ctx).Debug().Msgf("user %q not found in database hence skipping for code owner", owner)
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("error resolving user by email : %w", err)
			}
			if uid == dbUser.ID {
				ownerMatched = true
				break
			}
		}
		if !ownerMatched {
			denyEntries = append(denyEntries, entry)
		}
	}

	return denyEntries, nil
}
