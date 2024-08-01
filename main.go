package main

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"

	"tg_ai_service/internal/ai_model"
	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
	"tg_ai_service/service"
	"tg_ai_service/service/ai_app"
	"tg_ai_service/service/cmd"
)

func main() {
	// Initialize Viper to read configuration
	viper.SetConfigName("tg_ai") // name of config file (without extension)
	viper.SetConfigType("json")  // REQUIRED if the config file does not have the extension in the name
	//viper.AddConfigPath(".")                                // path to look for the config file in
	viper.AddConfigPath("$HOME/project/self/tg_ai_service") // call multiple times to add many search paths

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	var config common.TgAiConfig
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	if err = config.Check(); err != nil {
		log.Fatalf("config error: %s", err)
		return
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

	var aiS *ai_model.AiService = nil
	var externalS *service.ExternalService = nil
	var aiApps *ai_app.AiAppService = nil

	aiS, err = ai_model.NewAiService(&config.AiModels)
	if err != nil {
		log.Fatalf("new ai service error: %s", err)
		return
	}
	externalS, err = service.NewExternalService(&config.ExternalCfg)
	if err != nil {
		log.Fatalf("new external service error: %s", err)
		return
	}
	aiApps, err = ai_app.NewAiAppService(&config.AiAppCfg, aiS)
	if err != nil {
		log.Fatalf("new ai app service error: %s", err)
		return
	}

	cmdService := cmd.NewCommandService()
	{
		tsApp, err := aiApps.GetTranslateApp()
		url2Service := externalS.GetUrl2MdService()
		if err == nil {
			tsCmd, _ := cmd.NewTSCommand(tsApp, url2Service)
			cmdService.AddCommand(tsCmd)
			log.Infof("add %s command", tsCmd.Name())
		}

		summaryApp, err := aiApps.GetSummaryApp()
		if err == nil {
			summaryCmd, _ := cmd.NewSummaryCommand(summaryApp, url2Service)
			cmdService.AddCommand(summaryCmd)
			log.Infof("add %s command", summaryCmd.Name())
		}

		voice2TextApp, err := aiApps.GetVoiceToTextApp()
		if err == nil {
			voice2TextCmd, _ := cmd.NewVoice2TextCmd(voice2TextApp)
			cmdService.AddCommand(voice2TextCmd)
			log.Infof("add %s command", voice2TextCmd.Name())
		}
	}

	CmdHandle(b, cmdService)

	b.Start()
}

func CmdHandle(b *tele.Bot, cmdService *cmd.CommandService) {
	b.Handle(string(common.TS), func(c tele.Context) error {
		execute := cmdService.GetCommand(common.TS)
		if execute == nil {
			return c.Reply(fmt.Sprintf("%s command not found", common.TS))
		}

		execute.Execute(c)
		return nil
	})

	b.Handle(string(common.SUMMARY), func(c tele.Context) error {
		execute := cmdService.GetCommand(common.SUMMARY)
		if execute == nil {
			return c.Reply(fmt.Sprintf("%s command not found", common.SUMMARY))
		}
		execute.Execute(c)
		return nil
	})

	b.Handle(tele.OnVoice, func(c tele.Context) error {
		execute := cmdService.GetCommand(common.VOICE_TO_TEXT)
		if execute == nil {
			return c.Reply(fmt.Sprintf("%s command not found", common.VOICE_TO_TEXT))
		}
		execute.Execute(c)
		return nil
	})

	b.Handle(tele.OnText, func(context tele.Context) error {
		execute := cmdService.GetCommand(common.TS)
		if execute == nil {
			return context.Reply(fmt.Sprintf("%s command not found", common.TS))
		}

		execute.Execute(context)
		return nil
	})
}
