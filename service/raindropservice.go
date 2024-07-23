package service

//
//import (
//	"errors"
//	"time"
//
//	rd "github.com/andrewrong/raindrop-io-api-client/pkg/raindrop"
//	"golang.org/x/net/context"
//
//	"tg_ai_service/internal/common"
//	"tg_ai_service/internal/log"
//)
//
//func IsSpecialCollectionId(id int) bool {
//	return id == rd.UNSORT_COLLECTION_ID || id == rd.TRASH_COLLECTION_ID || id == rd.ALL_COLLECTION_ID
//}
//
//type RdConfig struct {
//	ApiToken  string `json:"api_token"`  // raindrop client的token
//	ClientId  string `json:"client_id"`  // client id
//	Redirect  string `json:"redirect"`   // redirect uri
//	TimeoutMs int32  `json:"timeout_ms"` // 超时时间
//	PerPage   int    `json:"per_page"`   //每页的个数
//
//	DefaultCollection   string `json:"default_collection"` // 默认添加的集合
//	defaultCollectionId int32
//}
//
//func (rd *RdConfig) Check() error {
//	if rd.ApiToken == "" {
//		return &common.InnerError{
//			ErrType: common.ParameterError,
//			ErrMsg:  "api token is empty",
//			Code:    0,
//		}
//	}
//
//	if rd.ClientId == "" {
//		return &common.InnerError{
//			ErrType: common.ParameterError,
//			ErrMsg:  "client id is empty",
//			Code:    0,
//		}
//	}
//
//	if rd.Redirect == "" {
//		return &common.InnerError{
//			ErrType: common.ParameterError,
//			ErrMsg:  "redirect is empty",
//			Code:    0,
//		}
//	}
//
//	if rd.TimeoutMs <= 0 {
//		rd.TimeoutMs = 5000
//	}
//
//	if rd.PerPage <= 0 {
//		rd.PerPage = 20
//	}
//	return nil
//}
//
//type RaindropService struct {
//	client        *rd.Client
//	config        *RdConfig
//	ai            *AiService //一个ai服务，用来做LLM工作
//	url2MdService *Url2MdService
//}
//
//func NewRaindropService(rdConfig *RdConfig) (*RaindropService, error) {
//	var err error = nil
//	if rdConfig == nil {
//		log.Errorf("init raindrop is error, rdConfig is nil")
//		return nil, &common.InnerError{
//			ErrType: common.ParameterError,
//			ErrMsg:  "rdConfig is nil",
//			Code:    0,
//		}
//	}
//
//	err = rdConfig.Check()
//	if err != nil {
//		log.Errorf("init raindrop is error, err:%v", err)
//		return nil, &common.InnerError{
//			ErrType: common.ParameterError,
//			ErrMsg:  err.Error(),
//			Code:    0,
//		}
//	}
//
//	rds := &RaindropService{
//		config: rdConfig,
//	}
//	rds.client, err = rd.NewClient(rdConfig.ClientId, rdConfig.ApiToken, rdConfig.Redirect)
//
//	if err != nil {
//		log.Errorf("init raindrop is error, err:%v", err)
//		return nil, &common.InnerError{
//			ErrType: common.InitError,
//			ErrMsg:  err.Error(),
//			Code:    0,
//		}
//	}
//	return rds, nil
//}
//
//// query
//
//func (rds *RaindropService) GetCollection(collectionId int) ([]rd.Collection, error) {
//	var response *rd.GetCollectionsResponse = nil
//	var err error = nil
//
//	ctx := context.Background()
//	ctx, cancel := context.WithTimeout(ctx, time.Duration(rds.config.TimeoutMs)*time.Millisecond)
//	defer cancel()
//
//	if collectionId <= 0 && !IsSpecialCollectionId(collectionId) {
//		response, err = rds.client.GetRootCollections(rds.config.ApiToken, ctx)
//	} else {
//		response, err = rds.client.GetChildCollections(rds.config.ApiToken, ctx)
//	}
//
//	if err != nil {
//		log.Errorf("get collection is error, err:%v", err)
//		if errors.Is(err, context.Canceled) {
//			return nil, &common.InnerError{
//				ErrType: common.ContextCancelError,
//				ErrMsg:  err.Error(),
//				Code:    0,
//			}
//		}
//		return nil, &common.InnerError{
//			ErrType: common.ExternalServiceError,
//			ErrMsg:  err.Error(),
//			Code:    0,
//		}
//	}
//	return response.Items, nil
//}
//
//func (rds *RaindropService) GetRaindropsInCollection(collectId int, search string) ([]rd.Raindrop, error) {
//	var response *rd.MultiRaindropsResponse = nil
//	var err error = nil
//
//	var ctx = context.Background()
//	ctx, cancel := context.WithTimeout(ctx, time.Duration(rds.config.TimeoutMs)*time.Millisecond)
//	defer cancel()
//
//	if collectId <= 0 && !IsSpecialCollectionId(collectId) {
//		log.Errorf("collection id is error, collectId:%v", collectId)
//		return nil, &common.InnerError{
//			ErrType: common.ParameterError,
//			ErrMsg:  "collection id is error",
//			Code:    0,
//		}
//	}
//	items := make([]rd.Raindrop, 0)
//	pageIdx := 0
//
//	searchParams := &rd.SearchRdParams{
//		Search:       search,
//		Page:         0,
//		PageSize:     rds.config.PerPage,
//		Sort:         "-created",
//		CollectionId: collectId,
//	}
//
//	for {
//		response, err = rds.client.GetRaindrops(rds.config.ApiToken, searchParams, ctx)
//
//		if err != nil {
//			log.Errorf("get raindrops is error, err:%v", err)
//			if errors.Is(err, context.Canceled) {
//				return nil, &common.InnerError{
//					ErrType: common.ContextCancelError,
//					ErrMsg:  err.Error(),
//					Code:    0,
//				}
//			}
//			return nil, &common.InnerError{
//				ErrType: common.ExternalServiceError,
//				ErrMsg:  err.Error(),
//				Code:    0,
//			}
//		}
//		if len(response.Items) == 0 {
//			log.Infof("get raindrops is empty, so end, raindrop count:%d", len(items))
//			break
//		}
//		items = append(items, response.Items...)
//		pageIdx++
//		searchParams.Page = pageIdx
//		break
//	}
//
//	return items, nil
//}
//
//// write
//
//func (rds *RaindropService) AddRd2Collection(url string, collectionId int) error {
//	if collectionId <= 0 && !IsSpecialCollectionId(collectionId) {
//		log.Errorf("collection id is error, collectionId:%v", collectionId)
//		return &common.InnerError{
//			ErrType: common.ParameterError,
//			ErrMsg:  "collection id is error",
//			Code:    0,
//		}
//	}
//
//	// 1. check url,是否符合一个正常的url
//	// 2. 检查collection，如果为空就设置成默认的collection
//	// 3. 检查当前url在rd中是否存在，如果存在就不进行添加, api不支持
//	// 4. 通过ai进行分析，获得几个信息
//	// 5. 组合成raindrop的数据结构然后添加到rd中去
//	if _, ok := common.IsValidURL(url); !ok {
//		log.Errorf("url is error, url:%v", url)
//		return &common.InnerError{
//			ErrType: common.ParameterError,
//			ErrMsg:  "url is error",
//			Code:    0,
//		}
//	}
//
//}
