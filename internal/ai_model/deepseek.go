package ai_model

import (
	"context"
	"io"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type DeepSeekConfig struct {
	ApiKey  string `json:"api_key"`
	BaseUrl string `json:"base_url"`
}

func (rd *DeepSeekConfig) Check() error {
	if rd.ApiKey == "" {
		return common.NewInnerErrorWithoutCode(common.ParameterError, "api key is empty")
	}

	if rd.BaseUrl == "" {
		rd.BaseUrl = "https://api.deepseek.com/v1"
	}
	return nil
}

type DeepSeekService struct {
	config *DeepSeekConfig
	client *CompatibleOpenAIService
	models []string
}

func NewDeepSeekService(config *DeepSeekConfig) (*DeepSeekService, error) {
	if config == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "config is empty")
	}

	err := config.Check()
	if err != nil {
		return nil, err
	}

	tmpC, err := NewCompatibleOpenAIService(&CompatibleOpenAIConfig{
		ApiKey:  config.ApiKey,
		BaseUrl: config.BaseUrl,
		AiTy:    common.DeepSeek,
	})
	if err != nil {
		return nil, err
	}

	models, err := tmpC.GetSupportModels(context.Background())
	if err != nil {
		return nil, err
	}

	if len(models) == 0 {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "no support model")
	}

	log.Infof("[%s] init success", tmpC.GetType())
	return &DeepSeekService{
		config: config,
		client: tmpC,
		models: models,
	}, nil
}

func (s *DeepSeekService) GetCompletion(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, error) {
	return s.client.GetCompletion(prompt, systemMessage, model, temperature, jsonModel, ctx)
}

func (s *DeepSeekService) GetTranscription(prompt string, reader io.Reader, mode string, temperature float32, format common.TranscriptionFormat, ctx context.Context) (string, error) {
	return "", &common.InnerError{
		ErrType: common.ParameterError,
		ErrMsg:  "deepseek not support transcription",
		Code:    0,
	}
}

func (s *DeepSeekService) GetSupportModels() []string {
	return s.models
}
