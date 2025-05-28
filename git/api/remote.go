// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"context"
	"strings"

	giturl "github.com/easysoft/gitfox/git/api/url"
	"github.com/easysoft/gitfox/git/command"
)

// GetRemoteAddress returns remote url of git repository in the repoPath with special remote name
func (g *Git) GetRemoteAddress(ctx context.Context, repoPath, remoteName string) (string, error) {
	// > 2.7 git remote get-url origin
	// < 2.7 git config --get remote.origin.url

	cmd := command.New("remote",
		command.WithFlag("get-url"),
		command.WithArg(remoteName),
	)
	output := &bytes.Buffer{}

	err := cmd.Run(ctx, command.WithDir(repoPath), command.WithStdout(output))
	if err != nil {
		return "", err
	}
	result := strings.TrimSpace(output.String())
	if len(result) > 0 {
		result = result[:len(result)-1]
	}
	return result, nil
}

// GetRemoteURL returns the url of a specific remote of the repository.
func (g *Git) GetRemoteURL(ctx context.Context, repoPath, remoteName string) (*giturl.GitURL, error) {
	addr, err := g.GetRemoteAddress(ctx, repoPath, remoteName)
	if err != nil {
		return nil, err
	}
	return giturl.Parse(addr)
}
