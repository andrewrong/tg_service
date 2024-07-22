package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sashabaranov/go-openai"

	"tg_ai_service/internal/common"
)

func TestIntegration_OpenAIService_DeepSeekService(t *testing.T) {
	openAIConfig := &OpenAiConfig{
		ApiKey: "test-api-key",
	}
	openAIService, err := NewOpenAIService(openAIConfig)
	assert.Nil(t, err)
	assert.NotNil(t, openAIService)

	deepSeekConfig := &DeepSeekConfig{
		ApiKey:  "test-api-key",
		BaseUrl: "https://api.deepseek.com/v1",
	}
	deepSeekService, err := NewDeepSeekService(deepSeekConfig)
	assert.Nil(t, err)
	assert.NotNil(t, deepSeekService)

	// Mock the client's CreateChatCompletion method for both services
	openAIService.client = &openai.Client{
		CreateChatCompletionFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			return openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "openai response",
						},
					},
				},
			}, nil
		},
	}

	deepSeekService.client = &openai.Client{
		CreateChatCompletionFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			return openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "deepseek response",
						},
					},
				},
			}, nil
		},
	}

	openAIResponse, err := openAIService.GetCompletion("test prompt", "test system message", "test-model", 0.5, false, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "openai response", openAIResponse)

	deepSeekResponse, err := deepSeekService.GetCompletion("test prompt", "test system message", "test-model", 0.5, false, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "deepseek response", deepSeekResponse)
}
