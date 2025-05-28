// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/easysoft/gitfox/app/services/settings"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/pkg/util/common"
	"github.com/easysoft/gitfox/pkg/util/dbtx"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/rs/zerolog/log"
)

const (
	CodeReviewTemplate = `以下是待审核的代码补丁，请根据代码上下文进行评审。如有任何错误风险、安全漏洞,请帮助我指出具体哪个文件存在什么问题，并给出修改建议;没有问题请回复/lgtm, 仅此情况下回复这个。
。
待审核的代码补丁：
{{ .file_diffs }}`

	DefaultDiffUnified = 3
	MaxCommentLength   = 64 << 10
	DefaultAITimeout   = 30 * time.Second
)

func (s *Service) aiReviewPullReq(ctx context.Context, pr *types.PullReq) error {
	log.Ctx(ctx).Info().
		Int64("repo_id", pr.TargetRepoID).
		Int64("pr_id", pr.ID).
		Msg("starting AI code review")

	// 查space
	repoID := pr.TargetRepoID
	repo, err := s.repoStore.Find(ctx, repoID)
	if err != nil {
		return err
	}
	//  如果未开启AI审核，则直接返回
	aiReviewEnabled, _ := settings.RepoGet(
		ctx,
		s.settings,
		repo.ID,
		settings.KeyAIReviewEnabled,
		settings.DefaultAIReviewEnabled,
	)
	if !aiReviewEnabled {
		log.Ctx(ctx).Debug().Msgf("AI review is disable for repo %d, skipping", repo.ID)
		return nil
	}
	now := time.Now().UnixMilli()
	aiReq := &types.AIRequest{
		Created:       now,
		Updated:       now,
		RepoID:        repoID,
		PullReqID:     pr.ID,
		Status:        enum.AIRequestStatusSuccess,
		ReviewStatus:  enum.AIRequestReviewUnknown,
		ReviewMessage: "网络错误, 请稍后再试",
		ReviewSHA:     pr.SourceSHA,
	}
	defer s.saveAIRequest(ctx, aiReq)
	// 查ai config
	cfg, err := s.aiStore.Default(ctx, repo.ParentID)
	if err != nil {
		return s.handleAIError(aiReq, err)
	}
	aiReq.ConfigID = cfg.ID
	// 调用ai
	client, err := common.GetClient(cfg)
	if err != nil {
		return s.handleAIError(aiReq, err)
	}
	// 获取diff文件数
	diffCount, err := s.git.DiffFileNames(ctx, &git.DiffParams{
		ReadParams: git.CreateReadParams(repo),
		BaseRef:    pr.MergeBaseSHA,
		HeadRef:    pr.SourceSHA,
	})
	if err != nil {
		return s.handleAIError(aiReq, fmt.Errorf("failed to get diff count: %w", err))
	}
	log.Ctx(ctx).Debug().Msgf("diff count: %d", len(diffCount.Files))
	if len(diffCount.Files) > 10 {
		aiReq.ReviewStatus = enum.AIRequestReviewIgnore
		aiReq.ReviewMessage = "变更文件过多, 忽略本次评审"
		log.Ctx(ctx).Debug().Msgf("diff count is too large %d > 10, skipping ai review", len(diffCount.Files))
		return nil
	}
	// 获取diff文件
	diff, err := s.git.DiffFiles(ctx, &git.DiffFilesParams{
		ReadParams:      git.CreateReadParams(repo),
		TargetCommitSHA: pr.MergeBaseSHA,
		SourceCommitSHA: pr.SourceSHA,
		DiffUnified:     DefaultDiffUnified,
	})
	if err != nil {
		return s.handleAIError(aiReq, fmt.Errorf("failed to get diff files: %w", err))
	}
	out, _ := getTemplateByString(
		CodeReviewTemplate,
		map[string]interface{}{
			"file_diffs": diff,
		},
	)
	log.Ctx(ctx).Trace().Msgf("ai review request: %s", out)
	if strings.HasPrefix(out, "\\ No newline at end of file") {
		out = strings.Replace(out, "\\ No newline at end of file", "", 1)
	}
	resp, err := client.Completion(ctx, out)
	if err != nil {
		return s.handleAIError(aiReq, err)
	}
	aiReq.Duration = time.Now().UnixMilli() - now
	aiReq.Token = int64(resp.Usage.TotalTokens)
	log.Ctx(ctx).Trace().Msgf("ai review response: %v", resp.Content)
	aiReq.ReviewMessage = resp.Content
	aiReq.ReviewStatus = enum.AIRequestReviewRejected
	if strings.Contains(resp.Content, "/lgtm") {
		aiReq.ReviewStatus = enum.AIRequestReviewApproved
	}
	// err = s.createPRComment(ctx, pr, resp.Content)
	// if err != nil {
	// 	return s.handleAIError(aiReq, fmt.Errorf("failed to create AI review comment: %w", err))
	// }

	log.Ctx(ctx).Info().
		Int64("duration_ms", aiReq.Duration).
		Int64("tokens", aiReq.Token).
		Msg("AI review completed successfully")

	return nil
}

func getTemplateByString(tpl string, data map[string]interface{}) (string, error) {
	var b bytes.Buffer
	t := template.Must(template.New("k3s").Parse(tpl))
	err := t.Execute(&b, &data)
	return b.String(), err
}

func (s *Service) createPRComment(ctx context.Context, pr *types.PullReq, comment string) error {
	now := time.Now().UnixMilli()
	bot, err := s.principalStore.FindByUID(ctx, "bot")
	if err != nil {
		return err
	}
	// This limit is deliberately larger than the limit in our API.
	if len(comment) > MaxCommentLength {
		log.Ctx(ctx).Warn().Msgf("comment is too long, truncate it: %d", len(comment))
		comment = comment[:MaxCommentLength]
	}
	// 创建评论活动记录
	// TODO: 评论活动记录的ResolvedBy和Resolved字段需要设置
	act := &types.PullReqActivity{
		Created:     now,
		Updated:     now,
		Edited:      now,
		CreatedBy:   bot.ID,
		RepoID:      pr.TargetRepoID,
		PullReqID:   pr.ID,
		Type:        enum.PullReqActivityTypeComment,
		Kind:        enum.PullReqActivityKindComment,
		Text:        comment,
		PayloadRaw:  json.RawMessage("{}"),
		Metadata:    nil,
		ResolvedBy:  nil, // &bot.ID,
		Resolved:    nil, // &now,
		CodeComment: nil,
		Mentions:    nil,
	}
	// 开启事务
	err = dbtx.TxOptLock(ctx, s.tx, func(ctx context.Context) error {
		// 更新PR活动序号
		prUpd, err := s.pullreqStore.UpdateActivitySeq(ctx, pr)
		if err != nil {
			return fmt.Errorf("failed to get pull request activity number: %w", err)
		}
		*pr = *prUpd

		// 设置活动顺序号
		act.Order = prUpd.ActivitySeq

		// 创建活动记录
		err = s.activityStore.Create(ctx, act)
		if err != nil {
			return fmt.Errorf("failed to create pull request activity: %w", err)
		}

		// 更新PR评论计数
		pr.CommentCount++
		err = s.pullreqStore.Update(ctx, pr)
		if err != nil {
			return fmt.Errorf("failed to increment pull request comment count: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) handleAIError(aiReq *types.AIRequest, err error) error {
	aiReq.Status = enum.AIRequestStatusFailed
	aiReq.Error = err.Error()
	return err
}

func (s *Service) saveAIRequest(ctx context.Context, aiReq *types.AIRequest) {
	if err := s.aiStore.Record(ctx, aiReq); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to save ai request record")
	}
}
