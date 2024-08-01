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
	ai     *ai_app.ArticleSummaryApp
	url2md *service.Url2MdService
}

func NewSummaryCommand(articleSummaryApp *ai_app.ArticleSummaryApp, url2md *service.Url2MdService) (*SummaryCommand, error) {
	if articleSummaryApp == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "article summary app is empty")
	}

	if url2md == nil {
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "url2md service is empty")
	}

	summaryCommand := &SummaryCommand{
		ai:     articleSummaryApp,
		url2md: url2md,
	}

	return summaryCommand, nil
}

type SummaryParams struct {
	Url   string `json:"url"`
	Model string `json:"mode"`
}

func (sc *SummaryCommand) parseArgs(tCtx tele.Context) (*SummaryParams, error) {
	tags := tCtx.Args()
	if len(tags) == 0 {
		log.Errorf("summary param is empty")
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "url is empty")
	}

	text := tags[0]
	if text == "" {
		log.Errorf("url is empty")
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "url is empty")
	}

	if _, err := url.ParseRequestURI(text); err != nil {
		log.Errorf("invalid url: %s", text)
		return nil, common.NewInnerErrorWithoutCode(common.ParameterError, "invalid url")
	}

	params := &SummaryParams{
		Url: text,
	}

	if len(tags) > 1 {
		model := tags[1]
		if model != "" {
			params.Model = model
		}
	}
	return params, nil
}

func (sc *SummaryCommand) Name() common.CmdName {
	return common.SUMMARY
}

func (sc *SummaryCommand) Execute(tCxt tele.Context) {
	var err error = nil
	var result string = ""
	var params *SummaryParams = nil

	defer func() {
		if err != nil {
			_ = tCxt.Reply(err.Error())
			return
		}

		chunkMsgs := common.SplitBySize(result, 4000)
		for i, msg := range chunkMsgs {
			err = tCxt.Reply(msg, &tele.SendOptions{
				ParseMode: tele.ModeMarkdown,
			})
			if err != nil {
				log.Errorf("[%d] Send tg is error:%s", i, err.Error())
				log.Infof("%s", msg)
			}
		}

	}()

	params, err = sc.parseArgs(tCxt)
	if err != nil {
		return
	}

	content, tmpE := sc.url2md.GetUrl2Md(params.Url)
	if tmpE != nil {
		err = tmpE
		return
	}

	ctx := context.Background()
	result, err = sc.ai.GetSummary(content.Content, params.Model, ctx)
	if err != nil {
		return
	}
}
