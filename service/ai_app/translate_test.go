package ai_app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiktoken-go/tokenizer"
	"github.com/tmc/langchaingo/textsplitter"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

func TestNumTokensInString(t *testing.T) {
	tests := []struct {
		name        string
		inputStr    string
		encoding    tokenizer.Encoding
		expected    int
		expectError bool
	}{
		{
			name:        "Valid string with cl100k_base encoding",
			inputStr:    "Hello, world!",
			encoding:    "cl100k_base",
			expected:    3,
			expectError: false,
		},
		{
			name:        "Empty string with cl100k_base encoding",
			inputStr:    "",
			encoding:    "cl100k_base",
			expected:    0,
			expectError: false,
		},
		{
			name:        "Invalid encoding",
			inputStr:    "Hello, world!",
			encoding:    "invalid_encoding",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numTokens, err := numTokensInString(tt.inputStr, tt.encoding)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, numTokens)
			}
		})
	}
}

func TestCalculateChunkSize(t *testing.T) {
	tests := []struct {
		name       string
		tokenCount int
		tokenLimit int
		expected   int
	}{
		{
			name:       "Single chunk",
			tokenCount: 500,
			tokenLimit: 1000,
			expected:   500,
		},
		{
			name:       "Multiple chunks",
			tokenCount: 2500,
			tokenLimit: 1000,
			expected:   834,
		},
		{
			name:       "Exact multiple chunks",
			tokenCount: 3000,
			tokenLimit: 1000,
			expected:   1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunkSize := calculateChunkSize(tt.tokenCount, tt.tokenLimit)
			assert.Equal(t, tt.expected, chunkSize)
		})
	}
}

func TestTranslateApp_Translate(t *testing.T) {
	// Mock AI service for testing
	mockAI := &MockAI{}

	app := &TranslateApp{
		ai:           mockAI,
		defaultModel: "gpt-3.5-turbo",
	}

	tests := []struct {
		name        string
		sourceLang  string
		targetLang  string
		sourceText  string
		country     string
		maxTokens   int
		expectError bool
	}{
		{
			name:        "Single chunk translation",
			sourceLang:  "English",
			targetLang:  "French",
			sourceText:  "Hello, world!",
			country:     "",
			maxTokens:   1000,
			expectError: false,
		},
		{
			name:        "Multiple chunks translation",
			sourceLang:  "English",
			targetLang:  "French",
			sourceText:  "Hello, world! This is a longer text that will require multiple chunks to be translated.",
			country:     "",
			maxTokens:   10,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.Translate(tt.sourceLang, tt.targetLang, tt.sourceText, tt.country, tt.maxTokens, context.Background())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// MockAI is a mock implementation of the common.AI interface for testing purposes
type MockAI struct{}

func (m *MockAI) GetCompletion(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, *common.InnerError) {
	return "mock translation", nil
}
