package service

import (
	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

type ExternalService struct {
	url2md *Url2MdService
	//raindrop *RaindropService
	fileReader *FileReaderService
}

func NewExternalService(tgCfg *common.ExternalConfig) (*ExternalService, error) {
	external := &ExternalService{}
	var err error = nil
	if tgCfg.Url2MdCfg != nil {
		external.url2md, err = NewUrl2MdService(tgCfg.Url2MdCfg)
		if err != nil {
			return nil, err
		}
	}

	//if tgCfg.RdConfig != nil {
	//	external.raindrop, err = NewRaindropService(tgCfg.RdConfig)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	if tgCfg.FileReaderCfg != nil {
		external.fileReader, err = NewFileReaderService(tgCfg.FileReaderCfg)
		if err != nil {
			return nil, err
		}
	}
	return external, nil
}

func (s *ExternalService) GetUrl2MdService() *Url2MdService {
	if s.url2md == nil {
		log.Errorf("this app is not support url2md")
		return nil
	}
	return s.url2md
}

func (s *ExternalService) GetFileReaderService() *FileReaderService {
	if s.fileReader == nil {
		log.Errorf("this app is not support file reader")
		return nil
	}
	return s.fileReader
}
