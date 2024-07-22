package ai

import (
	gemini "github.com/google/generative-ai-go"

	"tg_ai_service/internal/common"
)

type GeminiConfig struct {
	ApiKey    string `json:"api_key"`
	ApiSecret string `json:"api_secret"`
}

func (rd *GeminiConfig) Check() *common.InnerError {
	if rd.ApiKey == "" || rd.ApiSecret == "" {
		return &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "api key or api secret is empty",
			Code:    0,
		}
	}
	return nil
}

type GeminiAiService struct {
	config *GeminiConfig
	client *gemini.Client
}

func (s *GeminiAiService) GetCompletion(prompt string) (string, error) {
	resp, err := s.client.Complete(prompt)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}
