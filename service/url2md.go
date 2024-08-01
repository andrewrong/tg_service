package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type Url2MdService struct {
	config *common.Url2MdConfig
	client *http.Client
}

func NewUrl2MdService(config *common.Url2MdConfig) (*Url2MdService, error) {
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

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    60 * time.Second,
		DisableCompression: true,
	}

	return &Url2MdService{
		config: config,
		client: &http.Client{
			Transport: tr,
			Timeout:   time.Duration(config.TimeoutMs) * time.Millisecond,
		},
	}, nil
}

type UrlContent struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Excerpt string `json:"excerpt"`
}

func (u *Url2MdService) GetUrl2Md(url string) (*UrlContent, error) {
	if _, ok := common.IsValidURL(url); !ok {
		log.Errorf("url is error, url:%s", url)
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  "url is error",
			Code:    0,
		}
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(u.config.TimeoutMs)*time.Millisecond)
	defer cancel()

	req, err := u.newRequest(u.config.BearerToken, http.MethodGet, nil, ctx)
	if err != nil {
		return nil, &common.InnerError{
			ErrType: common.ParameterError,
			ErrMsg:  err.Error(),
			Code:    0,
		}
	}
	query := req.URL.Query()
	query.Add("url", url)
	req.URL.RawQuery = query.Encode()

	resp, err := u.client.Do(req)
	if err != nil {
		log.Errorf("url is error, url:%s, err:%v", url, err)
		return nil, &common.InnerError{
			ErrType: common.InitError,
			ErrMsg:  err.Error(),
			Code:    0,
		}
	}

	defer resp.Body.Close()
	var content UrlContent
	err = json.NewDecoder(resp.Body).Decode(&content)
	if err != nil {
		log.Errorf("url is error, url:%s, err:%v", url, err)
		return nil, &common.InnerError{
			ErrType: common.InitError,
			ErrMsg:  err.Error(),
			Code:    0,
		}
	}
	return &content, nil
}

func (c *Url2MdService) newRequest(token string, httpMethod string, body interface{}, ctx context.Context) (*http.Request, error) {

	u, err := url.QueryUnescape(c.config.ServiceUrl)
	if err != nil {
		log.Errorf("url is error, url:%s", c.config.ServiceUrl)
		return nil, err
	}

	var b bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&b).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	var req *http.Request
	if ctx != nil {
		req, err = http.NewRequestWithContext(ctx, httpMethod, u, &b)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(httpMethod, u, &b)
		if err != nil {
			return nil, err
		}
	}

	req.Header.Add("Content-Type", "application/json")

	if token != "" {
		bearerToken := fmt.Sprintf("Bearer %s", token)
		req.Header.Add("Authorization", bearerToken)
	}

	return req, nil
}
