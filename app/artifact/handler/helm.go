// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"

	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/artifact/adapter/helm"
	artctl "github.com/easysoft/gitfox/app/artifact/controller"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/app/url"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/Masterminds/semver"
)

const (
	_helmPathPattern = `^(?P<name>[a-z][a-z0-9\-]+[a-z0-9])-(?P<version>` + semver.SemVerRegex + `)\.tgz$`
)

// HandHelmUpload returns a http.HandlerFunc that Upload an artifact.
func HandHelmUpload(artCtl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		baseReq, err := artCtl.LoadBaseRequest(ctx, r)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		resWriter, err := artCtl.UploadHelm(ctx, r, baseReq)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		resWriter.Write(w)
	}
}

// HandHelmDownload returns a http.HandlerFunc that download a file.
func HandHelmDownload(artStore store.ArtifactStore, artCtrl *artctl.Controller) http.HandlerFunc {
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

		if filePath == "index.yaml" {
			helmIndexResponse(ctx, w, spaceRef, filePath, reqView.ViewID, artCtrl)
			return
		}

		name, version, err := parseHelmPath(filePath)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		pkgModel, err := artStore.Packages().GetByName(ctx, name, "", reqView.OwnerID, types.ArtifactHelmFormat)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		fr, meta, err := artCtrl.GetVersionAssetReader(ctx, reqView.ViewID, pkgModel.ID, version, filePath)

		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		meta.Write(w)
		render.Reader(ctx, w, http.StatusOK, fr)
	}
}

func helmIndexResponse(ctx context.Context, w http.ResponseWriter, spaceRef, path string, viewId int64, artCtrl *artctl.Controller) {
	rd, meta, err := artCtrl.GetMetaReader(ctx, path, types.ArtifactHelmFormat, viewId)
	if err != nil {
		if errors.Is(err, gitfox_store.ErrResourceNotFound) {
			emptyContent := helm.EmptyIndex(genDynamicContextPath(spaceRef))
			rd = io.NopCloser(bytes.NewReader(emptyContent))
			w.Header().Set("Content-Type", "application/x-yaml")
			render.Reader(ctx, w, http.StatusOK, rd)
			return
		} else {
			render.TranslatedUserError(ctx, w, err)
			return
		}
	}

	meta.Write(w)
	render.Reader(ctx, w, http.StatusOK, rd)
	return
}

func parseHelmPath(p string) (name, version string, err error) {
	fName := path.Base(p)
	reg := regexp.MustCompile(_helmPathPattern)
	match := reg.FindStringSubmatch(fName)

	if len(match) < 3 {
		return "", "", fmt.Errorf("invalid helm filename")
	}

	return match[1], match[2], nil
}

func genDynamicContextPath(spaceRef string) string {
	return path.Join("/", url.ArtifactMount, spaceRef, "helm")
}
