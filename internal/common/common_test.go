package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tg_ai_service/internal/log"
)

func TestGetFileType(t *testing.T) {
	// Test case 1: Valid file path
	filePath := "/tmp/audio/AwACAgUAAx0CXaIhXQACAbdmpX6MS2WM82W9lVnThWRmXNwb2QACyBAAAt-xKFUmQ9348S9iqzUE.ogg"
	ext, err := GetFileType(filePath)
	assert.NoError(t, err)
	log.Errorf("ext:%s", ext)

}
