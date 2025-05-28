// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"net/http"

	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/artifact/adapter/raw"
)

func (c *Controller) UploadRaw(ctx context.Context, r *http.Request, helmReq *BaseReq) (HttpResponseWriter, error) {
	if c.checkAuthArtifactPush(ctx, helmReq) != nil {
		return nil, usererror.ErrForbidden
	}

	u := raw.NewUploader(helmReq.view.Store, c.artStore, helmReq.view)
	if e := c.tx.WithTx(ctx, func(ctx context.Context) error {
		return handleUpload(ctx, r, u)
	}); e != nil {
		return nil, e
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusCreated)
	}), nil
}
