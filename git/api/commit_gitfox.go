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

package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/easysoft/gitfox/git/command"
	"github.com/easysoft/gitfox/pkg/util/common"
	"github.com/easysoft/gitfox/types"

	"github.com/golang-module/carbon/v2"
)

// CountCommitsWithShortstat lists the commits reachable from ref.
func (g *Git) CountCommitsWithShortstat(
	ctx context.Context,
	repoPath string,
	ref string,
	filter CountCommitFilter,
) ([]types.CommitCount, error) {
	if repoPath == "" {
		return nil, ErrRepositoryPathEmpty
	}
	cmd := command.New("log",
		command.WithFlag("--pretty=format:---%n%aE%n%as"),
		command.WithFlag("--no-merges"),
		command.WithFlag("--shortstat"),
	)
	if len(filter.Since) == 10 {
		since := common.DayStartUnixRFC3339(carbon.Parse(filter.Since))
		cmd.Add(command.WithFlag("--since", since))
	}
	if len(filter.Until) == 10 {
		until := common.DayEndUnixRFC3339(carbon.Parse(filter.Until))
		cmd.Add(command.WithFlag("--until", until))
	}
	if filter.Committer != "" {
		// TODO: this is not correct, we need to filter by committer
		cmd.Add(command.WithFlag("--author", filter.Committer))
	}
	if len(ref) == 0 {
		cmd.Add(command.WithFlag("--branches=*"))
	} else {
		cmd.Add(command.WithFlag("--first-parent", ref))
	}
	output := &bytes.Buffer{}
	err := cmd.Run(ctx, command.WithDir(repoPath), command.WithStdout(output))
	if err != nil {
		return nil, fmt.Errorf("failed to run git to count commit data: %w", err)
	}
	scanner := bufio.NewScanner(output)
	commitStatsMap := make(map[string]map[string]types.CommitCount)
	var currentAuthor, currentDate string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			scanner.Scan()
			currentAuthor = strings.TrimSpace(scanner.Text())
			scanner.Scan()
			currentDate = strings.TrimSpace(scanner.Text())
			scanner.Scan()
			stats := strings.TrimSpace(scanner.Text())
			if currentAuthor == "" || currentDate == "" || stats == "" {
				continue
			}
			if _, exists := commitStatsMap[currentDate]; !exists {
				commitStatsMap[currentDate] = make(map[string]types.CommitCount)
			}
			commitStat, exists := commitStatsMap[currentDate][currentAuthor]
			if !exists {
				commitStat = types.CommitCount{
					Date:   currentDate,
					Author: currentAuthor,
					CommitCountData: types.CommitCountData{
						Commits: 0,
					},
				}
			}
			commitStat.Commits++
			// Parse stats
			// 1 file changed, 1 insertion(+), 1 deletion(-)
			newStats := parseCommitStats(stats)
			// Merge with existing stats if any
			if existing, ok := commitStatsMap[currentDate][currentAuthor]; ok {
				commitStat.Additions = existing.Additions + newStats.Additions
				commitStat.Deletions = existing.Deletions + newStats.Deletions
				commitStat.Changes = existing.Changes + newStats.Changes
				commitStat.Commits = existing.Commits + newStats.Commits
			} else {
				commitStat.Additions = newStats.Additions
				commitStat.Deletions = newStats.Deletions
				commitStat.Changes = newStats.Changes
				commitStat.Commits = newStats.Commits
			}
			commitStatsMap[currentDate][currentAuthor] = commitStat
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning git output: %w", err)
	}
	var commitStats []types.CommitCount
	for date, authorsMap := range commitStatsMap {
		for author, count := range authorsMap {
			commitStats = append(commitStats, types.CommitCount{
				Date:   date,
				Author: author,
				CommitCountData: types.CommitCountData{
					Commits:   count.Commits,
					Additions: count.Additions,
					Deletions: count.Deletions,
					Changes:   count.Changes,
				},
			})
		}
	}
	sort.Slice(commitStats, func(i, j int) bool {
		if commitStats[i].Date == commitStats[j].Date {
			return commitStats[i].Author < commitStats[j].Author
		}
		return commitStats[i].Date < commitStats[j].Date
	})
	return commitStats, nil
}

// parseCommitStats parses the commit stats and returns the commit count data.
func parseCommitStats(stats string) types.CommitCountData {
	var result types.CommitCountData
	result.Commits = 1
	fields := strings.Split(stats, ",")
	for _, field := range fields[1:] {
		parts := strings.Split(strings.TrimSpace(field), " ")
		if len(parts) < 2 {
			continue
		}
		value, _ := strconv.Atoi(parts[0])
		switch {
		case strings.HasPrefix(parts[1], "insertion"):
			result.Additions = value
		case strings.HasPrefix(parts[1], "deletion"):
			result.Deletions = value
		case strings.HasPrefix(parts[1], "changed"):
			result.Changes = value
		}
	}
	return result
}
