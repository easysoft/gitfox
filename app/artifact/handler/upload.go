// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"net/http"

	"github.com/easysoft/gitfox/app/artifact/adapter"
)

// handUpload save artifact file, models for all of formats
func handUpload(ctx context.Context, req *http.Request, uploader adapter.ArtifactPackageUploader) error {
	_, err := uploader.Serve(ctx, req)
	if err != nil {
		return err
	}

	if err = uploader.IsValid(ctx); err != nil {
		_ = uploader.Cancel(ctx)
		return err
	}

	return uploader.Save(ctx)
}
