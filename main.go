// Package main is the primary entrypoint for the app.
package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/t-richards/lancache/internal/env"
	"github.com/t-richards/lancache/internal/lancache"
)

func main() {
	if !env.Production() {
		// Use pretty logging in development.
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		currentEnv := os.Getenv("APP_ENV")
		log.Info().Str("APP_ENV", currentEnv).Msg("non-production environment; using pretty logging")
	}

	// Setup application.
	app := lancache.New()

	// Background services.
	go lancache.StartPprofServer()
	go lancache.StartMetricsServer()

	// Run the cache server.
	app.StartCacheServer()
}
