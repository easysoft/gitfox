// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"context"
	"strconv"

	"github.com/easysoft/gitfox/git/command"
)

var excludeFromDiff = []string{
	"package-lock.json",
	"pnpm-lock.yaml",
	// yarn.lock, Cargo.lock, Gemfile.lock, Pipfile.lock, etc.
	"*.lock",
	"go.sum",
}

func (g *Git) excludeFiles(excludeList []string) []string {
	var excludedFiles []string
	// TODO: add more exclude files
	excludeList = append(excludeList, excludeFromDiff...)
	for _, f := range excludeList {
		excludedFiles = append(excludedFiles, ":(exclude,top)"+f)
	}
	return excludedFiles
}

func (g *Git) DiffFiles(ctx context.Context,
	repoPath string,
	baseRef string,
	headRef string,
	unified int,
	excludeList []string,
) (string, error) {
	cmd := command.New("diff",
		command.WithFlag("--ignore-all-space"),
		command.WithFlag("--diff-algorithm=minimal"),
		// command.WithFlag("--diff-filter=ad"),
		command.WithFlag("--unified="+strconv.Itoa(unified)),
	)
	cmd.Add(command.WithArg(baseRef, headRef))
	cmd.Add(command.WithArg(g.excludeFiles(excludeList)...))
	stdout := &bytes.Buffer{}
	err := cmd.Run(ctx,
		command.WithDir(repoPath),
		command.WithStdout(stdout),
	)
	if err != nil {
		return "", processGitErrorf(err, "failed to trigger diff command")
	}

	return stdout.String(), nil
}
