package ai_model

import (
	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

func ValidAiMode(ai common.AI) bool {
	switch ai.(type) {
	case *OpenAIService:
	case *GroqService:
	case *DeepSeekService:
	case *MoonShotService:
		{
			return true
		}
	default:
		return false
	}
	return false
}

func GetTextDefaultModel(ai common.AI) string {
	switch ai.(type) {
	case *OpenAIService:
		return "gpt-4o-mini"
	case *GroqService:
		return "llama-3.1-8b-instant"
	case *DeepSeekService:
		return "deepseek-chat"
	case *MoonShotService:
		return "moonshot-v1-32k"
	}
	return ""
}

func GetVoiceDefaultModel(ai common.AI) string {
	switch ai.(type) {
	case *OpenAIService:
		return "whisper-1"
	case *GroqService:
		return "whisper-large-v3"
	}
	return ""
}

type AiService struct {
	config   *common.AiConfig
	openai   *OpenAIService
	deep     *DeepSeekService
	groq     *GroqService
	moonshot *MoonShotService
}

func NewAiService(config *common.AiConfig) (*AiService, error) {
	if config == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "config is empty")
	}

	err := config.Check()
	if err != nil {
		return nil, err
	}
	s := &AiService{
		config: config,
	}

	if config.OpenAiCfg != nil {
		tmp, err := NewOpenAIService(config.OpenAiCfg)
		if err != nil {
			return nil, err
		}
		s.openai = tmp
	}

	if config.DeepseekCfg != nil {
		tmp, err := NewDeepSeekService(config.DeepseekCfg)
		if err != nil {
			return nil, err
		}
		s.deep = tmp
	}

	if config.GroqAiCfg != nil {
		tmp, err := NewGroqService(config.GroqAiCfg)
		if err != nil {
			return nil, err
		}
		s.groq = tmp
	}

	if config.MoonshotCfg != nil {
		tmp, err := NewMoonShotService(config.MoonshotCfg)
		if err != nil {
			return nil, err
		}
		s.moonshot = tmp
	}

	return s, nil
}

func (ai *AiService) GetAiByType(aiType common.AiType) common.AI {
	switch aiType {
	case common.OpenAI:
		return ai.openai
	case common.DeepSeek:
		return ai.deep
	case common.Groq:
		return ai.groq
	case common.MoonShot:
		return ai.moonshot
	default:
		{
			log.Infof("don't support %s ai", aiType)
			return nil
		}
	}
}
