package log

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
	).Level(zerolog.TraceLevel).With().Timestamp().CallerWithSkipFrameCount(3).Logger()
}

func Infof(format string, a ...any) {
	logger.Info().Msg(fmt.Sprintf(format, a...))
}

func Errorf(format string, a ...any) {
	logger.Error().Msg(fmt.Sprintf(format, a...))
}

func Fatalf(format string, a ...any) {
	logger.Fatal().Msg(fmt.Sprintf(format, a...))
}

func Debugf(format string, a ...any) {
	logger.Debug().Msg(fmt.Sprintf(format, a...))
}

func Tracef(format string, a ...any) {
	logger.Trace().Msg(fmt.Sprintf(format, a...))
}

func Warnf(format string, a ...any) {
	logger.Warn().Msg(fmt.Sprintf(format, a...))
}

func Info(a ...any) {
	logger.Info().Msg(fmt.Sprint(a...))
}

func Error(a ...any) {
	logger.Error().Msg(fmt.Sprint(a...))
}

func Fatal(a ...any) {
	logger.Fatal().Msg(fmt.Sprint(a...))
}

func Debug(a ...any) {
	logger.Debug().Msg(fmt.Sprint(a...))
}

func Trace(a ...any) {
	logger.Trace().Msg(fmt.Sprint(a...))
}

func Warn(a ...any) {
	logger.Warn().Msg(fmt.Sprint(a...))
}
