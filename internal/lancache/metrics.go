package lancache

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	requests = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "lancache_requests_total",
			Help: "The total number of processed requests.",
		},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "lancache_http_duration_seconds",
			Help:    "The response time of requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"depot"},
	)

	cacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lancache_cache_hits_total",
			Help: "The total number of cache hits.",
		},
		[]string{"depot"},
	)

	cacheHitBytes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lancache_cache_hit_bytes_total",
			Help: "The total number of bytes served from the cache.",
		},
		[]string{"depot"},
	)

	cacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lancache_cache_misses_total",
			Help: "The total number of cache misses.",
		},
		[]string{"depot"},
	)

	cacheMissBytes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lancache_cache_miss_bytes_total",
			Help: "The total number of bytes fetched from the upstream server.",
		},
		[]string{"depot"},
	)

	cacheSkips = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lancache_cache_skips_total",
			Help: "The total number of cache skips.",
		},
		[]string{"depot"},
	)
)

func StartMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              ":9090",
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	log.Info().Str("addr", server.Addr).Msg("running metrics server")

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("while starting metrics server")
	}
}
