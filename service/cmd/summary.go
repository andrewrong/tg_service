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

type SummaryCommand struct {
	ai          *ai_app.ArticleSummaryApp
	url2md      *service.Url2MdService
	url         string
	teleContext tele.Context
	model       string
}

func NewSummaryCommand(articleSummaryApp *ai_app.ArticleSummaryApp, url2md *service.Url2MdService, teleContext tele.Context) (*SummaryCommand, error) {
	if articleSummaryApp == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "article summary app is empty")
	}

	if url2md == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "url2md service is empty")
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
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "url is empty")
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
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "invalid url")
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
