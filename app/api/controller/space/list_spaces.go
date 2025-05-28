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

package space

import (
	"context"
	"fmt"

	apiauth "github.com/easysoft/gitfox/app/api/auth"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/store/database/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

// ListSpaces lists the child spaces of a space.
func (c *Controller) ListSpaces(ctx context.Context,
	session *auth.Session,
	spaceRef string,
	filter *types.SpaceFilter,
) ([]*SpaceOutput, int64, error) {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return nil, 0, err
	}

	if err = apiauth.CheckSpaceScope(
		ctx,
		c.authorizer,
		session,
		space,
		enum.ResourceTypeSpace,
		enum.PermissionSpaceView,
	); err != nil {
		return nil, 0, err
	}

	return c.ListSpacesNoAuth(ctx, space.ID, filter)
}

// ListSpacesNoAuth lists spaces WITHOUT checking PermissionSpaceView.
func (c *Controller) ListSpacesNoAuth(
	ctx context.Context,
	spaceID int64,
	filter *types.SpaceFilter,
) ([]*SpaceOutput, int64, error) {
	var spaces []*types.Space
	var count int64

	err := c.tx.WithTx(ctx, func(ctx context.Context) (err error) {
		count, err = c.spaceStore.Count(ctx, spaceID, filter)
		if err != nil {
			return fmt.Errorf("failed to count child spaces: %w", err)
		}

		spaces, err = c.spaceStore.List(ctx, spaceID, filter)
		if err != nil {
			return fmt.Errorf("failed to list child spaces: %w", err)
		}

		return nil
	}, dbtx.TxDefaultReadOnly)
	if err != nil {
		return nil, 0, err
	}

	// backfill public access mode
	var spacesOut []*SpaceOutput
	for _, space := range spaces {
		spaceOut, err := GetSpaceOutput(ctx, c.publicAccess, space)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get space %q output: %w", space.Path, err)
		}

		spacesOut = append(spacesOut, spaceOut)
	}

	return spacesOut, count, nil
}

func (c *Controller) ListAllSpaces(
	ctx context.Context,
	session *auth.Session,
	filter *types.SpaceFilter,
) ([]*SpaceOutput, int64, error) {
	if session.User.Admin {
		return c.ListSpacesNoAuth(ctx, 0, filter)
	} else {
		return c.ListSpacesByMembership(ctx, session.User.ID, filter)
	}
}

func (c *Controller) ListSpacesByMembership(ctx context.Context, userID int64, filter *types.SpaceFilter,
) ([]*SpaceOutput, int64, error) {
	var membershipSpaces []types.MembershipSpace
	var membershipsCount int64
	var err error

	memberFilter := types.MembershipSpaceFilter{
		ListQueryFilter: types.ListQueryFilter{
			Pagination: types.Pagination{Page: filter.Page, Size: filter.Size},
			Query:      filter.Query,
		},
		Order: filter.Order,
	}

	switch filter.Sort {
	case enum.SpaceAttrUID, enum.SpaceAttrIdentifier, enum.SpaceAttrNone:
		memberFilter.Sort = enum.MembershipSpaceSortIdentifier
	case enum.SpaceAttrCreated:
		memberFilter.Sort = enum.MembershipSpaceSortCreated
	default:
		memberFilter.Sort = enum.MembershipSpaceSortIdentifier
	}

	err = c.tx.WithTx(ctx, func(ctx context.Context) error {
		membershipSpaces, err = c.membershipStore.ListSpaces(ctx, userID, memberFilter)
		if err != nil {
			return fmt.Errorf("failed to list membership spaces for user: %w", err)
		}

		if filter.Page == 1 && len(membershipSpaces) < filter.Size {
			membershipsCount = int64(len(membershipSpaces))
			return nil
		}

		membershipsCount, err = c.membershipStore.CountSpaces(ctx, userID, memberFilter)
		if err != nil {
			return fmt.Errorf("failed to count memberships for user: %w", err)
		}

		return nil
	}, dbtx.TxDefaultReadOnly)
	if err != nil {
		return nil, 0, err
	}

	// backfill public access mode
	spacesOut := make([]*SpaceOutput, len(membershipSpaces))
	for i, m := range membershipSpaces {
		s := m.Space
		spaceOut, err := GetSpaceOutput(ctx, c.publicAccess, &s)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get space %q output: %w", s.Path, err)
		}
		spacesOut[i] = spaceOut
	}
	return spacesOut, membershipsCount, nil
}
