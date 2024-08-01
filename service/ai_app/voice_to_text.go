package ai_app

import (
	"context"
	"io"

	"tg_ai_service/internal/ai_model"
	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type VoiceToTextApp struct {
	voiceAi     common.AI
	improveAi   common.AI
	voiceModel  string
	importModel string
}

func NewVoiceToTextApp(cfg *common.VoiceToTextAppConfig, aiS *ai_model.AiService) (*VoiceToTextApp, error) {
	voiceAi := aiS.GetAiByType(cfg.VoiceAiType)
	importAi := aiS.GetAiByType(cfg.ImportAiType)
	return &VoiceToTextApp{
		voiceAi:     voiceAi,
		improveAi:   importAi,
		voiceModel:  cfg.Model,
		importModel: cfg.ImportModel,
	}, nil
}

func (v *VoiceToTextApp) GetTranscription(prompt string, reader io.Reader, mode string, temperature float32, format common.TranscriptionFormat, ctx context.Context) (string, error) {
	voiceModel := mode
	if voiceModel == "" {
		voiceModel = v.voiceModel
	}
	result, err := v.voiceAi.GetTranscription(prompt, reader, voiceModel, temperature, format, ctx)
	if err != nil {
		return "", err
	}

	systemPrompt := "you are an expert in text proofreading and editing, capable of efficiently identifying and correcting typos, checking punctuation marks, and segmenting text. Additionally, I can make semantic modifications based on context to ensure the overall accuracy and fluency of the text."
	result, err = v.improveAi.GetCompletion(result, systemPrompt, v.importModel, 0.3, false, ctx)

	if err != nil {
		log.Errorf("improve voice text error:%s", err)
		return "", err
	}

	return result, nil
}
