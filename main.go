package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/uc-cdis/ws-storage/storage"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var configPath string
	if len(os.Args) > 2 && strings.HasSuffix(os.Args[1], "-config") {
		configPath = os.Args[2]
	} else {
		fmt.Printf(
			`Use: ws-storage --config path/to/config.json
`)
		return
	}
	//, log.New(os.Stdout, "", log.LstdFlags)
	config, err := storage.LoadConfig(configPath)
	if err != nil {
		log.Error().Msgf("Failed to load config - got %v", err)
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
	mgr, err := storage.NewManager(config)
	if nil != err {
		log.Error().Msgf("Failed to initialize storage manager - got %v", err)
		os.Exit(1)
	}

	http.Handle("/metrics", promhttp.Handler())
	err = storage.SetupHttpListeners(mgr)
	if nil != err {
		log.Error().Msgf("Failed to setup listeners - got %v", err)
	}
	log.Info().Msg("ws-storage launching on port 8000")
	err = http.ListenAndServe("0.0.0.0:8000", nil)
	if nil != err {
		log.Error().Msgf("Failed to launch server on port 8000 - got %v", err)
		os.Exit(1)
	}
}
