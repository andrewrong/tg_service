package service

import (
	"tg_ai_service/internal/ai_model"
	"tg_ai_service/internal/common"
)

type AiConfig struct {
	OpenaiConfig   *ai_model.OpenAiConfig   `json:"openai_config"`
	DeepseekConfig *ai_model.DeepSeekConfig `json:"deepseek_config"`
	GroqConfig     *ai_model.GroqConfig     `json:"groq_config"`
}

func (rd *AiConfig) Check() error {
	valid := false

	if rd.OpenaiConfig != nil {
		err := rd.OpenaiConfig.Check()
		if err != nil {
			return err
		}
		valid = true
	}

	if rd.DeepseekConfig != nil {
		err := rd.DeepseekConfig.Check()
		if err != nil {
			return err
		}
		valid = true
	}

	if rd.GroqConfig != nil {
		err := rd.GroqConfig.Check()
		if err != nil {
			return err
		}
		valid = true
	}

	if !valid {
		return common.NewInnerErrorWithoutCode(common.ParameterError, "ai config is empty")
	}

	return nil
}

type AiService struct {
	config *AiConfig
	openai *ai_model.OpenAIService
	deep   *ai_model.DeepSeekService
	groq   *ai_model.GroqService
}

func NewAiService(config *AiConfig) (*AiService, error) {
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
	s := &AiService{
		config: config,
	}

	if config.OpenaiConfig != nil {
		tmp, err := ai_model.NewOpenAIService(config.OpenaiConfig)
		if err != nil {
			return nil, err
		}
		s.openai = tmp
	}

	if config.DeepseekConfig != nil {
		tmp, err := ai_model.NewDeepSeekService(config.DeepseekConfig)
		if err != nil {
			return nil, err
		}
		s.deep = tmp
	}

	if config.GroqConfig != nil {
		tmp, err := ai_model.NewGroqService(config.GroqConfig)
		if err != nil {
			return nil, err
		}
		s.groq = tmp
	}
	return s, nil
}

func (s *AiService) GetOpenAiService() common.AI {
	return s.openai
}

func (s *AiService) GetDeepSeekService() common.AI {
	return s.deep
}

func (s *AiService) GetGroqService() common.AI {
	return s.groq
}
