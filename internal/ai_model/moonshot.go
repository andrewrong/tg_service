package ai_model

import (
	"context"
	"io"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type MoonShotService struct {
	config *common.MoonShotConfig
	client *CompatibleOpenAIService
	models []string
}

func NewMoonShotService(config *common.MoonShotConfig) (*MoonShotService, error) {
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
		AiTy:    common.MoonShot,
	})
	if err != nil {
		return nil, err
	}

	models, err := tmpC.GetSupportModels(context.Background())
	if err != nil {
		return nil, err
	}
	log.Infof("[%s] init success", tmpC.GetType())
	return &MoonShotService{
		config: config,
		client: tmpC,
		models: models,
	}, nil
}

func (s *MoonShotService) GetCompletion(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, error) {
	return s.client.GetCompletion(prompt, systemMessage, model, temperature, jsonModel, ctx)
}

func (s *MoonShotService) GetTranscription(prompt string, reader io.Reader, mode string, temperature float32, format common.TranscriptionFormat, ctx context.Context) (string, error) {
	return "", common.NewInnerErrorWithoutCode(common.ParameterError, "moonshot not support transcription")
}

func (s *MoonShotService) GetSupportModels() []string {
	return s.models
}

func (s *MoonShotService) GetDefaultModel() string {
	return "moonshot-v1-8k"
}
