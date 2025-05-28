// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/types"

	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
)

type ArtifactAuthRequest interface {
	Session() *auth.Session
	Space() *types.Space
}

type ContainerReq struct {
	spaceName string
	repoName  string

	session *auth.Session
	view    *adapter.ViewDescriptor
}

func (cr *ContainerReq) FullName() string {
	return cr.spaceName + "/" + cr.repoName
}

func (cr *ContainerReq) Session() *auth.Session {
	return cr.session
}

func (cr *ContainerReq) Space() *types.Space {
	return cr.view.Space
}

type ContainerManifestRequest struct {
	*ContainerReq

	IsTag  bool
	Tag    string
	Digest digest.Digest
}

type ContainerUploadRequest struct {
	*ContainerReq
	resume *container.ResumeRequest
}

func (c *Controller) LoadContainerRequest(ctx context.Context, r *http.Request) (*ContainerReq, error) {
	spaceName, repoName, err := request.GetContainerRepositoryFromPath(r)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("could not get container repository name")
		return nil, container.ErrNameInvalid.WithDetail(err.Error())
	}

	reqView, err := c.ParseSpaceView(ctx, spaceName)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("error parsing view")
		return nil, container.ErrUnknown.WithDetail(err.Error())
	}

	session, _ := request.AuthSessionFrom(ctx)
	return &ContainerReq{
		spaceName: spaceName,
		repoName:  repoName,
		session:   session,
		view:      reqView,
	}, nil
}

func (c *Controller) LoadContainerManifestRequest(ctx context.Context, r *http.Request) (*ContainerManifestRequest, error) {
	baseReq, err := c.LoadContainerRequest(ctx, r)
	if err != nil {
		return nil, err
	}

	reference, err := request.GetReferenceFromPath(r)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("could not get reference")
		return nil, container.ErrManifestUnknown.WithDetail(err.Error())
	}

	req := &ContainerManifestRequest{
		ContainerReq: baseReq,
	}

	// validate tag
	if validateTag(reference) {
		req.Tag = reference
		req.IsTag = true
		return req, nil
	}

	dg, err := digest.Parse(reference)
	if err != nil {
		return nil, container.ErrManifestUnknown.WithDetail(err.Error())
	}

	req.Digest = dg
	return req, nil
}

func (c *Controller) LoadContainerResumeUploadRequest(ctx context.Context, r *http.Request, requireState bool) (*ContainerUploadRequest, error) {
	baseReq, err := c.LoadContainerRequest(ctx, r)
	if err != nil {
		return nil, err
	}

	paramUUID, err := request.GetUUIDFromPath(r)
	if err != nil {
		return nil, err
	}

	token := r.FormValue("_state")
	if token == "" && requireState {
		return nil, errors.New("missing _state")
	}

	req, err := container.ParseResumeRequest(ctx, paramUUID, baseReq.FullName(), token, c.artStore, baseReq.view)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("parse resume upload request failed")
		return nil, err
	}
	return &ContainerUploadRequest{
		ContainerReq: baseReq, resume: req,
	}, nil
}
