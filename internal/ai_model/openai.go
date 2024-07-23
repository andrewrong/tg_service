package ai_model

import (
	"context"

	openai "github.com/sashabaranov/go-openai"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type OpenAiConfig struct {
	ApiKey string `json:"api_key"`
}

func (rd *OpenAiConfig) Check() error {
	if rd.ApiKey == "" {
		return &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "api key is empty",
			Code:    0,
		}
	}
	return nil
}

type OpenAIService struct {
	config *OpenAiConfig
	client *openai.Client
}

func NewOpenAIService(config *OpenAiConfig) (*OpenAIService, error) {
	if config == nil {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "config is empty",
			Code:    0,
		}
	}

	err := config.Check()
	if err != nil {
		return nil, err
	}

	return &OpenAIService{
		config: config,
		client: openai.NewClient(config.ApiKey),
	}, nil
}

func (s *OpenAIService) GetCompletion(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:       model,
		Temperature: temperature,
		TopP:        1,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: systemMessage,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	if jsonModel {
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Errorf("get openai is error:%s", err)
		return "", &common.InnerError{
			ErrType: common.ExternalServiceError,
			ErrMsg:  "get openai is error",
			Code:    0,
		}
	}

	return resp.Choices[0].Message.Content, nil
}
