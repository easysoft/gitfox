// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package inputs

import (
	harness "github.com/easysoft/gitfox/internal/pipeline/spec"

	"github.com/drone/drone-go/drone"
)

// Build converts a build to input variables.
func Build(v *drone.Build) map[string]interface{} {
	return map[string]interface{}{
		"event":         v.Event,
		"number":        v.Number,
		"action":        v.Action,
		"cron":          v.Cron,
		"environment":   v.Deploy,
		"link":          v.Link,
		"branch":        v.Target,
		"source":        v.Source,
		"before":        v.Before,
		"after":         v.After,
		"target":        v.Target,
		"ref":           v.Ref,
		"commit":        v.After,
		"title":         v.Title,
		"message":       v.Message,
		"source_repo":   v.Fork,
		"author_login":  v.Author,
		"author_name":   v.AuthorName,
		"author_email":  v.AuthorEmail,
		"author_avatar": v.AuthorAvatar,
		"sender":        v.Sender,
		"debug":         v.Debug,
		"params":        v.Params,
	}
}

// Build converts a build to input variables.
func Repo(v *drone.Repo) map[string]interface{} {
	return map[string]interface{}{
		"uid":                  v.UID,
		"name":                 v.Name,
		"namespace":            v.Namespace,
		"slug":                 v.Slug,
		"git_http_url":         v.HTTPURL,
		"git_ssh_url":          v.SSHURL,
		"link":                 v.Link,
		"branch":               v.Branch,
		"config":               v.Config,
		"private":              v.Private,
		"visibility":           v.Visibility,
		"active":               v.Active,
		"trusted":              v.Trusted,
		"protected":            v.Protected,
		"ignore_forks":         v.IgnoreForks,
		"ignore_pull_requests": v.IgnorePulls,
	}
}

// Input converts a Inputs to input variables.
func Inputs(in1 map[string]*harness.Input, in2 map[string]string) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range in1 {
		out[k] = v.Default
	}
	for k, v := range in2 {
		out[k] = v
	}
	return out
}
