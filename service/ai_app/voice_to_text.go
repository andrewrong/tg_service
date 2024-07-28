package ai_app

import (
	"context"
	"io"

	"tg_ai_service/internal/common"
)

type VoiceToTextApp struct {
	ai           common.AI
	defaultModel string
}

func NewVoiceToTextApp(ai common.AI, defaultModel string) *VoiceToTextApp {
	if defaultModel == "" {
		defaultModel = "whisper-1"
	}
	return &VoiceToTextApp{
		ai:           ai,
		defaultModel: defaultModel,
	}
}

func (v *VoiceToTextApp) GetTranscription(prompt string, reader io.Reader, mode string, temperature float32, format common.TranscriptionFormat, ctx context.Context) (string, error) {
	return v.ai.GetTranscription(prompt, reader, mode, temperature, format, ctx)
}
