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

// Fetch retrieves the environment variable with the given name,
// or returns the given default value if the variable is not set.
func Fetch(name, defaultValue string) string {
	value, ok := os.LookupEnv(name)

	if ok {
		return value
	}

	return defaultValue
}
