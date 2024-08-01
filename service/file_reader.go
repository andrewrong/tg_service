package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/resty.v1"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type FileReaderService struct {
	config *common.FileReaderConfig
}

func NewFileReaderService(config *common.FileReaderConfig) (*FileReaderService, error) {
	if config == nil {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "config is empty",
			Code:    0,
		}
	}

	err := config.Check()
	if err != nil {
		return nil, err
	}
	return &FileReaderService{
		config: config,
	}, nil
}

func (s *FileReaderService) OnlinePdfReader(url string, ctx context.Context) (string, error) {
	if url == "" {
		return "", nil
	}

	if _, ok := common.IsValidURL(url); !ok {
		log.Errorf("invalid url: %s", url)
		return "", &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "invalid url",
			Code:    0,
		}
	}

	jinaUrl := fmt.Sprintf("%s%s", s.config.PdfBaseUrl, url)
	resp, err := resty.R().SetContext(ctx).Get(jinaUrl)

	if err != nil {
		log.Errorf("jina pdf is error: %v", err)
		return "", &common.InnerError{
			ErrType: common.ExternalServiceError,
			ErrMsg:  "error",
			Code:    0,
		}
	}
	content := resp.String()
	return content, nil
}

func (s *FileReaderService) readDoc(path string) (string, error) {
	ext := filepath.Ext(path)
	switch ext {
	default:
		return s.extractText(path)
	}
}

func (s *FileReaderService) extractText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *FileReaderService) extractEpub(path string) (string, error) {
	return "", nil
}
