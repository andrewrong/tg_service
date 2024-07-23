package ai_model

import (
	"context"

	"github.com/sashabaranov/go-openai"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type DeepSeekConfig struct {
	ApiKey  string `json:"api_key"`
	BaseUrl string `json:"base_url"`
}

func (rd *DeepSeekConfig) Check() error {
	if rd.ApiKey == "" {
		return &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "api key is empty",
			Code:    0,
		}
	}

	if rd.BaseUrl == "" {
		rd.BaseUrl = "https://api.deepseek.com/v1"
	}
	return nil
}

type DeepSeekService struct {
	config *DeepSeekConfig
	client *openai.Client
}

func NewDeepSeekService(config *DeepSeekConfig) (*DeepSeekService, error) {
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

	tmpClientCfg := openai.DefaultConfig(config.ApiKey)
	tmpClientCfg.BaseURL = config.BaseUrl
	client := openai.NewClientWithConfig(tmpClientCfg)

	return &DeepSeekService{
		config: config,
		client: client,
	}, nil
}

func (s *DeepSeekService) GetCompletion(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, error) {
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
