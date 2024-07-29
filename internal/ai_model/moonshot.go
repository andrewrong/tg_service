package ai_model

import (
	"context"
	"io"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type MoonShotConfig struct {
	ApiKey  string `json:"api_key"`
	BaseUrl string `json:"base_url"`
}

func (rd *MoonShotConfig) Check() error {
	if rd.ApiKey == "" {
		return common.NewInnerErrorWithoutCode(common.ParameterError, "api key is empty")
	}

	if rd.BaseUrl == "" {
		rd.BaseUrl = "https://api.moonshot.cn/v1"
	}
	return nil
}

type MoonShotService struct {
	config *MoonShotConfig
	client *CompatibleOpenAIService
	models []string
}

func NewMoonShotService(config *MoonShotConfig) (*MoonShotService, error) {
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

	tmpC, err := NewCompatibleOpenAIService(&CompatibleOpenAIConfig{
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
	return "", &common.InnerError{
		ErrType: common.ParameterError,
		ErrMsg:  "moonshot not support transcription",
		Code:    0,
	}
}

func (s *MoonShotService) GetSupportModels() []string {
	return s.models
}
