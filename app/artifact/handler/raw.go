// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	artctl "github.com/easysoft/gitfox/app/artifact/controller"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"
)

// HandRawUpload returns a http.HandlerFunc that Upload an artifact.
func HandRawUpload(artCtl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		baseReq, err := artCtl.LoadBaseRequest(ctx, r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		resWriter, err := artCtl.UploadRaw(ctx, r, baseReq)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		resWriter.Write(w)
	}
}

func HandRawDownload(artStore store.ArtifactStore, artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		spaceRef, err := request.GetSpaceRefFromPath(r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		reqView, err := artCtrl.ParseSpaceView(ctx, spaceRef)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		filePath, err := request.PathParamOrError(r, "*")
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		frames := strings.Split(strings.TrimPrefix(filePath, "/"), "/")
		fNum := len(frames)
		if fNum < 3 {
			render.TranslatedUserError(ctx, w, errors.New("invalid request path"))
			return
		}
		filename := frames[fNum-1]
		pkgVersion := frames[fNum-2]
		pkgName := frames[fNum-3]
		pkgGroup := ""
		if fNum > 3 {
			pkgGroup = strings.Join(frames[0:fNum-3], ".")
		}

		pkgModel, err := artStore.Packages().GetByName(ctx, pkgName, pkgGroup, reqView.OwnerID, types.ArtifactRawFormat)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		fr, meta, err := artCtrl.GetVersionAssetReader(ctx, reqView.ViewID, pkgModel.ID, pkgVersion, filename)

		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		meta.Write(w)
		render.Reader(ctx, w, http.StatusOK, fr)
	}
}
