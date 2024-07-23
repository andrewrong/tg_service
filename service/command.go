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
			_ = ts.teleContext.Send(err.Error())
			return
		}
		_ = ts.teleContext.Send(result)
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
	result, err = ts.translateApp.Translate(ts.SourceLanguage, ts.TargetLanguage, ts.text, "", 500, ctx)
	if err != nil {
		return
	}
}
