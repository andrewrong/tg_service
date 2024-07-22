package ai

import (
	"context"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"tg_ai_service/internal/common"
)

func TestDeepSeekConfig_Check(t *testing.T) {
	config := &DeepSeekConfig{ApiKey: ""}
	err := config.Check()
	assert.NotNil(t, err)
	assert.Equal(t, common.ParameterError, err.ErrType)
	assert.Equal(t, "api key is empty", err.ErrMsg)

	config.ApiKey = "test_api_key"
	err = config.Check()
	assert.Nil(t, err)
	assert.Equal(t, "https://api.deepseek.com/v1", config.BaseUrl)
}

func TestNewDeepSeekService(t *testing.T) {
	config := &DeepSeekConfig{ApiKey: ""}
	service, err := NewDeepSeekService(config)
	assert.Nil(t, service)
	assert.NotNil(t, err)
	assert.Equal(t, common.ParameterError, err.ErrType)
	assert.Equal(t, "config is empty", err.ErrMsg)

	config.ApiKey = "test_api_key"
	service, err = NewDeepSeekService(config)
	assert.NotNil(t, service)
	assert.Nil(t, err)
}

func TestDeepSeekService_GetCompletion(t *testing.T) {
	config := &DeepSeekConfig{ApiKey: "test_api_key"}
	service, err := NewDeepSeekService(config)
	assert.NotNil(t, service)
	assert.Nil(t, err)

	ctx := context.Background()
	prompt := "Hello, world!"
	systemMessage := "You are a helpful assistant."
	model := "text-davinci-003"
	temperature := 0.7
	jsonModel := false

	// Mock the response from DeepSeek API
	service.client = &openai.Client{
		// Mock the CreateChatCompletion method
		CreateChatCompletionFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			return openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Hello, user!",
						},
					},
				},
			}, nil
		},
	}

	response, err := service.GetCompletion(prompt, systemMessage, model, temperature, jsonModel, ctx)
	assert.Nil(t, err)
	assert.Equal(t, "Hello, user!", response)
}
