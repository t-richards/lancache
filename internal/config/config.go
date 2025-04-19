// Package config provides a minimal API for using the application's config file.
package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
)

const defaultConfigPath = "lancache.toml"

var (
	// ErrConfigDecode is returned when the config file cannot be decoded.
	ErrConfigDecode = errors.New("while decoding config")
)

// LancacheConfig is the whole config file.
type LancacheConfig struct {
	Steam SteamConfig `toml:"steam"`
}

// SteamConfig provides Steam-specific configuration options.
type SteamConfig struct {
	Depots   []uint32 `toml:"depots"` // From steamtypes.h
	CacheAll bool     `toml:"cache_all"`
}

// Load reads and decodes the config file from the default location.
func Load() (*LancacheConfig, error) {
	config := &LancacheConfig{}

	_, err := toml.DecodeFile(defaultConfigPath, &config)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigDecode, err)
	}

	slices.Sort(config.Steam.Depots)

	if config.Steam.CacheAll {
		log.Info().Msg("caching all depots")
	} else {
		log.Info().Uints32("depots", config.Steam.Depots).Msg("caching depots")
	}

	return config, nil
}

// HasDepot checks if a depot is in the list of depots to cache.
//
// Binary search performance is superior to a map lookup until ~250 elements;
// if you're caching more games than that, consider using the CacheAll option.
func (c *LancacheConfig) HasDepot(depot string) bool {
	intDepot, err := strconv.ParseUint(depot, 10, 32)
	if err != nil {
		log.Error().Str("depot", depot).Msg("could not parse depot ID as int64")

		return false
	}

	uintDepot := uint32(intDepot)
	_, found := slices.BinarySearch(c.Steam.Depots, uintDepot)

	return found
}
