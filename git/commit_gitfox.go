// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package git

import (
	"context"

	"github.com/easysoft/gitfox/git/api"
)

func (s *Service) ListCommitSHAs(ctx context.Context, params *ListCommitsParams) ([]string, error) {
	if params == nil {
		return nil, ErrNoParamsProvided
	}

	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	gitCommits, err := s.git.ListCommitSHAs(
		ctx,
		repoPath,
		params.AlternateObjectDirs,
		params.GitREF,
		int(params.Page),
		int(params.Limit),
		api.CommitFilter{
			AfterRef:  params.After,
			Path:      params.Path,
			Since:     params.Since,
			Until:     params.Until,
			Committer: params.Committer,
			Author:    params.Author,
			Regex:     params.Regex,
		},
	)
	if err != nil {
		return nil, err
	}

	return gitCommits, nil
}
