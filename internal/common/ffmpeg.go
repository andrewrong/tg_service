package common

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"tg_ai_service/internal/log"
)

func FormatTransformer(inputFile, outputFile string) error {
	err := ffmpeg.Input(inputFile).
		Output(outputFile).
		OverWriteOutput().
		Run()
	if err != nil {
		log.Errorf("ffmpeg error: %v", err)
		return err
	}
	return nil
}

// ogg to mp3
func OggToMp3(inputFile, outputFile string) error {
	err := ffmpeg.Input(inputFile).
		Output(outputFile).
		OverWriteOutput().
		Run()

	if err != nil {
		log.Errorf("ffmpeg error: %v", err)
		return err
	}
	return nil
}
