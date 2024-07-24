package common

import (
	"context"
	"net/url"
)

type Checker interface {
	// Check checks the internal state of the object and returns an error if any.
	//
	// It returns a pointer to an InnerError object, which represents an error that
	// occurred during the check. If the check was successful, it returns nil.
	//
	// Example:
	//   err := obj.Check()
	//   if err != nil {
	//       // handle the error
	//   }
	Check() error
}

type AI interface {
	GetCompletion(prompt, systemMessage, mode string, temperature float32, jsonModel bool, ctx context.Context) (string, int, error)
}

// isValidURL 解析并验证 URL 的合法性
func IsValidURL(u string) (*url.URL, bool) {
	if u == "" {
		return nil, false
	}
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return nil, false
	}
	return parsedURL, true
}

type AiType string

const (
	OpenAI   AiType = "openai"
	DeepSeek AiType = "deepseek"
	MoonShot AiType = "moonshot"
	Groq     AiType = "groq"
)

var AiTypes = []AiType{OpenAI, DeepSeek, MoonShot, Groq}

type OneAiRecord struct {
	Cost        int64 `json:"cost"` // 耗时
	InputToken  int   `json:"input_token"`
	OutputToken int   `json:"output_token"`
}

type TranslateStepRecord struct {
	StepName string         `json:"sname"`    //本次Step的名字
	SumCost  int64          `json:"sum_cost"` //本次Step的耗时
	Records  []*OneAiRecord `json:"records"`
}

type TranslateRecord struct {
	SourceLang  string                 `json:"source_lang"`
	TargetLang  string                 `json:"target_lang"`
	MaxTokens   int                    `json:"max_token"`    // 最大token
	InputTokens int                    `json:"input_tokens"` // 输入token个数
	ChunksCount int                    `json:"chunks_count"` // 分块个数
	Steps       []*TranslateStepRecord `json:"steps"`
}

func (t *TranslateRecord) String() string {

}
