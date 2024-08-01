package ai_model

import (
	"context"
	"io"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type OpenAIService struct {
	config *common.OpenAiConfig
	client *CompatibleOpenAIService
	models []string
}

func NewOpenAIService(config *common.OpenAiConfig) (*OpenAIService, error) {
	if config == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "config is empty")
	}

	err := config.Check()
	if err != nil {
		return nil, err
	}

	tmpC, err := NewCompatibleOpenAIService(&common.CompatibleOpenAIConfig{
		ApiKey:  config.ApiKey,
		AiTy:    common.OpenAI,
		BaseUrl: "https://api.openai.com/v1",
	})
	if err != nil {
		return nil, err
	}

	models, err := tmpC.GetSupportModels(context.Background())
	if err != nil {
		return nil, err
	}

	if len(models) == 0 {
		log.Errorf("no support models")
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "no support models")
	}

	log.Infof("[%s] init success", tmpC.GetType())
	return &OpenAIService{
		config: config,
		client: tmpC,
		models: models,
	}, nil
}

func (s *OpenAIService) GetCompletion(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, error) {
	return s.client.GetCompletion(prompt, systemMessage, model, temperature, jsonModel, ctx)
}

func (s *OpenAIService) GetTranscription(prompt string, reader io.Reader, model string, temperature float32, format common.TranscriptionFormat, ctx context.Context) (string, error) {
	if prompt == "" {
		prompt = "生于忧患，死于欢乐。不亦快哉！"
	}

	if model == "" {
		model = "whisper-1"
	}
	return s.client.GetTranscription(prompt, reader, model, temperature, format, ctx)
}

func (s *OpenAIService) GetSupportModels() []string {
	return s.models
}

func (s *OpenAIService) GetDefaultModel() string {
	return "gpt-4o-mini"
}
