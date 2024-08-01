package main

import (
	"os"
	"time"

	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
	"tg_ai_service/service"
	"tg_ai_service/service/ai_app"
	"tg_ai_service/service/cmd"
)

func main() {
	// Initialize Viper to read configuration
	viper.SetConfigName("xxx") // name of config file (without extension)
	viper.SetConfigType("json") // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")    // path to look for the config file in
	viper.AddConfigPath("/etc/tg_ai_service/") // call multiple times to add many search paths
	viper.AddConfigPath("$HOME/.tg_ai_service") // call multiple times to add many search paths

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	var config common.TgAiConfig
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	// 关于tg_bot的使用
	pref := tele.Settings{
		Token:  config.TgToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// 启动一些服务做测试
	url2md, _ := service.NewUrl2MdService(&service.Url2MdConfig{
		BearerToken: config.AiAppCfg.SummaryCfg.Model, // Assuming the model is used as BearerToken for simplicity
		ServiceUrl:  config.AiAppCfg.SummaryCfg.Model, // Assuming the model is used as ServiceUrl for simplicity
	})

	var aiS *service.AiService = nil
	{
	}

	cmdService := cmd.NewCommandService()
	{
		//初始化各种ai app
		translateS, tmpE := ai_app.NewTranslateApp(aiS.GetOpenAiService(), aiS.GetOpenAiService().GetDefaultModel(), 6)
		if tmpE != nil {
			log.Errorf("[main] new translate app error: %s", tmpE.Error())
			return
		}

		cmdService.AddCommand(translateS)
	}

	translateS := ai_app.NewTranslateApp(aiS.GetGroqService(), "llama-3.1-70b-versatile", 6)
	b.Handle("/ts", func(c tele.Context) error {
		execute, err := service.NewTSCommand(translateS, url2md, c)
		if err != nil {
			c.Send(err.Error())
			return err
		}
		execute.Execute()
		return nil
	})

	summary := ai_app.NewArticleSummaryApp(aiS.GetOpenAiService(), "")
	b.Handle("/summary", func(c tele.Context) error {
		execute, err := service.NewSummaryCommand(summary, url2md, c)
		if err != nil {
			c.Send(err.Error())
			return err
		}
		execute.Execute()
		return nil
	})

	voice2TextApp := ai_app.NewVoiceToTextApp(aiS.GetOpenAiService(), "whisper-1")
	b.Handle(tele.OnVoice, func(c tele.Context) error {
		execute, err := service.NewVoice2TextCmd(voice2TextApp, c)
		if err != nil {
			c.Send(err.Error())
			return err
		}
		execute.Execute()
		return nil
	})

	b.Start()
}
