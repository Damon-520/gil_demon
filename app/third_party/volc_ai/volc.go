package volc_ai

// 火山引擎大模型官方文档：https://www.volcengine.com/docs/82379/1319853

import (
	"context"
	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"
	"strings"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

// Client 火山引擎 AI 客户端
type Client struct {
	client *arkruntime.Client
	model  string
	log    *logger.ContextLogger
}

// NewClient 创建火山引擎 AI 客户端
func NewClient(c *conf.Conf, log *logger.ContextLogger) *Client {
	return &Client{
		client: arkruntime.NewClientWithApiKey(
			c.VolcAI.APIKey,
			arkruntime.WithBaseUrl(c.VolcAI.BaseURL),
		),
		model: c.VolcAI.Model,
		log:   log,
	}
}

// CQC 检查内容是否合规
func (c *Client) CQC(ctx context.Context, content string) (bool, error) {
	c.log.Info(ctx, "CQC 检查内容: %s，Prompt 版本: %s", content, consts.CQCPromptV1.Version)
	req := &model.CreateChatCompletionRequest{
		Model: c.model,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleSystem,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(consts.CQCPromptV1.SystemPrompt),
				},
			},
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(content),
				},
			},
		},
	}
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		c.log.Error(ctx, "CQC 请求失败: %v", err)
		return false, err
	}
	answer := *resp.Choices[0].Message.Content.StringValue
	c.log.Info(ctx, "CQC 检查内容结果: %s", answer)

	// 根据响应内容判断是否合规
	if strings.Contains(answer, consts.CQCPromptV1.AnswerLegalKeyword) {
		return true, nil
	}
	return false, nil
}
