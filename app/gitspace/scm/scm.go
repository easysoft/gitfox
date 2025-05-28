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

package scm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/easysoft/gitfox/git/command"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"
)

var (
	ErrNoDefaultBranch = errors.New("no default branch")
)

const devcontainerDefaultPath = ".devcontainer/devcontainer.json"

var _ SCM = (*scm)(nil)

type SCM interface {
	// GetSCMRepoDetails fetches repository name, credentials & devcontainer config file from the given repo and branch.
	GetSCMRepoDetails(
		ctx context.Context,
		gitspaceConfig types.GitspaceConfig,
	) (*ResolvedDetails, error)

	// CheckValidCodeRepo checks if the current URL is a valid and accessible code repo,
	// input can be connector info, user token etc.
	CheckValidCodeRepo(ctx context.Context, request CodeRepositoryRequest) (*CodeRepositoryResponse, error)
}

type scm struct {
	scmProviderFactory Factory
}

func NewSCM(factory Factory) SCM {
	return &scm{scmProviderFactory: factory}
}

func (s scm) CheckValidCodeRepo(
	ctx context.Context,
	codeRepositoryRequest CodeRepositoryRequest,
) (*CodeRepositoryResponse, error) {
	codeRepositoryResponse := &CodeRepositoryResponse{
		URL:               codeRepositoryRequest.URL,
		CodeRepoIsPrivate: true,
	}
	defaultBranch, err := detectDefaultGitBranch(ctx, codeRepositoryRequest.URL)
	if err == nil {
		branch := "main"
		if defaultBranch != "" {
			branch = defaultBranch
		}
		return &CodeRepositoryResponse{
			URL:               codeRepositoryRequest.URL,
			Branch:            branch,
			CodeRepoIsPrivate: false,
		}, nil
	}
	repoType := enum.CodeRepoTypeUnknown
	if codeRepositoryRequest.RepoType != "" {
		repoType = codeRepositoryRequest.RepoType
	}
	scmProvider, err := s.scmProviderFactory.GetSCMProvider(repoType)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scm provider: %w", err)
	}
	codeRepo := types.CodeRepo{URL: codeRepositoryRequest.URL}
	gitspaceUser := types.GitspaceUser{Identifier: codeRepositoryRequest.UserIdentifier}
	gitspaceConfig := types.GitspaceConfig{
		CodeRepo:     codeRepo,
		SpacePath:    codeRepositoryRequest.SpacePath,
		GitspaceUser: gitspaceUser,
	}
	resolvedCreds, err := scmProvider.ResolveCredentials(ctx, gitspaceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repo credentials and url: %w", err)
	}
	defaultBranch, err = detectDefaultGitBranch(ctx, resolvedCreds.CloneURL)
	if err == nil {
		branch := "main"
		if defaultBranch != "" {
			branch = defaultBranch
		}
		codeRepositoryResponse = &CodeRepositoryResponse{
			URL:               codeRepositoryRequest.URL,
			Branch:            branch,
			CodeRepoIsPrivate: true,
		}
	}
	return codeRepositoryResponse, nil
}

func (s scm) GetSCMRepoDetails(
	ctx context.Context,
	gitspaceConfig types.GitspaceConfig,
) (*ResolvedDetails, error) {
	filePath := devcontainerDefaultPath
	if gitspaceConfig.CodeRepo.Type == "" {
		gitspaceConfig.CodeRepo.Type = enum.CodeRepoTypeUnknown
	}
	scmProvider, err := s.scmProviderFactory.GetSCMProvider(gitspaceConfig.CodeRepo.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scm provider: %w", err)
	}
	resolvedCredentials, err := scmProvider.ResolveCredentials(ctx, gitspaceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repo credentials and url: %w", err)
	}
	var resolvedDetails = &ResolvedDetails{
		ResolvedCredentials: resolvedCredentials,
	}

	catFileOutputBytes, err := scmProvider.GetFileContent(ctx, gitspaceConfig, filePath, resolvedCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to read devcontainer file : %w", err)
	}
	if len(catFileOutputBytes) == 0 {
		resolvedDetails.DevcontainerConfig = &types.DevcontainerConfig{}
	} else {
		sanitizedJSON := removeComments(catFileOutputBytes)
		var config *types.DevcontainerConfig
		if err = json.Unmarshal(sanitizedJSON, &config); err != nil {
			return nil, fmt.Errorf("failed to parse devcontainer json: %w", err)
		}
		resolvedDetails.DevcontainerConfig = config
	}
	return resolvedDetails, nil
}

func removeComments(input []byte) []byte {
	blockCommentRegex := regexp.MustCompile(`(?s)/\*.*?\*/`)
	input = blockCommentRegex.ReplaceAll(input, nil)
	lineCommentRegex := regexp.MustCompile(`//.*`)
	return lineCommentRegex.ReplaceAll(input, nil)
}

func detectDefaultGitBranch(ctx context.Context, gitRepoDir string) (string, error) {
	cmd := command.New("ls-remote",
		command.WithFlag("--symref"),
		command.WithFlag("-q"),
		command.WithArg(gitRepoDir),
		command.WithArg("HEAD"),
	)
	output := &bytes.Buffer{}
	if err := cmd.Run(ctx, command.WithStdout(output)); err != nil {
		return "", fmt.Errorf("failed to ls remote repo")
	}
	var lsRemoteHeadRegexp = regexp.MustCompile(`ref: refs/heads/([^\s]+)\s+HEAD`)
	match := lsRemoteHeadRegexp.FindStringSubmatch(strings.TrimSpace(output.String()))
	if match == nil {
		return "", ErrNoDefaultBranch
	}
	return match[1], nil
}
