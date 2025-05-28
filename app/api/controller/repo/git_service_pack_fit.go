// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package repo

import (
	"context"
	"net/http"
	"strings"

	ctxrequest "github.com/easysoft/gitfox/pkg/context/request"
	ctxssh "github.com/easysoft/gitfox/pkg/context/ssh"
	"github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/gliderlabs/ssh"
	"github.com/rs/zerolog/log"
)

const sshEnvUserAgent = "SSH_USER_AGENT"

type gitAgentValidator interface {
	Validate() error
}

func ParseGitAgentValidator(ctx context.Context) gitAgentValidator {
	req := ctxrequest.FromContext(ctx)
	if req != nil {
		return &gitHttpAgentValidator{
			ctx: ctx,
			req: req,
		}
	}

	session := ctxssh.FromContext(ctx)
	if session != nil {
		return &gitSSHAgentValidator{
			ctx:     ctx,
			session: session,
		}
	}

	return nil
}

type gitHttpAgentValidator struct {
	ctx context.Context
	req *http.Request
}

func (h *gitHttpAgentValidator) Validate() error {
	ua := h.req.Header.Get("User-Agent")
	if strings.HasPrefix(ua, types.UserAgentFitPrefix) {
		return nil
	}

	if ua == types.UserAgentPipeline {
		return nil
	}

	log.Ctx(h.ctx).Error().Msgf("unexpect user agent %v", ua)
	return store.ErrRequireFitClient
}

type gitSSHAgentValidator struct {
	ctx     context.Context
	session ssh.Session
}

func (s *gitSSHAgentValidator) Validate() error {
	for _, env := range s.session.Environ() {
		frames := strings.SplitN(env, "=", 2)
		if frames[0] == sshEnvUserAgent {
			if strings.HasPrefix(frames[1], types.UserAgentFitPrefix) {
				return nil
			}

			if frames[1] == types.UserAgentPipeline {
				return nil
			}
		}
	}

	return store.ErrRequireFitClient
}
