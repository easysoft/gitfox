// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package container

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/easysoft/gitfox/app/api/render"
	gitfox_store "github.com/easysoft/gitfox/store"
)

type RegistryErr struct {
	Code            string `json:"code"`
	Message         string `json:"message"`
	Detail          string `json:"detail"`
	HTTPStatusValue int    `json:"-"`
}

func (e *RegistryErr) WithDetail(s string) *RegistryErr {
	return &RegistryErr{
		Code:            e.Code,
		Message:         e.Message,
		Detail:          s,
		HTTPStatusValue: e.HTTPStatusValue,
	}
}

func (e *RegistryErr) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Detail)
}

var (
	ErrBlobUnknown         = &RegistryErr{Code: "BLOB_UNKNOWN", Message: "blob unknown to registry", HTTPStatusValue: http.StatusNotFound}
	ErrBlobUploadInvalid   = &RegistryErr{Code: "BLOB_UPLOAD_INVALID", HTTPStatusValue: http.StatusBadRequest}
	ErrBlobUploadUnknown   = &RegistryErr{Code: "BLOB_UPLOAD_UNKNOWN", HTTPStatusValue: http.StatusNotFound}
	ErrDigestInvalid       = &RegistryErr{Code: "DIGEST_INVALID", HTTPStatusValue: http.StatusBadRequest}
	ErrManifestBlobUnknown = &RegistryErr{Code: "MANIFEST_BLOB_UNKNOWN", HTTPStatusValue: http.StatusNotFound}
	ErrManifestInvalid     = &RegistryErr{Code: "MANIFEST_INVALID", HTTPStatusValue: http.StatusBadRequest}
	ErrManifestUnknown     = &RegistryErr{Code: "MANIFEST_UNKNOWN", HTTPStatusValue: http.StatusNotFound}
	ErrNameInvalid         = &RegistryErr{Code: "NAME_INVALID", HTTPStatusValue: http.StatusBadRequest}
	ErrNameUnknown         = &RegistryErr{Code: "NAME_UNKNOWN", HTTPStatusValue: http.StatusNotFound}
	ErrSizeInvalid         = &RegistryErr{Code: "SIZE_INVALID", HTTPStatusValue: http.StatusBadRequest}
	ErrUnauthorized        = &RegistryErr{Code: "UNAUTHORIZED", HTTPStatusValue: http.StatusUnauthorized}
	ErrDenied              = &RegistryErr{Code: "DENIED", Message: "requested access to the resource is denied", HTTPStatusValue: http.StatusForbidden}
	ErrUnsupported         = &RegistryErr{Code: "UNSUPPORTED", HTTPStatusValue: http.StatusNotImplemented}
	ErrCodeRangeInvalid    = &RegistryErr{Code: "RANGE_INVALID", Message: "invalid content range", HTTPStatusValue: http.StatusRequestedRangeNotSatisfiable}
	ErrClientClosed        = &RegistryErr{Code: "CLIENT_CLOSED", HTTPStatusValue: 499}
	ErrUnknown             = &RegistryErr{Code: "UNKNOWN", HTTPStatusValue: http.StatusInternalServerError}
)

type registryErrs struct {
	Errors []*RegistryErr `json:"errors"`
}

func RenderError(ctx context.Context, w http.ResponseWriter, err error) {
	e := translateError(err)
	render.JSON(w, e.HTTPStatusValue, &registryErrs{
		Errors: []*RegistryErr{e},
	})
}

func translateError(err error) *RegistryErr {
	var expect *RegistryErr
	if errors.As(err, &expect) {
		return expect
	}
	switch {
	case errors.Is(err, gitfox_store.ErrResourceNotFound):
		return ErrBlobUnknown.WithDetail("resource not found")
	default:
		return ErrUnknown.WithDetail(err.Error())
	}
}
