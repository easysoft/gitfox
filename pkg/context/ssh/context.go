// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package ssh

import (
	"context"

	"github.com/gliderlabs/ssh"
)

type ctxSessionKey struct{}

func WithContext(ctx context.Context, session ssh.Session) context.Context {
	return context.WithValue(ctx, ctxSessionKey{}, session)
}

func FromContext(ctx context.Context) ssh.Session {
	obj, ok := ctx.Value(ctxSessionKey{}).(ssh.Session)
	if !ok {
		return nil
	}
	return obj
}
