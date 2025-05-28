// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"net/http"

	"github.com/easysoft/gitfox/app/artifact/adapter"
)

func handleUpload(ctx context.Context, req *http.Request, upload adapter.ArtifactPackageUploader) error {
	if _, err := upload.Serve(ctx, req); err != nil {
		return err
	}

	if err := upload.IsValid(ctx); err != nil {
		_ = upload.Cancel(ctx)
		return err
	}
	return upload.Save(ctx)
}
