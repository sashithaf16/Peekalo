package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger is an alias for zerolog.Logger.
// This avoids referencing a third-party package directly in the codebase
type Logger = zerolog.Logger

func CreateLogger(lgL string) Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevel, err := zerolog.ParseLevel(lgL)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger.Info().Msg("Logger initialized")
	return logger
}
