// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package handler

import (
	"errors"
	"net/http"

	"github.com/easysoft/gitfox/app/api/render"
	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	artctl "github.com/easysoft/gitfox/app/artifact/controller"
	gitfox_store "github.com/easysoft/gitfox/store"

	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
)

func HandleAPIBase() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data = make(map[string]interface{})
		render.JSON(w, http.StatusOK, data)
		w.WriteHeader(http.StatusAccepted)
	}
}

func HandleBlobHead(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		artReq, err := artCtrl.LoadContainerRequest(ctx, r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		digestFromPath, err := request.GetDigestFromPath(r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		if _, err = digest.Parse(digestFromPath); err != nil {
			container.RenderError(ctx, w, container.ErrDigestInvalid)
			return
		}

		resWriter, err := artCtrl.HeadBlob(ctx, artReq, digestFromPath)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		resWriter.Write(w)
		w.WriteHeader(http.StatusOK)
	}
}

func HandleBlobGet(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		artReq, err := artCtrl.LoadContainerRequest(ctx, r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		digestFromPath, err := request.GetDigestFromPath(r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		if _, err = digest.Parse(digestFromPath); err != nil {
			container.RenderError(ctx, w, container.ErrDigestInvalid)
			return
		}

		resWriter, err := artCtrl.GetContainerBlob(ctx, artReq, digestFromPath)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		resWriter.Write(w)
	}
}

func HandleBlobDelete(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		artReq, err := artCtrl.LoadContainerRequest(ctx, r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		digestFromPath, err := request.GetDigestFromPath(r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		resWriter, err := artCtrl.DeleteContainerBlob(ctx, artReq, digestFromPath)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		resWriter.Write(w)
	}
}

func HandleBlobUploadStart(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// cross mount feature is not implemented
		if r.FormValue("from") != "" || r.FormValue("mount") != "" {
			container.RenderError(ctx, w, container.ErrUnsupported.WithDetail("cross mount is not supported"))
			return
		}

		ctnReq, err := artCtrl.LoadContainerRequest(ctx, r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		var resWriter artctl.HttpResponseWriter
		paramDigest := r.FormValue(request.ParamDigest)
		if paramDigest != "" {
			// found digest, do single upload
			dgst, e := digest.Parse(paramDigest)
			if e != nil {
				container.RenderError(ctx, w, container.ErrDigestInvalid)
				return
			}
			resWriter, err = artCtrl.SaveContainerBlob(ctx, ctnReq, r, dgst)
		} else {
			// init blob upload
			resWriter, err = artCtrl.StartBlobUpload(ctx, ctnReq)
		}

		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		resWriter.Write(w)
		w.WriteHeader(http.StatusAccepted)
	}
}

func HandleBlobUploadCancel(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		uploadReq, err := artCtrl.LoadContainerResumeUploadRequest(ctx, r, true)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		resWriter, err := artCtrl.CancelBlobUpload(ctx, uploadReq)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		resWriter.Write(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleBlobUploadStatus(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		uploadReq, err := artCtrl.LoadContainerResumeUploadRequest(ctx, r, false)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		resWriter, err := artCtrl.GetBlobUploadStatus(ctx, uploadReq)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		resWriter.Write(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleBlobUploadPatch(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		uploadReq, err := artCtrl.LoadContainerResumeUploadRequest(ctx, r, true)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		resWriter, err := artCtrl.AppendBlobContent(ctx, r, uploadReq)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("append blob content failed")
			container.RenderError(ctx, w, err)
			return
		}
		resWriter.Write(w)
		w.WriteHeader(http.StatusAccepted)
	}
}

func HandleBlobUploadFinish(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		uploadReq, err := artCtrl.LoadContainerResumeUploadRequest(ctx, r, true)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		resWriter, err := artCtrl.FinishBlobUpload(ctx, r, uploadReq)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		resWriter.Write(w)
		w.WriteHeader(http.StatusCreated)
	}
}

func HandleManifestHead(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req, err := artCtrl.LoadContainerManifestRequest(ctx, r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		var resWriter artctl.HttpResponseWriter
		if req.IsTag {
			resWriter, err = artCtrl.HeadContainerTag(ctx, r, req)
		} else {
			resWriter, err = artCtrl.HeadContainerManifest(ctx, r, req)
		}

		if err != nil {
			if errors.Is(err, gitfox_store.ErrResourceNotFound) {
				container.RenderError(ctx, w, container.ErrManifestUnknown)
			} else {
				container.RenderError(ctx, w, err)
			}
			return
		}
		resWriter.Write(w)
	}
}

func HandleManifestGet(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req, err := artCtrl.LoadContainerManifestRequest(ctx, r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		var resWriter artctl.HttpResponseWriter
		if req.IsTag {
			resWriter, err = artCtrl.GetContainerTag(ctx, r, req)
		} else {
			resWriter, err = artCtrl.GetContainerManifest(ctx, r, req)
		}

		if err != nil {
			if errors.Is(err, gitfox_store.ErrResourceNotFound) {
				container.RenderError(ctx, w, container.ErrManifestUnknown)
			} else {
				container.RenderError(ctx, w, err)
			}
			return
		}
		resWriter.Write(w)
	}
}

func HandleManifestPut(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req, err := artCtrl.LoadContainerManifestRequest(ctx, r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		var resWriter artctl.HttpResponseWriter
		if req.IsTag {
			resWriter, err = artCtrl.PutContainerTag(ctx, r, req)
		} else {
			resWriter, err = artCtrl.PutManifest(ctx, r, req)
		}

		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		resWriter.Write(w)
	}
}

func HandleManifestDelete(artCtrl *artctl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req, err := artCtrl.LoadContainerManifestRequest(ctx, r)
		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}

		var resWriter artctl.HttpResponseWriter
		if req.IsTag {
			resWriter, err = artCtrl.DeleteContainerTag(ctx, req)
		} else {
			resWriter, err = artCtrl.DeleteContainerManifest(ctx, req)
		}

		if err != nil {
			container.RenderError(ctx, w, err)
			return
		}
		resWriter.Write(w)
	}
}
