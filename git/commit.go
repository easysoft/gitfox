// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/easysoft/gitfox/errors"
	"github.com/easysoft/gitfox/git/api"
	"github.com/easysoft/gitfox/git/command"
	"github.com/easysoft/gitfox/git/enum"
	"github.com/easysoft/gitfox/git/parser"
	"github.com/easysoft/gitfox/git/sha"
	"github.com/easysoft/gitfox/types"
)

func CommitMessage(subject, body string) string {
	if body == "" {
		return subject
	}
	return subject + "\n\n" + body
}

type GetCommitParams struct {
	ReadParams
	Revision string
}

type Commit struct {
	SHA        sha.SHA           `json:"sha"`
	ParentSHAs []sha.SHA         `json:"parent_shas,omitempty"`
	Title      string            `json:"title"`
	Message    string            `json:"message,omitempty"`
	Author     Signature         `json:"author"`
	Committer  Signature         `json:"committer"`
	FileStats  []CommitFileStats `json:"file_stats,omitempty"`
}

type GetCommitOutput struct {
	Commit Commit `json:"commit"`
}

type Signature struct {
	Identity Identity  `json:"identity"`
	When     time.Time `json:"when"`
}

type Identity struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (i *Identity) Validate() error {
	if i.Name == "" {
		return errors.InvalidArgument("identity name is mandatory")
	}

	if i.Email == "" {
		return errors.InvalidArgument("identity email is mandatory")
	}

	return nil
}

func (s *Service) GetCommit(ctx context.Context, params *GetCommitParams) (*GetCommitOutput, error) {
	if params == nil {
		return nil, ErrNoParamsProvided
	}
	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)
	result, err := s.git.GetCommit(ctx, repoPath, params.Revision)
	if err != nil {
		return nil, err
	}

	commit, err := mapCommit(result)
	if err != nil {
		return nil, fmt.Errorf("failed to map rpc commit: %w", err)
	}

	return &GetCommitOutput{
		Commit: *commit,
	}, nil
}

type ListCommitsParams struct {
	ReadParams
	// GitREF is a git reference (branch / tag / commit SHA)
	GitREF string
	// After is a git reference (branch / tag / commit SHA)
	// If provided, commits only up to that reference will be returned (exlusive)
	After string
	Page  int32
	Limit int32
	Path  string

	// Since allows to filter for commits since the provided UNIX timestamp - Optional, ignored if value is 0.
	Since int64

	// Until allows to filter for commits until the provided UNIX timestamp - Optional, ignored if value is 0.
	Until int64

	// Committer allows to filter for commits based on the committer - Optional, ignored if string is empty.
	Committer string

	// Author allows to filter for commits based on the author - Optional, ignored if string is empty.
	Author string

	// IncludeStats allows to include information about inserted, deletions and status for changed files.
	IncludeStats bool

	// Regex allows to use regular expression in the Committer and Author fields
	Regex bool
}

type RenameDetails struct {
	OldPath         string
	NewPath         string
	CommitShaBefore sha.SHA
	CommitShaAfter  sha.SHA
}

type ListCommitsOutput struct {
	Commits       []Commit
	RenameDetails []*RenameDetails
	TotalCommits  int
}

type CommitFileStats struct {
	Status     enum.FileDiffStatus
	Path       string
	OldPath    string // populated only in case of renames
	Insertions int64
	Deletions  int64
}

func (s *Service) ListCommits(ctx context.Context, params *ListCommitsParams) (*ListCommitsOutput, error) {
	if params == nil {
		return nil, ErrNoParamsProvided
	}

	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	gitCommits, renameDetails, err := s.git.ListCommits(
		ctx,
		repoPath,
		params.GitREF,
		int(params.Page),
		int(params.Limit),
		params.IncludeStats,
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

	// try to get total commits between gitref and After refs
	totalCommits := 0
	if params.Page == 1 && len(gitCommits) < int(params.Limit) {
		totalCommits = len(gitCommits)
	} else if params.After != "" && params.GitREF != params.After {
		div, err := s.git.GetCommitDivergences(ctx, repoPath, []api.CommitDivergenceRequest{
			{From: params.GitREF, To: params.After},
		}, 0)
		if err != nil {
			return nil, err
		}
		if len(div) > 0 {
			totalCommits = int(div[0].Ahead)
		}
	}

	commits := make([]Commit, len(gitCommits))
	for i := range gitCommits {
		commit, err := mapCommit(gitCommits[i])
		if err != nil {
			return nil, fmt.Errorf("failed to map rpc commit: %w", err)
		}

		commits[i] = *commit
	}

	return &ListCommitsOutput{
		Commits:       commits,
		RenameDetails: mapRenameDetails(renameDetails),
		TotalCommits:  totalCommits,
	}, nil
}

type GetCommitDivergencesParams struct {
	ReadParams
	MaxCount int32
	Requests []CommitDivergenceRequest
}

type GetCommitDivergencesOutput struct {
	Divergences []api.CommitDivergence
}

// CommitDivergenceRequest contains the refs for which the converging commits should be counted.
type CommitDivergenceRequest struct {
	// From is the ref from which the counting of the diverging commits starts.
	From string
	// To is the ref at which the counting of the diverging commits ends.
	To string
}

// CommitDivergence contains the information of the count of converging commits between two refs.
type CommitDivergence struct {
	// Ahead is the count of commits the 'From' ref is ahead of the 'To' ref.
	Ahead int32
	// Behind is the count of commits the 'From' ref is behind the 'To' ref.
	Behind int32
}

func (s *Service) GetCommitDivergences(
	ctx context.Context,
	params *GetCommitDivergencesParams,
) (*GetCommitDivergencesOutput, error) {
	if params == nil {
		return nil, ErrNoParamsProvided
	}

	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	requests := make([]api.CommitDivergenceRequest, len(params.Requests))
	for i, req := range params.Requests {
		requests[i] = api.CommitDivergenceRequest{
			From: req.From,
			To:   req.To,
		}
	}

	divergences, err := s.git.GetCommitDivergences(
		ctx,
		repoPath,
		requests,
		params.MaxCount,
	)
	if err != nil {
		return nil, err
	}

	return &GetCommitDivergencesOutput{
		Divergences: divergences,
	}, nil
}

type FindOversizeFilesParams struct {
	RepoUID       string
	GitObjectDirs []string
	SizeLimit     int64
}

type FindOversizeFilesOutput struct {
	FileInfos []FileInfo
}

type FileInfo struct {
	SHA  sha.SHA
	Size int64
}

//nolint:gocognit
func (s *Service) FindOversizeFiles(
	ctx context.Context,
	params *FindOversizeFilesParams,
) (*FindOversizeFilesOutput, error) {
	if params.RepoUID == "" {
		return nil, api.ErrRepositoryPathEmpty
	}
	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	var fileInfos []FileInfo
	for _, gitObjDir := range params.GitObjectDirs {
		objects, err := catFileBatchCheckAllObjects(ctx, repoPath, gitObjDir)
		if err != nil {
			return nil, err
		}

		for _, obj := range objects {
			if obj.Type == string(TreeNodeTypeBlob) {
				if obj.Size > params.SizeLimit {
					fileInfos = append(fileInfos, FileInfo{
						SHA:  obj.SHA,
						Size: obj.Size,
					})
				}
			}
		}
	}

	return &FindOversizeFilesOutput{
		FileInfos: fileInfos,
	}, nil
}

func catFileBatchCheckAllObjects(
	ctx context.Context,
	repoPath string,
	gitObjDir string,
) ([]parser.BatchCheckObject, error) {
	// "info/alternates" points to the original repository.
	const oldFilename = "/info/alternates"
	const newFilename = "/info/alternates.bkp"

	// --batch-all-objects reports objects in the current repository and in all alternate directories.
	// We want to report objects in the current repository only.
	if err := os.Rename(gitObjDir+oldFilename, gitObjDir+newFilename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("failed to rename %s to %s: %w", oldFilename, newFilename, err)
	}

	cmd := command.New("cat-file",
		command.WithFlag("--batch-check"),
		command.WithFlag("--batch-all-objects"),
		command.WithFlag("--unordered"),
		command.WithFlag("-Z"),
		command.WithEnv(command.GitObjectDir, gitObjDir),
	)
	buffer := bytes.NewBuffer(nil)
	err := cmd.Run(
		ctx,
		command.WithDir(repoPath),
		command.WithStdout(buffer),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to cat-file batch check all objects: %w", err)
	}

	objects, err := parser.CatFileBatchCheckAllObjects(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output of cat-file batch check all objects: %w", err)
	}

	if err := os.Rename(gitObjDir+newFilename, gitObjDir+oldFilename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("failed to rename %s to %s: %w", newFilename, oldFilename, err)
	}

	return objects, nil
}

type CountCommitsParams struct {
	ReadParams
	// GitREF is a git reference (branch / tag / commit SHA)
	GitREF string
	// Since allows to filter for commits since the provided
	Since string
	// Until allows to filter for commits until the provided
	Until string
	// Committer allows to filter for commits based on the committer - Optional, ignored if string is empty.
	Committer string
}

type CountCommitsOutput struct {
	CommitStats []types.CommitCount
	Total       int
	AuthorCount map[string]int
}

func (s *Service) CountCommits(ctx context.Context, params *CountCommitsParams) (*CountCommitsOutput, error) {
	if params == nil {
		return nil, ErrNoParamsProvided
	}
	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)
	commitsStat, err := s.git.CountCommits(
		ctx,
		repoPath,
		params.GitREF,
		api.CountCommitFilter{
			Since:     params.Since,
			Until:     params.Until,
			Committer: params.Committer,
		},
	)
	if err != nil {
		return nil, err
	}
	countCommitsOutput := &CountCommitsOutput{
		CommitStats: commitsStat,
		AuthorCount: make(map[string]int),
	}

	total := 0
	for _, c := range commitsStat {
		total += c.Commits
		countCommitsOutput.AuthorCount[c.Author] += c.Commits
	}
	countCommitsOutput.Total = total
	return countCommitsOutput, nil
}

type CountCommitsShortstatOutput struct {
	CommitStats []types.CommitCount
	Total       int
	AuthorCount map[string]types.CommitCountData
}

func (s *Service) CountCommitsWithShortstat(ctx context.Context, params *CountCommitsParams) (*CountCommitsShortstatOutput, error) {
	if params == nil {
		return nil, ErrNoParamsProvided
	}
	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)
	commitsStat, err := s.git.CountCommitsWithShortstat(
		ctx,
		repoPath,
		params.GitREF,
		api.CountCommitFilter{
			Since:     params.Since,
			Until:     params.Until,
			Committer: params.Committer,
		},
	)
	if err != nil {
		return nil, err
	}
	countCommitsOutput := &CountCommitsShortstatOutput{
		CommitStats: commitsStat,
		AuthorCount: make(map[string]types.CommitCountData),
	}

	total := 0
	for _, c := range commitsStat {
		total += c.Commits
		authorData := countCommitsOutput.AuthorCount[c.Author]
		authorData.Commits += c.Commits
		authorData.Additions += c.Additions
		authorData.Deletions += c.Deletions
		authorData.Changes += c.Changes
		countCommitsOutput.AuthorCount[c.Author] = authorData
	}
	countCommitsOutput.Total = total
	return countCommitsOutput, nil
}
