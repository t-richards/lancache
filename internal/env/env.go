// Package env provides helper functions for accessing environment variables.
package env

import (
	"os"
	"strings"
)

// BypassCache checks if the cache should be bypassed.
func BypassCache() bool {
	return strings.EqualFold(os.Getenv("BYPASS_CACHE"), "true")
}

// Production checks if the app is running in production mode.
func Production() bool {
	return strings.EqualFold(os.Getenv("APP_ENV"), "production")
}
