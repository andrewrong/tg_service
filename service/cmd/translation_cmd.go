package cmd

import (
	"context"
	"net/url"

	tele "gopkg.in/telebot.v3"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
	"tg_ai_service/service"
	"tg_ai_service/service/ai_app"
)

const (
	DEFAULT_SOURCE_LANG = "English"
	DEFAULT_TARGET_LANG = "Chinese"
)

type TranslationParams struct {
	IsUrl bool
	Text  string
	// 源语言
	SourceLanguage string
	// 目标语言
	TargetLanguage string
}

// 翻译的命令, /ts url|text sourcelanguage targetlanguage (内部识别是text还是url)
type TSCommand struct {
	translateApp *ai_app.TranslateApp
	url2md       *service.Url2MdService
}

func NewTSCommand(translateApp *ai_app.TranslateApp, url2md *service.Url2MdService) (*TSCommand, error) {
	if translateApp == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "translate app is empty")
	}

	if url2md == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "url2md service is empty")
	}

	tsCommand := &TSCommand{
		translateApp: translateApp,
		url2md:       url2md,
	}

	return tsCommand, nil
}

func (ts *TSCommand) Name() common.CmdName {
	return common.TS
}

func (ts *TSCommand) parseArgs(tCtx tele.Context) (*TranslationParams, error) {
	tsCommand := &TranslationParams{}

	tags := tCtx.Args()
	if len(tags) == 0 {
		tsCommand.IsUrl = false
		tsCommand.Text = tCtx.Message().Text
		tsCommand.SourceLanguage = DEFAULT_SOURCE_LANG
		tsCommand.TargetLanguage = DEFAULT_TARGET_LANG
		return tsCommand, nil
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
		tsCommand.IsUrl = true
	}
	tsCommand.Text = text
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

func (ts *TSCommand) Execute(tCtx tele.Context) {
	var err error = nil
	result := ""
	var args *TranslationParams = nil

	defer func() {
		if err != nil {
			_ = tCtx.Reply(err.Error())
			return
		}

		chunkMsgs := common.SplitBySize(result, 4000)
		for i, msg := range chunkMsgs {
			err = tCtx.Reply(msg, &tele.SendOptions{
				ParseMode: tele.ModeMarkdown,
			})
			if err != nil {
				log.Errorf("[%d] Send tg is error:%s", i, err.Error())
			}
		}
	}()

	args, err = ts.parseArgs(tCtx)
	if err != nil {
		return
	}

	if args.IsUrl {
		content, tmpE := ts.url2md.GetUrl2Md(args.Text)
		if tmpE != nil {
			err = tmpE
			return
		}
		args.Text = content.Content
	}

	ctx := context.Background()
	result, err = ts.translateApp.Translate(args.SourceLanguage, args.TargetLanguage, args.Text, "", 0, ctx)
	if err != nil {
		return
	}
}
