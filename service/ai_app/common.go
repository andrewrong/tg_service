package ai_app

import (
	"tg_ai_service/internal/ai_model"
	"tg_ai_service/internal/common"
)

type AiAppService struct {
	translateApp   *TranslateApp
	voiceToTextApp *VoiceToTextApp
	summaryApp     *ArticleSummaryApp
}

func NewAiAppService(cfg *common.AiAppConfig, aiS *ai_model.AiService) (*AiAppService, error) {
	app := &AiAppService{}

	if cfg.TranslateCfg != nil {
		tmp, err := NewTranslateApp(cfg.TranslateCfg, aiS)
		if err != nil {
			return nil, err
		}
		app.translateApp = tmp
	}

	if cfg.Voice2TextCfg != nil {
		tmp, err := NewVoiceToTextApp(cfg.Voice2TextCfg, aiS)
		if err != nil {
			return nil, err
		}
		app.voiceToTextApp = tmp
	}

	if cfg.SummaryCfg != nil {
		tmp, err := NewArticleSummaryApp(cfg.SummaryCfg, aiS)
		if err != nil {
			return nil, err
		}
		app.summaryApp = tmp
	}

	return app, nil
}

func (s *AiAppService) GetTranslateApp() (*TranslateApp, error) {
	if s.translateApp == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "translate app is not supported")
	}
	return s.translateApp, nil
}

func (s *AiAppService) GetVoiceToTextApp() (*VoiceToTextApp, error) {
	if s.voiceToTextApp == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "voice to text app is not supported")
	}
	return s.voiceToTextApp, nil
}

func (s *AiAppService) GetSummaryApp() (*ArticleSummaryApp, error) {
	if s.summaryApp == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "article summary app is not supported")
	}
	return s.summaryApp, nil
}
