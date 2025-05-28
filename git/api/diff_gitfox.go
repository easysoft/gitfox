// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/easysoft/gitfox/errors"
	"github.com/easysoft/gitfox/git/command"
)

type CommitChangeLines struct {
	AddLines int64
	DelLines int64
	FileName string
}

// ReceiveCommitNumStat will stream diff for provided ref.
func (g *Git) ReceiveCommitNumStat(
	ctx context.Context,
	repoPath string,
	rev string,
	objectsDir []string,
) ([]*CommitChangeLines, error) {
	if repoPath == "" {
		return nil, ErrRepositoryPathEmpty
	}
	if rev == "" {
		return nil, errors.InvalidArgument("git revision cannot be empty")
	}

	cmd := command.New("show",
		command.WithFlag("--numstat"),
		command.WithFlag("--pretty="),
		command.WithFlag("--diff-filter=ad"),
		command.WithArg(rev),
	)

	alterDirs := make([]string, len(objectsDir))
	for i, v := range objectsDir {
		alterDirs[i] = filepath.Dir(v)
	}
	cmd.Add(command.WithEnv("GIT_OBJECT_DIRECTORY", strings.Join(objectsDir, ":")))
	cmd.Add(command.WithAlternateObjectDirs(alterDirs...))

	stdout := &bytes.Buffer{}
	if err := cmd.Run(ctx,
		command.WithDir(repoPath),
		command.WithStdout(stdout),
	); err != nil {
		return nil, processGitErrorf(err, "commit diff error")
	}
	return parseChangeLines(stdout.Bytes()), nil
}

func (g *Git) ReceiveDiffSNumStatus(ctx context.Context,
	repoPath string,
	baseRef string,
	headRef string,
	objectsDir []string,
) ([]*CommitChangeLines, error) {
	alterDirs := make([]string, len(objectsDir))
	for i, v := range objectsDir {
		alterDirs[i] = filepath.Dir(v)
	}
	cmd := command.New("diff", command.WithFlag("--numstat"), command.WithFlag("--diff-filter=ad"))
	cmd.Add(command.WithEnv("GIT_OBJECT_DIRECTORY", strings.Join(objectsDir, ":")))
	cmd.Add(command.WithAlternateObjectDirs(alterDirs...))
	cmd.Add(command.WithArg(baseRef, headRef))

	stdout := &bytes.Buffer{}
	err := cmd.Run(ctx,
		command.WithDir(repoPath),
		command.WithStdout(stdout),
	)
	if err != nil {
		return nil, processGitErrorf(err, "failed to trigger diff command")
	}

	return parseChangeLines(stdout.Bytes()), nil
}

func parseChangeLines(in []byte) []*CommitChangeLines {
	result := make([]*CommitChangeLines, 0)
	for _, byteLine := range bytes.Split(in, []byte("\n")) {
		if len(byteLine) == 0 {
			continue
		}

		frames := strings.Split(string(byteLine), "\t")
		addNum, _ := strconv.ParseInt(frames[0], 10, 64)
		delNum, _ := strconv.ParseInt(frames[1], 10, 64)
		item := &CommitChangeLines{
			AddLines: addNum,
			DelLines: delNum,
			FileName: frames[2],
		}
		result = append(result, item)
	}
	return result
}
