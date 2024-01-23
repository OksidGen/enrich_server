package pkg

import (
	"github.com/rs/zerolog"
	"os"
)

func SetupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevel := zerolog.InfoLevel

	if debugMode := os.Getenv("DEBUG"); debugMode == "TRUE" {
		logLevel = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(logLevel)
}
