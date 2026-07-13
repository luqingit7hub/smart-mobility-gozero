package ai

import (
	"context"
	"fmt"

	"common/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// OpenAI 兼容接口
func newChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	cfg := config.DataConfig.AI
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("请在 config.yaml 配置 AI.APIKey")
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-5.5"
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://gptproxy.site/v1"
	}

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: baseURL,
	})
}
