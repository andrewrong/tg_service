package ai_model

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type CompatibleOpenAIConfig struct {
	ApiKey  string        `json:"api_key"`
	BaseUrl string        `json:"base_url"`
	AiTy    common.AiType `json:"ai_ty"`
}

func (rd *CompatibleOpenAIConfig) Check() error {
	if rd.ApiKey == "" {
		return &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "api key is empty",
			Code:    0,
		}
	}

	if err := rd.AiTy.Check(); err != nil {
		return err
	}

	if rd.BaseUrl == "" && rd.AiTy != common.OpenAI {
		return common.NewInnerErrorWithoutCode(common.ParameterError, "base url is empty")
	}
	return nil
}

type CompatibleOpenAIService struct {
	config *CompatibleOpenAIConfig
	client *openai.Client
}

func NewCompatibleOpenAIService(config *CompatibleOpenAIConfig) (*CompatibleOpenAIService, error) {
	if config == nil {
		log.Errorf("CompatibleOpenAIConfig is empty")
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

	return &CompatibleOpenAIService{
		config: config,
		client: client,
	}, nil
}

func (s *CompatibleOpenAIService) GetCompletion(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, error) {
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

func (s *CompatibleOpenAIService) GetType() common.AiType {
	return s.config.AiTy
}

func (s *CompatibleOpenAIService) GetCompletionStream(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, error) {
	return "", nil
}

func (s *CompatibleOpenAIService) GetTranscription(prompt string, reader io.Reader, mode string, temperature float32, format common.TranscriptionFormat, ctx context.Context) (string, error) {
	req := openai.AudioRequest{
		Model:       mode,
		Temperature: temperature,
		Prompt:      prompt,
		Format:      openai.AudioResponseFormat(format),
		Reader:      reader,
		FilePath:    "tmp.ogg", //主要用reader，这个参数只是为了不报错
	}

	resp, err := s.client.CreateTranscription(ctx, req)
	if err != nil {
		log.Errorf("[CreateTranscription] get openai is error:%s", err)
		return "", &common.InnerError{
			ErrType: common.ExternalServiceError,
			ErrMsg:  fmt.Sprintf("get openai is error:%s", err),
			Code:    0,
		}
	}
	return resp.Text, nil
}

func (s *CompatibleOpenAIService) GetSupportModels(ctx context.Context) ([]string, error) {
	models, err := s.client.ListModels(ctx)
	if err != nil {
		log.Errorf("ai get support models is error:%s", err.Error())
		return nil, err
	}
	result := make([]string, 0)
	for _, m := range models.Models {
		result = append(result, m.ID)
	}
	return result, nil
}
