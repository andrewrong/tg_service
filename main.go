package main

import (
	"os"
	"time"

	"github.com/joho/godotenv"

	tele "gopkg.in/telebot.v3"

	"tg_ai_service/internal/ai_model"
	"tg_ai_service/internal/log"
	"tg_ai_service/service"
	"tg_ai_service/service/ai_app"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// 关于tg_bot的使用
	pref := tele.Settings{
		Token:  os.Getenv("TG_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// 启动一些服务做测试
	url2md, _ := service.NewUrl2MdService(&service.Url2MdConfig{
		BearerToken: os.Getenv("URL2MD_TOKEN"),
		ServiceUrl:  os.Getenv("URL2MD_SERVICE_URL"),
	})
	aiS, _ := service.NewAiService(&service.AiConfig{
		DeepseekConfig: &ai_model.DeepSeekConfig{
			ApiKey: os.Getenv("DEEPSEEK_TOKEN"),
		},
		OpenaiConfig: nil,
	})

	translateS := ai_app.NewTranslateApp(aiS.GetDeepSeekService(), "deepseek-chat", 8)
	b.Handle("/ts", func(c tele.Context) error {
		execute, err := service.NewTSCommand(translateS, url2md, c)
		if err != nil {
			c.Send(err.Error())
			return err
		}
		execute.Execute()
		return nil
	})

	b.Start()
}
