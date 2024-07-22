package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sashabaranov/go-openai"

	"tg_ai_service/internal/common"
)

func TestDeepSeekService_GetCompletion(t *testing.T) {
	config := &DeepSeekConfig{
		ApiKey:  "test-api-key",
		BaseUrl: "https://api.deepseek.com/v1",
	}
	service, err := NewDeepSeekService(config)
	assert.Nil(t, err)
	assert.NotNil(t, service)

	// Mock the client's CreateChatCompletion method
	service.client = &openai.Client{
		CreateChatCompletionFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			return openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "test response",
						},
					},
				},
			}, nil
		},
	}

	response, err := service.GetCompletion("test prompt", "test system message", "test-model", 0.5, false, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "test response", response)
}
