package ai_app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiktoken-go/tokenizer"
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
			expected:    4,
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
				if err != nil {
					assert.Fail(t, "Unexpected error:", err)
				}
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
			tokenCount: 1530,
			tokenLimit: 500,
			expected:   389,
		},
		{
			name:       "Exact multiple chunks",
			tokenCount: 2242,
			tokenLimit: 500,
			expected:   496,
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
				if err != nil {
					assert.Fail(t, "Unexpected error:", err)
				}
			}
		})
	}
}

// MockAI is a mock implementation of the common.AI interface for testing purposes
type MockAI struct{}

func (m *MockAI) GetCompletion(prompt, systemMessage, model string, temperature float32, jsonModel bool, ctx context.Context) (string, error) {
	return "mock translation", nil
}

func TestGetContextBoundary(t *testing.T) {
	app := &TranslateApp{chunksInContext: 2}

	tests := []struct {
		name          string
		i             int
		maxBoundary   int
		expectedStart int
		expectedEnd   int
	}{
		{"Middle of the range", 5, 10, 4, 7},
		{"Start of the range", 0, 10, 0, 2},
		{"End of the range", 9, 10, 8, 10},
		{"Single element range", 0, 1, 0, 1},
		{"Small range", 1, 2, 0, 2},
		{"Large range", 50, 100, 49, 52},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := app.getContextBoundary(tt.i, tt.maxBoundary)
			if start != tt.expectedStart || end != tt.expectedEnd {
				t.Errorf("getContextBoundary(%d, %d) = (%d, %d); want (%d, %d)", tt.i, tt.maxBoundary, start, end, tt.expectedStart, tt.expectedEnd)
			}
		})
	}
}
