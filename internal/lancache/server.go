// Package lancache is responsible for handling HTTP requests and caching files.
package lancache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/t-richards/lancache/internal/config"
	"github.com/t-richards/lancache/internal/env"
)

// Directory where cached files are stored.
const cacheDir = "cache"

// Default permissions on created directories.
const cacheDirPerms = 0755

// Timeout for server shutdown.
const shutdownTimeout = 5 * time.Second

// Timeout for reading headers on requests.
const readHeaderTimeout = 10 * time.Second

// ResponseError is returned when the upstream server returns a non-200 response.
type ResponseError interface {
	error

	StatusCode() int
}

type responseError struct {
	error

	code int
}

func (err responseError) Error() string {
	return fmt.Sprintf("upstream server returned %d", err.code)
}
func (err responseError) Unwrap() error {
	return err.error
}
func (err responseError) StatusCode() int {
	return err.code
}

// Application is the primary "thing" we're running here. It handles requests and caching.
type Application struct {
	cacheConfig *config.LancacheConfig
	httpClient  *http.Client
}

// New creates an instance of the application, and instantiates required dependencies for it to run.
func New() *Application {
	// Set up cache directory.
	err := os.MkdirAll(cacheDir, cacheDirPerms)
	if err != nil {
		log.Fatal().Err(err).Msg("while creating cache directory")
	}

	// Create application with HTTP client.
	app := &Application{
		httpClient: newHTTPClient(),
	}

	// Parse configuration.
	app.cacheConfig, err = config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("while loading configuration")
	}

	return app
}

func newHTTPClient() *http.Client {
	t, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		log.Fatal().Msg("http.DefaultTransport is not an *http.Transport")
	}

	// Allow either HTTP/1 or HTTP/2.
	t = t.Clone()
	t.Protocols = new(http.Protocols)
	t.Protocols.SetHTTP1(true)
	t.Protocols.SetHTTP2(true)

	return &http.Client{Transport: t}
}

// StartCacheServer runs the primary HTTP server for the app.
func (a *Application) StartCacheServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/lancache-heartbeat", heartbeatHandler)
	mux.HandleFunc("/depot/{depot}/", a.lancacheHandler)
	server := &http.Server{
		Addr:              ":80",
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	log.Info().Str("addr", server.Addr).Msg("running lancache server")

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("while running lancache server")
		}
	}()

	// Allow for graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info().Stringer("signal", sig).Msg("signal received, stopping lancache server")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}

	log.Info().Msg("stopped lancache server")
}

func shouldCache(cacheConfig *config.LancacheConfig, depot string) bool {
	if env.BypassCache() {
		return false
	}

	if cacheConfig.HasDepot(depot) {
		return true
	}

	if cacheConfig.Steam.CacheAll {
		return true
	}

	return false
}

// Returns a normalized path for use in the cache directory.
// May not be valid for Windows, but we don't care about Windows.
func clean(path string) string {
	if strings.HasPrefix(path, "/") {
		return filepath.Clean(path)
	}

	return filepath.Clean("/" + path)
}

func (a *Application) lancacheHandler(w http.ResponseWriter, r *http.Request) {
	requests.Inc()

	startTime := time.Now()
	depot := r.PathValue("depot")
	logger := log.With().Str("depot", depot).Str("host", r.Host).Str("path", r.URL.Path).Logger()

	// Record the duration of the request.
	defer func() {
		duration := time.Since(startTime).Seconds()
		httpDuration.WithLabelValues(depot).Observe(duration)
	}()

	// Do we care about caching this file?
	if !shouldCache(a.cacheConfig, depot) {
		// We don't want to cache this. Tell the Steam client to fetch it directly.
		cacheSkips.WithLabelValues(depot).Inc()
		logger.Info().Msg("skip")

		newLocation := "https://" + r.Host + r.URL.Path
		http.Redirect(w, r, newLocation, http.StatusSeeOther)

		return
	}

	cachePath := filepath.Join(cacheDir, clean(r.URL.Path))

	// We are interested in caching the file.
	fileInfo, err := os.Stat(cachePath)
	if err == nil {
		// We have this file already. Deliver it directly from the cache.
		cacheHits.WithLabelValues(depot).Inc()
		cacheHitBytes.WithLabelValues(depot).Add(float64(fileInfo.Size()))
		logger.Info().Msg("hit")
		http.ServeFile(w, r, cachePath)

		return
	}

	// We don't have the file.
	cacheMisses.WithLabelValues(depot).Inc()
	logger.Info().Msg("miss")

	// Prepare the directory to store the file.
	err = os.MkdirAll(filepath.Dir(cachePath), cacheDirPerms)
	if err != nil {
		logger.Error().Err(err).Msg("while creating cache directory")
		http.Error(w, "Failed to create cache directory", http.StatusInternalServerError)

		return
	}

	// Fetch it from the upstream server and cache it.
	err = a.fetchAndCache(logger, w, r, depot, cachePath)
	if err != nil {
		var responseErr ResponseError
		if errors.As(err, &responseErr) {
			// The upstream server returned a non-200 response.
			http.Error(w, err.Error(), responseErr.StatusCode())

			return
		}
		// By this point, we may have already written some response data to the client.
		// We can't change response headers now, so we log the error and move on.
		logger.Error().Err(err).Msg("while caching upstream")

		return
	}
}

// fetchAndCache fetches a file from the upstream server and caches it.
//
// Using a temporary file ensures:
// - No exposure of partially downloaded files.
// - No need for complex locking mechanisms.
// - Atomic disk writes via sync/rename.
//
// Multiple clients may trigger multiple fetches, which is acceptable.
func (a *Application) fetchAndCache(
	logger zerolog.Logger,
	w http.ResponseWriter,
	r *http.Request,
	depotID string,
	filename string,
) (err error) {
	// Construct the request to the upstream server.
	upstreamReq, err := createUpstreamRequest(r)
	if err != nil {
		return err
	}

	// Create a temporary file to store the response.
	tmpFile, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp")
	if err != nil {
		return fmt.Errorf("while creating temporary file: %w", err)
	}

	// Ensure the temporary file is cleaned up if anything fails.
	defer func() {
		if err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpFile.Name())
		}
	}()

	// Perform the request.
	resp, err := a.httpClient.Do(upstreamReq)
	if err != nil {
		return responseError{
			error: err,
			code:  http.StatusBadGateway,
		}
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			logger.Error().Err(closeErr).Msg("while closing response body")
		}
	}()

	// Ensure we have a 200 OK response.
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)

		return responseError{
			code: resp.StatusCode,
		}
	}

	if resp.ContentLength > 0 {
		w.Header().Add("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
		cacheMissBytes.WithLabelValues(depotID).Add(float64(resp.ContentLength))
	}

	// Send file contents to the Steam client and the temporary file.
	multiWriter := io.MultiWriter(w, tmpFile)

	_, err = io.Copy(multiWriter, resp.Body)
	if err != nil {
		return fmt.Errorf("while copying response: %w", err)
	}

	return finalizeTmpFile(tmpFile, filename)
}

func finalizeTmpFile(tmpFile *os.File, filename string) error {
	tmpName := tmpFile.Name()

	err := tmpFile.Sync()
	if err != nil {
		return fmt.Errorf("while syncing temporary file: %w", err)
	}

	err = tmpFile.Close()
	if err != nil {
		return fmt.Errorf("while closing temporary file: %w", err)
	}

	err = os.Rename(tmpName, filename)
	if err != nil {
		return fmt.Errorf("while renaming temporary file: %w", err)
	}

	return nil
}

func createUpstreamRequest(r *http.Request) (*http.Request, error) {
	url := "https://" + r.Host + r.URL.Path

	upstreamReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("while creating upstream request: %w", err)
	}

	upstreamReq.Header.Add("User-Agent", "lancache/0.0.1 (+https://trnet.cc/lancache)")

	return upstreamReq, nil
}
