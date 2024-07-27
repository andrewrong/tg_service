package service

import (
	"context"
	"net/url"

	tele "gopkg.in/telebot.v3"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
	"tg_ai_service/service/ai_app"
)

const (
	DEFAULT_SOURCE_LANG = "English"
	DEFAULT_TARGET_LANG = "Chinese"
)

// 翻译的命令, /ts url|text sourcelanguage targetlanguage (内部识别是text还是url)
type TSCommand struct {
	translateApp *ai_app.TranslateApp
	url2md       *Url2MdService
	text         string
	isUrl        bool
	teleContext  tele.Context

	// 源语言
	SourceLanguage string
	// 目标语言
	TargetLanguage string
}

func NewTSCommand(translateApp *ai_app.TranslateApp, url2md *Url2MdService, teleContext tele.Context) (*TSCommand, error) {
	if translateApp == nil {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "translate app is empty",
			Code:    0,
		}
	}

	if url2md == nil {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "url2md service is empty",
			Code:    0,
		}
	}

	tsCommand := &TSCommand{
		translateApp: translateApp,
		teleContext:  teleContext,
		url2md:       url2md,
	}

	tags := teleContext.Args()
	if len(tags) == 0 {
		log.Errorf("ts param is empty")
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "text is empty",
			Code:    0,
		}
	}

	text := tags[0]
	if text == "" {
		log.Errorf("text is empty")
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "text is empty",
			Code:    0,
		}
	}

	if _, err := url.ParseRequestURI(text); err == nil {
		tsCommand.isUrl = true
	}
	tsCommand.text = text
	tsCommand.SourceLanguage = DEFAULT_SOURCE_LANG
	tsCommand.TargetLanguage = DEFAULT_TARGET_LANG

	if len(tags) == 2 {
		source := tags[1]
		if source == "" {
			source = DEFAULT_SOURCE_LANG
		}
		tsCommand.SourceLanguage = source
	} else if len(tags) > 2 {
		source := tags[1]
		target := tags[2]
		if source == "" {
			source = DEFAULT_SOURCE_LANG
		}
		if target == "" {
			target = DEFAULT_TARGET_LANG
		}
		tsCommand.SourceLanguage = source
		tsCommand.TargetLanguage = target
	}

	return tsCommand, nil
}

func (ts *TSCommand) Execute() {
	var err error = nil
	var result string = ""

	defer func() {
		if err != nil {
			_ = ts.teleContext.Reply(err.Error())
			return
		}

		chunkMsgs := common.SplitBySize(result, 4000)

		for i, msg := range chunkMsgs {
			err = ts.teleContext.Reply(msg, &tele.SendOptions{
				ParseMode: tele.ModeMarkdown,
			})
			if err != nil {
				log.Errorf("[%d] Send tg is error:%s", i, err.Error())
				log.Infof("%s", msg)
			}
		}

	}()

	if ts.isUrl {
		content, tmpE := ts.url2md.GetUrl2Md(ts.text)
		if tmpE != nil {
			err = tmpE
			return
		}
		ts.text = content.Content
	}

	ctx := context.Background()
	result, err = ts.translateApp.Translate(ts.SourceLanguage, ts.TargetLanguage, ts.text, "", 2000, ctx)
	if err != nil {
		return
	}
}

type SummaryCommand struct {
	ai          *ai_app.ArticleSummaryApp
	url2md      *Url2MdService
	url         string
	teleContext tele.Context
	model       string
}

func NewSummaryCommand(articleSummaryApp *ai_app.ArticleSummaryApp, url2md *Url2MdService, teleContext tele.Context) (*SummaryCommand, error) {
	if articleSummaryApp == nil {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "article summary app is empty",
			Code:    0,
		}
	}

	if url2md == nil {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "url2md service is empty",
			Code:    0,
		}
	}

	summaryCommand := &SummaryCommand{
		ai:          articleSummaryApp,
		teleContext: teleContext,
		url2md:      url2md,
		model:       "gpt-4o-mini",
	}

	tags := teleContext.Args()
	if len(tags) == 0 {
		log.Errorf("summary param is empty")
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "url is empty",
			Code:    0,
		}
	}

	text := tags[0]
	if text == "" {
		log.Errorf("url is empty")
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "url is empty",
			Code:    0,
		}
	}

	if _, err := url.ParseRequestURI(text); err != nil {
		log.Errorf("invalid url: %s", text)
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "invalid url",
			Code:    0,
		}
	}
	summaryCommand.url = text
	if len(tags) > 1 {
		model := tags[1]
		if model != "" {
			summaryCommand.model = model
		}
	}

	return summaryCommand, nil
}

func (sc *SummaryCommand) Execute() {
	var err error = nil
	var result string = ""

	defer func() {
		if err != nil {
			_ = sc.teleContext.Reply(err.Error())
			return
		}

		chunkMsgs := common.SplitBySize(result, 4000)
		for i, msg := range chunkMsgs {
			err = sc.teleContext.Reply(msg, &tele.SendOptions{
				ParseMode: tele.ModeMarkdown,
			})
			if err != nil {
				log.Errorf("[%d] Send tg is error:%s", i, err.Error())
				log.Infof("%s", msg)
			}
		}

	}()

	content, tmpE := sc.url2md.GetUrl2Md(sc.url)
	if tmpE != nil {
		err = tmpE
		return
	}

	ctx := context.Background()
	result, err = sc.ai.GetSummary(content.Content, sc.model, ctx)
	if err != nil {
		return
	}
}

type Voice2TextCmd struct {
	ai          *ai_app.VoiceToTextApp
	teleContext tele.Context
	voice       *tele.Voice
}

func NewVoice2TextCmd(voiceToTextApp *ai_app.VoiceToTextApp, teleContext tele.Context) (*Voice2TextCmd, error) {
	if voiceToTextApp == nil {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "voice to text app is empty",
			Code:    0,
		}
	}

	if teleContext.Message().Voice == nil {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "voice is empty",
			Code:    0,
		}
	}

	return &Voice2TextCmd{
		ai:          voiceToTextApp,
		teleContext: teleContext,
		voice:       teleContext.Message().Voice,
	}, nil
}

func (vc *Voice2TextCmd) Execute() {
	log.Infof("this voice duration:%ds", vc.voice.Duration)
	input := &common.OggReadCloser{
		Reader: vc.voice.MediaFile().FileReader,
	}

	output := common.NewMemoryWriteSeeker()
	err := common.OggToWav(input, output)
	if err != nil {
		_ = vc.teleContext.Reply(err.Error())
		return
	}

	ctx := context.Background()
	result, err := vc.ai.GetTranscription("", output, "", 0, common.TEXT, ctx)
	if err != nil {
		log.Errorf("get transcription error:%s", err.Error())
		return
	}

	log.Infof("result:%s", result)

	_ = vc.teleContext.Reply(result)
}
