package cmd

import (
	"context"
	"io"

	tele "gopkg.in/telebot.v3"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
	"tg_ai_service/service/ai_app"
)

type Voice2TextCmd struct {
	ai *ai_app.VoiceToTextApp
}

func NewVoice2TextCmd(voiceToTextApp *ai_app.VoiceToTextApp) (*Voice2TextCmd, error) {
	if voiceToTextApp == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "voice to text app is empty")
	}

	return &Voice2TextCmd{
		ai: voiceToTextApp,
	}, nil
}

func (vc *Voice2TextCmd) Name() common.CmdName {
	return common.VOICE_TO_TEXT
}

func (vc *Voice2TextCmd) Execute(ctx tele.Context) {
	voice := ctx.Message().Voice
	if voice == nil {
		log.Errorf("this message doesn't have voice")
		_ = ctx.Reply("this message doesn't have voice")
		return
	}

	log.Infof("this voice duration:%ds", voice.Duration)
	var err error = nil
	var result string = ""

	defer func() {
		if err != nil {
			log.Errorf("get transcription error:%s", err.Error())
			_ = ctx.Reply(err.Error())
			return
		}

		err = ctx.Reply(result)
		if err != nil {
			log.Errorf("reply is err:%s", err.Error())
			return
		}
	}()

	var input io.ReadCloser = nil
	input, err = ctx.Bot().File(voice.MediaFile())
	if err != nil {
		log.Errorf("download voice error:%s", err.Error())
		_ = ctx.Reply(err.Error())
		return
	}
	defer func() {
		_ = input.Close()
	}()

	//if vc.transCode {
	//	outputFilename, tmpErr := vc.transcode()
	//	if tmpErr != nil {
	//		err = tmpErr
	//		return
	//	}
	//
	//	input, tmpErr = os.Open(outputFilename)
	//	if tmpErr != nil {
	//		log.Errorf("open output file error:%s", err.Error())
	//		err = tmpErr
	//		return
	//	}
	//	defer func() {
	//		_ = os.Remove(outputFilename)
	//	}()
	//}

	ctxB := context.Background()
	result, err = vc.ai.GetTranscription("", input, "", 0, common.TEXT, ctxB)
	if err != nil {
		log.Errorf("get transcription error:%s", err.Error())
		return
	}
}

//func (vc *Voice2TextCmd) transcode() (string, error) {
//	err := vc.teleContext.Bot().Download(vc.voice.MediaFile(), common.TMP_AUDIO_PATH+vc.voice.FileID)
//	if err != nil {
//		log.Errorf("download voice error:%s", err.Error())
//		return "", err
//	}
//	defer func() {
//		_ = os.Remove(common.TMP_AUDIO_PATH + vc.voice.FileID)
//	}()
//
//	err = common.OggToMp3(common.TMP_AUDIO_PATH+vc.voice.FileID, common.TMP_AUDIO_PATH+vc.voice.FileID+".mp3")
//	if err != nil {
//		log.Errorf("convert voice error:%s", err.Error())
//		return "", err
//	}
//
//	inputExt, _ := common.GetFileType(common.TMP_AUDIO_PATH + vc.voice.FileID)
//	outputExt, _ := common.GetFileType(common.TMP_AUDIO_PATH + vc.voice.FileID + ".mp3")
//	log.Infof("inputExt:%s, outputExt:%s", inputExt, outputExt)
//	return common.TMP_AUDIO_PATH + vc.voice.FileID + ".mp3", nil
//}
