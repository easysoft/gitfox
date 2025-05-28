// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package types

import (
	"errors"

	"github.com/golang-module/carbon/v2"
)

type RepositoryStatisticsFilter struct {
	// IncludeMergeRequest bool
	// IncludePushRequest  bool
	// IncludeReview       bool
	// IncludeCode         bool
	Branch string
	// Period    enum.PeriodSortOption
	BeginTime string
	EndTime   string
	// Top       int64
	Author string
}

type RepositoryStatistics struct {
	CommitStats []CommitCount
	Committer   string `json:"author,omitempty"`
	Total       int    `json:"total,omitempty"`
}

type StatisticsCommitStat struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type StatisticsCodeFrequencyStat struct {
	Key       string `json:"key"`
	Commits   int    `json:"commits"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
}

type TimeRangeInput struct {
	BeginTime string `json:"begin_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	Period    string `json:"period,omitempty"`
}

func (t *TimeRangeInput) ValidateTimeRange() error {
	if len(t.BeginTime) == 0 {
		t.BeginTime = carbon.Now().SubMonth().ToDateString()
	} else {
		t.BeginTime = carbon.Parse(t.BeginTime).ToDateString()
	}
	if len(t.EndTime) == 0 {
		t.EndTime = carbon.Now().ToDateString()
	} else {
		t.EndTime = carbon.Parse(t.EndTime).ToDateString()
	}

	if len(t.Period) == 0 {
		beginDate := carbon.Parse(t.BeginTime)
		endDate := carbon.Parse(t.EndTime)
		daysDiff := beginDate.DiffInDays(endDate)
		if daysDiff > 31 {
			t.Period = "month"
		} else {
			t.Period = "day"
		}
	}
	return nil
}

type RepoActiveInput struct {
	TimeRangeInput
	Repos []string `json:"repos"`
}

type RepoActiveOutput struct {
	RepoCount   int `json:"repo_count"`
	CommitCount int `json:"commit_count"`
	UserCount   int `json:"user_count"`
}

func (in *RepoActiveInput) Validate() error {
	if err := in.ValidateTimeRange(); err != nil {
		return err
	}
	if len(in.Repos) == 0 {
		return errors.New("repos is required")
	}
	return nil
}

type RepoCommitsInput struct {
	TimeRangeInput
	Repo string `json:"repo"`
}

func (in *RepoCommitsInput) Validate() error {
	if len(in.Repo) == 0 {
		return errors.New("repo is required")
	}
	if err := in.ValidateTimeRange(); err != nil {
		return err
	}
	return nil
}

type UserCodeFrequencyInput struct {
	TimeRangeInput
	User string `json:"user"`
}

func (in *UserCodeFrequencyInput) Validate() error {
	if len(in.User) == 0 {
		return errors.New("user is required")
	}
	if err := in.ValidateTimeRange(); err != nil {
		return err
	}
	return nil
}

type RepoCommitsOutput struct {
	Total       int                    `json:"total"`
	RepoPath    string                 `json:"repo_path"`
	RepoID      int64                  `json:"repo_id"`
	Period      string                 `json:"period"`
	CommitStats []StatisticsCommitStat `json:"stats"`
}

type UserCodeFrequencyOutput struct {
	Total       int                           `json:"total"`
	User        string                        `json:"user"`
	Period      string                        `json:"period"`
	CommitStats []StatisticsCodeFrequencyStat `json:"stats"`
}

type RepoCodeFrequencyOutput struct {
	Total       int                           `json:"total"`
	RepoPath    string                        `json:"repo_path"`
	RepoID      int64                         `json:"repo_id"`
	Period      string                        `json:"period"`
	CommitStats []StatisticsCodeFrequencyStat `json:"stats"`
}

type RepoCommitUsersOutput struct {
	Total       int                    `json:"total"`
	RepoPath    string                 `json:"repo_path"`
	RepoID      int64                  `json:"repo_id"`
	CommitStats []StatisticsCommitStat `json:"stats"`
}

type RepoTop10Output struct {
	RepoID        string `json:"repo_id"`
	RepoName      string `json:"repo_name"`
	CommitCount   int    `json:"commit_count"`
	BranchCount   int    `json:"branch_count"`
	TagCount      int    `json:"tag_count"`
	PullReqCount  int    `json:"pull_req_count"`
	PushReqCount  int    `json:"push_req_count"`
	TotalReqCount int    `json:"total_req_count"`
}

type SystemCommitsInput struct {
	TimeRangeInput
}

func (in *SystemCommitsInput) Validate() error {
	if err := in.ValidateTimeRange(); err != nil {
		return err
	}
	return nil
}

type SystemCommitsOutput struct {
	Total       int                    `json:"total"`
	Period      string                 `json:"period"`
	CommitStats []StatisticsCommitStat `json:"stats"`
}
