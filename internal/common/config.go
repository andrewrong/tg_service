package common

import "tg_ai_service/internal/log"

type Url2MdConfig struct {
	BearerToken string `mapstructure:"bearer_token"` // bear token
	TimeoutMs   int    `mapstructure:"timeout_ms"`   // 超时时间
	ServiceUrl  string `mapstructure:"service_url"`  // service url
}

func (rd *Url2MdConfig) Check() error {
	if rd.BearerToken == "" {
		return NewInnerErrorWithoutCode(ParameterError, "bear token is empty")
	}
	if rd.ServiceUrl == "" {
		return NewInnerErrorWithoutCode(ParameterError, "service url is empty")
	}
	if rd.TimeoutMs <= 0 {
		rd.TimeoutMs = 60 * 1000
	}
	return nil
}

type OpenAiConfig struct {
	ApiKey string `mapstructure:"api_key"`
}

func (rd *OpenAiConfig) Check() error {
	if rd.ApiKey == "" {
		return NewInnerErrorWithoutCode(ParameterError, "api key is empty")
	}
	return nil
}

type CompatibleOpenAIConfig struct {
	ApiKey  string `mapstructure:"api_key"`
	BaseUrl string `mapstructure:"base_url"`
	AiTy    AiType `mapstructure:"ai_ty"`
}

func (rd *CompatibleOpenAIConfig) Check() error {
	if rd.ApiKey == "" {
		return NewInnerErrorWithoutCode(ParameterError, "api key is empty")
	}

	if err := rd.AiTy.Check(); err != nil {
		return err
	}

	if rd.BaseUrl == "" && rd.AiTy != OpenAI {
		return NewInnerErrorWithoutCode(ParameterError, "base url is empty")
	}
	return nil
}

type DeepSeekConfig struct {
	ApiKey  string `mapstructure:"api_key"`
	BaseUrl string `mapstructure:"base_url"`
}

func (rd *DeepSeekConfig) Check() error {
	if rd.ApiKey == "" {
		return NewInnerErrorWithoutCode(ParameterError, "api key is empty")
	}

	if rd.BaseUrl == "" {
		rd.BaseUrl = "https://api.deepseek.com/v1"
	}
	return nil
}

type GroqConfig struct {
	ApiKey  string `mapstructure:"api_key"`
	BaseUrl string `mapstructure:"base_url"`
}

func (rd *GroqConfig) Check() error {
	if rd.ApiKey == "" {
		return NewInnerErrorWithoutCode(ParameterError, "api key is empty")
	}

	if rd.BaseUrl == "" {
		rd.BaseUrl = "https://api.groq.com/openai/v1"
	}
	return nil
}

type MoonShotConfig struct {
	ApiKey  string `mapstructure:"api_key"`
	BaseUrl string `mapstructure:"base_url"`
}

func (rd *MoonShotConfig) Check() error {
	if rd.ApiKey == "" {
		return NewInnerErrorWithoutCode(ParameterError, "api key is empty")
	}

	if rd.BaseUrl == "" {
		rd.BaseUrl = "https://api.moonshot.cn/v1"
	}
	return nil
}

type AiConfig struct {
	OpenAiCfg   *OpenAiConfig   `mapstructure:"open_ai_cfg"`
	GroqAiCfg   *GroqConfig     `mapstructure:"groq_ai_cfg"`
	DeepseekCfg *DeepSeekConfig `mapstructure:"deepseek_cfg"`
	MoonshotCfg *MoonShotConfig `mapstructure:"moonshot_cfg"`
}

func (ai *AiConfig) Check() error {
	if ai.OpenAiCfg != nil {
		err := ai.OpenAiCfg.Check()
		if err != nil {
			log.Errorf("open cfg is err:%s", err.Error())
			return err
		}
	}

	if ai.DeepseekCfg != nil {
		err := ai.DeepseekCfg.Check()
		if err != nil {
			log.Errorf("deepseek cfg is err:%s", err.Error())
			return err
		}
	}

	if ai.GroqAiCfg != nil {
		err := ai.GroqAiCfg.Check()
		if err != nil {
			log.Errorf("groq cfg is err:%s", err.Error())
			return err
		}
	}

	if ai.MoonshotCfg != nil {
		err := ai.MoonshotCfg.Check()
		if err != nil {
			log.Errorf("moonshot cfg is err:%s", err.Error())
			return err
		}
	}

	if ai.OpenAiCfg == nil && ai.DeepseekCfg == nil && ai.GroqAiCfg == nil && ai.MoonshotCfg == nil {
		return NewInnerErrorWithoutCode(ParameterError, "ai config is empty")
	}

	return nil
}

type TranslateAppConfig struct {
	AiT             AiType `mapstructure:"ai_type"`
	Model           string `mapstructure:"model"`
	ChunksInContext int    `mapstructure:"chunks_in_context"`
	MaxToken        int    `mapstructure:"max_token"`
}

func (t *TranslateAppConfig) Check() error {
	if err := t.AiT.Check(); err != nil {
		return err
	}

	if t.Model == "" {
		log.Errorf("translateAppConfig model is empty")
		return NewInnerErrorWithoutCode(ParameterError, "translateAppConfig model is empty")
	}

	if t.ChunksInContext < 2 {
		t.ChunksInContext = 2
	}

	if t.MaxToken <= 0 {
		t.MaxToken = 500
	}

	return nil
}

type ArticleSummaryAppConfig struct {
	AiT   AiType `mapstructure:"ai_type"`
	Model string `mapstructure:"model"`
}

func (t *ArticleSummaryAppConfig) Check() error {
	if err := t.AiT.Check(); err != nil {
		return err
	}

	if t.Model == "" {
		log.Errorf("ArticleSummaryAppConfig model is empty")
		return NewInnerErrorWithoutCode(ParameterError, "ArticleSummaryAppConfig model is empty")
	}

	return nil
}

type VoiceToTextAppConfig struct {
	VoiceAiType  AiType `mapstructure:"voice_ai_type"`
	ImportAiType AiType `mapstructure:"import_ai_type"`
	Model        string `mapstructure:"voice_ai_model"`
	ImportModel  string `mapstructure:"import_ai_model"`
}

func (t *VoiceToTextAppConfig) Check() error {
	if err := t.VoiceAiType.Check(); err != nil {
		log.Errorf("voice ai type is err:%s", err.Error())
		return err
	}

	if err := t.ImportAiType.Check(); err != nil {
		log.Errorf("import ai type is err:%s", err.Error())
		return err
	}

	if t.Model == "" {
		log.Errorf("ArticleSummaryAppConfig model is empty")
		return NewInnerErrorWithoutCode(ParameterError, "ArticleSummaryAppConfig model is empty")
	}

	if t.ImportModel == "" {
		log.Errorf("ArticleSummaryAppConfig model is empty")
		return NewInnerErrorWithoutCode(ParameterError, "ArticleSummaryAppConfig model is empty")
	}

	return nil
}

type AiAppConfig struct {
	TranslateCfg  *TranslateAppConfig      `mapstructure:"translate_cfg"`
	SummaryCfg    *ArticleSummaryAppConfig `mapstructure:"summary_cfg"`
	Voice2TextCfg *VoiceToTextAppConfig    `mapstructure:"voice_2_text_cfg"`
}

func (ai *AiAppConfig) Check() error {
	if ai.TranslateCfg != nil {
		if err := ai.TranslateCfg.Check(); err != nil {
			return err
		}
	}

	if ai.SummaryCfg != nil {
		if err := ai.SummaryCfg.Check(); err != nil {
			return err
		}
	}

	if ai.Voice2TextCfg != nil {
		if err := ai.Voice2TextCfg.Check(); err != nil {
			return err
		}
	}

	return nil
}

type FileReaderConfig struct {
	TimeoutMs  int32  `mapstructure:"timeout_ms"` // 超时时间
	PdfBaseUrl string `mapstructure:"base_url"`   //jinra的base url
}

func (rd *FileReaderConfig) Check() error {
	if rd.PdfBaseUrl == "" {
		rd.PdfBaseUrl = "https://r.jina.ai/"
	}

	if rd.TimeoutMs <= 0 {
		rd.TimeoutMs = 30 * 1000
	}
	return nil
}

type ExternalConfig struct {
	Url2MdCfg     *Url2MdConfig     `mapstructure:"url2md_cfg"`
	FileReaderCfg *FileReaderConfig `mapstructure:"file_reader_cfg"`
}

func (rd *ExternalConfig) Check() error {
	if rd.Url2MdCfg != nil {
		if err := rd.Url2MdCfg.Check(); err != nil {
			return err
		}
	} else {
		return NewInnerErrorWithoutCode(ParameterError, "url2md config is empty")
	}
	if rd.FileReaderCfg != nil {
		if err := rd.FileReaderCfg.Check(); err != nil {
			return err
		}
	} else {
		return NewInnerErrorWithoutCode(ParameterError, "file reader config is empty")
	}
	return nil
}

type TgAiConfig struct {
	TgToken     string         `mapstructure:"tg_token"`
	AiModels    AiConfig       `mapstructure:"ai_models"`
	AiAppCfg    AiAppConfig    `mapstructure:"ai_app_cfg"`
	ExternalCfg ExternalConfig `mapstructure:"external_cfg"`
}

func (rd *TgAiConfig) Check() error {
	if rd.TgToken == "" {
		return NewInnerErrorWithoutCode(ParameterError, "tg token is empty")
	}
	if err := rd.AiModels.Check(); err != nil {
		return err
	}
	if err := rd.AiAppCfg.Check(); err != nil {
		return err
	}
	if err := rd.ExternalCfg.Check(); err != nil {
		return err
	}
	return nil
}
