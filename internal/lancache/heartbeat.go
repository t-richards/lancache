package lancache

import "net/http"

const (
	lancacheProcessedByHeader = "X-LanCache-Processed-By"
)

func heartbeatHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	w.Header().Set(lancacheProcessedByHeader, "lancache") //nolint:canonicalheader
}
