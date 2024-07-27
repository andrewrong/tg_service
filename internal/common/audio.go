package common

import (
	"fmt"
	"io"

	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"

	"tg_ai_service/internal/log"
)

func OggToWav(reader io.ReadCloser, writer io.WriteSeeker) error {
	if reader == nil || writer == nil {
		log.Errorf("invalid reader or writer")
		return &InnerError{
			ErrType: ParameterError,
			ErrMsg:  "invalid reader or writer",
			Code:    0,
		}
	}

	streamer, format, err := vorbis.Decode(reader)
	if err != nil {
		log.Errorf("decode ogg error: %v", err)
		return &InnerError{
			ErrType: ExternalServiceError,
			ErrMsg:  fmt.Sprintf("decode ogg error:%s", err.Error()),
			Code:    0,
		}
	}
	defer streamer.Close()

	// 将音频数据从OGG流写入WAV文件
	err = wav.Encode(writer, streamer, format)
	if err != nil {
		log.Errorf("encode wav error: %v", err)
		return &InnerError{
			ErrType: ExternalServiceError,
			ErrMsg:  fmt.Sprintf("encode wav error:%s", err.Error()),
			Code:    0,
		}
	}
	return nil
}
