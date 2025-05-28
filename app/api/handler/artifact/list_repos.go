// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/space"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

func HandleListRepo(spaceCtrl *space.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)
		if !session.Principal.Admin {
			render.Forbidden(ctx, w)
			return
		}

		spaceFilter, err := request.ParseSpaceFilter(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		if spaceFilter.Order == enum.OrderDefault {
			spaceFilter.Order = enum.OrderAsc
		}

		spaces, totalCount, err := spaceCtrl.ListAllSpaces(ctx, session, spaceFilter)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		repos := make([]*types.ArtifactRepositoryRes, len(spaces))
		for id, spaceObj := range spaces {
			repos[id] = &types.ArtifactRepositoryRes{
				ID:         spaceObj.ID,
				Path:       spaceObj.Path,
				Identifier: spaceObj.Identifier,
				IsPublic:   spaceObj.IsPublic,
				CreatedBy:  spaceObj.CreatedBy,
				Created:    spaceObj.Created,
				Updated:    spaceObj.Updated,
			}
		}

		render.Pagination(r, w, spaceFilter.Page, spaceFilter.Size, int(totalCount))
		render.JSON(w, http.StatusOK, repos)
	}
}
