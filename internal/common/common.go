package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/yuin/goldmark"

	"tg_ai_service/internal/log"
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
	GetCompletion(prompt, systemMessage, mode string, temperature float32, jsonModel bool, ctx context.Context) (string, error)
	GetTranscription(prompt string, reader io.Reader, mode string, temperature float32, audioFormat TranscriptionFormat, ctx context.Context) (string, error)
}

type TranscriptionFormat string

const (
	JSON  TranscriptionFormat = "json"
	TEXT  TranscriptionFormat = "text"
	SRT   TranscriptionFormat = "srt"
	VTT   TranscriptionFormat = "vtt"
	VJSON TranscriptionFormat = "verbose_json"
)

var TranscriptionFormats = []TranscriptionFormat{JSON, TEXT, SRT, VTT, VJSON}

func IsValidTranscriptionFormat(t TranscriptionFormat) bool {
	for _, v := range TranscriptionFormats {
		if v == t {
			return true
		}
	}
	return false
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
	ChunkTokens int                    `json:"chunk_tokens"` // 分块token
	Steps       []*TranslateStepRecord `json:"steps"`
}

func (t *TranslateRecord) String() string {
	var stepsInfo []string
	for _, step := range t.Steps {
		stepInfo := fmt.Sprintf("Step: %s, Cost: %dms\n", step.StepName, step.SumCost)
		for _, record := range step.Records {
			stepInfo += fmt.Sprintf("  - Cost: %dms, Input Tokens: %d, Output Tokens: %d\n", record.Cost, record.InputToken, record.OutputToken)
		}
		stepsInfo = append(stepsInfo, stepInfo)
	}
	return fmt.Sprintf("\nSource Lang: %s\nTarget Lang: %s\nMax Tokens: %d\nInput Tokens: %d\nChunks Count: %d\nSteps:\n%s",
		t.SourceLang, t.TargetLang, t.MaxTokens, t.InputTokens, t.ChunksCount, strings.Join(stepsInfo, "\n"))
}

func EscapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
		"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
		"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
		"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
	)
	return replacer.Replace(text)
}

func EscapeHtml(text string) string {
	replacer := strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	return replacer.Replace(text)
}

func Markdown2Html(text string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(text), &buf); err != nil {
		log.Errorf("Error converting markdown:%s", err)
		return "", &InnerError{
			ErrType: ExternalServiceError,
			ErrMsg:  fmt.Sprintf("Error converting markdown:%s", err),
			Code:    0,
		}
	}
	return buf.String(), nil
}

// 实现按照一定大小切分
func SplitBySize(text string, size int) []string {
	var result []string
	for i := 0; i < len(text); i += size {
		end := i + size
		if end > len(text) {
			end = len(text)
		}
		result = append(result, text[i:end])
	}
	return result
}

// MemoryWriteSeeker 是一个基于内存的 WriteSeeker 实现
type MemoryWriteSeeker struct {
	buf    *bytes.Buffer
	offset int64
}

// NewMemoryWriteSeeker 创建一个新的 MemoryWriteSeeker 实例
func NewMemoryWriteSeeker() *MemoryWriteSeeker {
	return &MemoryWriteSeeker{
		buf: new(bytes.Buffer),
	}
}

// Write 实现 io.Writer 接口
func (mws *MemoryWriteSeeker) Write(p []byte) (n int, err error) {
	// 先调整缓冲区大小以适应新的写入
	if mws.offset != int64(mws.buf.Len()) {
		mws.buf.Bytes() = append(mws.buf.Bytes()[:mws.offset], p...)
	} else {
		mws.buf.Write(p)
	}
	n, err = len(p), nil
	mws.offset += int64(n)
	return
}

// Seek 实现 io.Seeker 接口
func (mws *MemoryWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = mws.offset + offset
	case io.SeekEnd:
		newOffset = int64(mws.buf.Len()) + offset
	default:
		return 0, fmt.Errorf("invalid whence")
	}
	if newOffset < 0 {
		return 0, fmt.Errorf("negative result position")
	}
	mws.offset = newOffset
	return mws.offset, nil
}
