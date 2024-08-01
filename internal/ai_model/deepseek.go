package ai_model

import (
	"context"
	"io"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type DeepSeekService struct {
	config *common.DeepSeekConfig
	client *CompatibleOpenAIService
	models []string
}

func NewDeepSeekService(config *common.DeepSeekConfig) (*DeepSeekService, error) {
	if config == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "config is empty")
	}

	err := config.Check()
	if err != nil {
		return nil, err
	}

	tmpC, err := NewCompatibleOpenAIService(&common.CompatibleOpenAIConfig{
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

func (s *DeepSeekService) GetDefaultModel() string {
	return "deepseek-chat"
}
