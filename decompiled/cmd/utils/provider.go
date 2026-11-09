package utils

import (
	"fmt"
	"os"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
)

// CreateProvider creates a Provider from environment variables
// Priority: standard env vars > QODER_* env vars
func CreateProvider() (provider.Client, string, error) {
	// First check standard OpenAI env vars (highest priority)
	standardOpenAIKey := os.Getenv("OPENAI_API_KEY")
	standardBaseURL := os.Getenv("OPENAI_BASE_URL")
	standardModel := os.Getenv("OPENAI_MODEL")

	// Check QODER_* format env vars
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

	// Prefer standard OpenAI env vars (supports any OpenAI-compatible service)
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

		// If model specified, return model name
		if standardModel != "" {
			log.Debug("Using model from OPENAI_MODEL: %s", standardModel)
			return client, standardModel, nil
		}

		return client, "", nil
	}

	// Then check QODER_OPENAI_* format
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

	// Qoder official service
	if qoderToken != "" {
		log.Debug("Using Qoder personal access token")
		client, err := provider.NewQoderClient(provider.WithAPIKey(qoderToken))
		return client, "", err
	}

	// Other providers (not yet implemented)
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
