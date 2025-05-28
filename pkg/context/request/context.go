// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import (
	"context"
	"net/http"
)

type contextKey struct{}

func WithContext(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, contextKey{}, req)
}

func FromContext(ctx context.Context) *http.Request {
	obj, ok := ctx.Value(contextKey{}).(*http.Request)
	if !ok {
		return nil
	}
	return obj
}
