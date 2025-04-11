//go:build !prod

package lancache

import (
	"net/http"

	_ "net/http/pprof" //nolint:gosec

	"github.com/rs/zerolog/log"
)

const (
	pprofAddr = "localhost:6060"
)

// StartPprofServer starts a pprof server on localhost:6060.
// This is only enabled in non-production builds.
func StartPprofServer() {
	log.Info().Str("addr", pprofAddr).Msg("running pprof server")

	err := http.ListenAndServe(pprofAddr, nil) //nolint:gosec
	if err != nil {
		log.Fatal().Err(err).Msg("while starting pprof server")
	}
}
