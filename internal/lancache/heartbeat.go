package lancache

import "net/http"

const (
	lancacheProcessedByHeader = "X-LanCache-Processed-By"
)

func heartbeatHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header()[lancacheProcessedByHeader] = []string{"lancache"}
	w.WriteHeader(http.StatusNoContent)
}
