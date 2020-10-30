package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/uc-cdis/ws-storage/storage"
	"github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	configPath := "/ws-storage.json"
	if len(os.Args) > 2 && strings.HasSuffix(os.Args[1], "-config") {
		configPath = os.Args[2]
	} else if len(os.Args) > 1 {
		fmt.Printf(
			`Use: ws-storage [-config path/to/config.json]
		- default config loaded from /ws-config.json
`)
		return
	}
	//, log.New(os.Stdout, "", log.LstdFlags)
	config, err := storage.LoadConfig(configPath)
	if err != nil {
		log.Error().Msg("Failed to load config - got %v", err)
		os.Exit(1)
	}
	switch config.LogLevel {
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

	}

	log.Info().Msg("whatever")
}
