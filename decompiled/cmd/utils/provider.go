package utils

import (
	"fmt"
	"os"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
)

// CreateProvider 根据环境变量创建 Provider
// 优先级：标准环境变量 > QODER_* 环境变量
func CreateProvider() (provider.Client, string, error) {
	// 首先检查标准 OpenAI 环境变量（优先级最高）
	standardOpenAIKey := os.Getenv("OPENAI_API_KEY")
	standardBaseURL := os.Getenv("OPENAI_BASE_URL")
	standardModel := os.Getenv("OPENAI_MODEL")

	// 检查 QODER_* 格式的环境变量
	qoderToken := os.Getenv("QODER_PERSONAL_ACCESS_TOKEN")
	anthropicKey := os.Getenv("QODER_ANTHROPIC_API_KEY")
	qoderOpenAIKey := os.Getenv("QODER_OPENAI_API_KEY")
	qoderBaseURL := os.Getenv("QODER_OPENAI_BASE_URL")
	qoderModel := os.Getenv("QODER_OPENAI_MODEL")
	dashscopeKey := os.Getenv("QODER_DASHSCOPE_API_KEY")
	idealabKey := os.Getenv("QODER_IDEALAB_API_KEY")

	log.Debug("Checking environment variables for API keys")
	if standardOpenAIKey != "" {
		log.Debug("Found OPENAI_API_KEY, using standard OpenAI provider")
	}
	if qoderOpenAIKey != "" {
		log.Debug("Found QODER_OPENAI_API_KEY, using Qoder OpenAI provider")
	}

	// 优先使用标准 OpenAI 环境变量（支持任意 OpenAI 兼容服务）
	if standardOpenAIKey != "" {
		var opts []provider.ClientOption
		opts = append(opts, provider.WithAPIKey(standardOpenAIKey))

		if standardBaseURL != "" {
			opts = append(opts, provider.WithBaseURL(standardBaseURL))
			log.Debug("Using custom base URL: %s", standardBaseURL)
		}

		client, err := provider.NewOpenAIClient(opts...)
		if err != nil {
			return nil, "", fmt.Errorf("create OpenAI client: %w", err)
		}

		// 如果指定了模型，返回模型名
		if standardModel != "" {
			log.Debug("Using model from OPENAI_MODEL: %s", standardModel)
			return client, standardModel, nil
		}

		return client, "", nil
	}

	// 其次检查 QODER_OPENAI_* 格式
	if qoderOpenAIKey != "" {
		var opts []provider.ClientOption
		opts = append(opts, provider.WithAPIKey(qoderOpenAIKey))

		if qoderBaseURL != "" {
			opts = append(opts, provider.WithBaseURL(qoderBaseURL))
			log.Debug("Using Qoder base URL: %s", qoderBaseURL)
		}

		client, err := provider.NewOpenAIClient(opts...)
		if err != nil {
			return nil, "", fmt.Errorf("create OpenAI client: %w", err)
		}

		if qoderModel != "" {
			log.Debug("Using model from QODER_OPENAI_MODEL: %s", qoderModel)
			return client, qoderModel, nil
		}

		return client, "", nil
	}

	// Qoder 官方服务
	if qoderToken != "" {
		log.Debug("Using Qoder personal access token")
		client, err := provider.NewQoderClient(provider.WithAPIKey(qoderToken))
		return client, "", err
	}

	// 其他 Provider（暂未实现）
	if anthropicKey != "" {
		return nil, "", fmt.Errorf("Anthropic provider: use OPENAI_API_KEY with compatible base URL instead")
	}
	if dashscopeKey != "" {
		return nil, "", fmt.Errorf("DashScope provider: use OPENAI_API_KEY with https://dashscope.aliyuncs.com/compatible-mode/v1 as base URL")
	}
	if idealabKey != "" {
		return nil, "", fmt.Errorf("IdeaLab provider not implemented yet")
	}

	return nil, "", fmt.Errorf("no API key found. Set OPENAI_API_KEY or QODER_PERSONAL_ACCESS_TOKEN")
}
