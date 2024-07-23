package service

import (
	tele "gopkg.in/telebot.v3"

	"tg_ai_service/service/ai_app"
)

// 翻译的命令, /ts url|text (内部识别是text还是url)
type TSCommand struct {
	translateApp *ai_app.TranslateApp
	text         string
	isUrl        bool
	teleContext  tele.Context
}

func NewTSCommand(translateApp *ai_app.TranslateApp, text string, teleContext tele.Context) *TSCommand {
	
	return &TSCommand{
		translateApp: translateApp,
		text:         text,
		teleContext:  teleContext,
	}
}
