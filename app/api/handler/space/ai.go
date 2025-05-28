// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package space

import (
	"encoding/json"
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/space"
	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
)

func HandleListAIConfigs(spaceCtrl *space.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)
		spaceRef, err := request.GetSpaceRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		ret, _, err := spaceCtrl.ListAIConfigs(ctx, session, spaceRef)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, ret)
	}
}

func HandleCreateAIConfig(spaceCtrl *space.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)

		spaceRef, err := request.GetSpaceRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		in := new(space.AiConfigCreateInput)
		err = json.NewDecoder(r.Body).Decode(in)
		if err != nil {
			render.BadRequestf(ctx, w, "Invalid request body: %s.", err)
			return
		}

		cfg, err := spaceCtrl.CreateAIConfig(ctx, session, spaceRef, in)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		render.JSON(w, http.StatusCreated, cfg)
	}
}

func HandleTestAIConfig(spaceCtrl *space.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)

		spaceRef, err := request.GetSpaceRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		id, err := request.GetAIConfigIDFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		if err := spaceCtrl.TestAIConfig(ctx, session, spaceRef, id); err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, map[string]string{"message": "ok"})
	}
}

func HandleSetDefaultAIConfig(spaceCtrl *space.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)
		spaceRef, err := request.GetSpaceRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		id, err := request.GetAIConfigIDFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		if err := spaceCtrl.SetDefaultAIConfig(ctx, session, spaceRef, id); err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		render.JSON(w, http.StatusOK, map[string]string{"message": "ok"})
	}
}

func HandleUpdateAIConfig(spaceCtrl *space.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)

		spaceRef, err := request.GetSpaceRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		id, err := request.GetAIConfigIDFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		in := new(space.AiConfigUpdateInput)
		err = json.NewDecoder(r.Body).Decode(in)
		if err != nil {
			render.BadRequestf(ctx, w, "Invalid request body: %s.", err)
			return
		}

		cfg, err := spaceCtrl.UpdateAIConfig(ctx, session, spaceRef, id, in)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}
		render.JSON(w, http.StatusOK, cfg)
	}
}

func HandleDeleteAIConfig(spaceCtrl *space.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, _ := request.AuthSessionFrom(ctx)

		spaceRef, err := request.GetSpaceRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		id, err := request.GetAIConfigIDFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		if err := spaceCtrl.DeleteAIConfig(ctx, session, spaceRef, id); err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, map[string]string{"message": "ok"})
	}
}
