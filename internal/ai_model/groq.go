package ai_model

import (
	"context"
	"io"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type GroqConfig struct {
	ApiKey  string `json:"api_key"`
	BaseUrl string `json:"base_url"`
}

func (rd *GroqConfig) Check() error {
	if rd.ApiKey == "" {
		return common.NewInnerErrorWithoutCode(common.ParameterError, "api key is empty")
	}

	if rd.BaseUrl == "" {
		rd.BaseUrl = "https://api.groq.com/openai/v1"
	}
	return nil
}

type GroqService struct {
	config *GroqConfig
	client *CompatibleOpenAIService
	models []string
}

func NewGroqService(config *GroqConfig) (*GroqService, error) {
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
		AiTy:    common.Groq,
	})
	if err != nil {
		return nil, err
	}

	models, err := tmpC.GetSupportModels(context.Background())
	if err != nil {
		return nil, err
	}

	if len(models) == 0 {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "no support model",
			Code:    0,
		}
	}

	log.Infof("[%s] init success", tmpC.GetType())
	return &GroqService{
		config: config,
		client: tmpC,
		models: models,
	}, nil
}

func (s *GroqService) GetCompletion(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, error) {
	return s.client.GetCompletion(prompt, systemMessage, model, temperature, jsonModel, ctx)
}

func (s *GroqService) GetTranscription(prompt string, reader io.Reader, mode string, temperature float32, format common.TranscriptionFormat, ctx context.Context) (string, error) {
	if prompt == "" {
		prompt = "生于忧患，死于欢乐。不亦快哉！"
	}

	if mode == "" {
		mode = "whisper-large-v3"
	}
	return s.client.GetTranscription(prompt, reader, mode, temperature, format, ctx)
}

func (s *GroqService) GetSupportModels() []string {
	return s.models
}
