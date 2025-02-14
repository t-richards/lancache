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
	ErrConfigDecode = errors.New("while decoding config")
)

// Exported types for consumers of the configuration API.
type LancacheConfig struct {
	Steam SteamConfig `toml:"steam"`
}

type SteamConfig struct {
	Depots   []int64 `toml:"depots"`
	CacheAll bool    `toml:"cache_all"`
}

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
		log.Info().Ints64("depots", config.Steam.Depots).Msg("caching depots")
	}

	return config, nil
}

// HasDepot checks if a depot is in the list of depots to cache.
//
// Binary search performance is superior to a map lookup until ~250 elements;
// if you're caching more games than that, consider using the CacheAll option.
func (c *LancacheConfig) HasDepot(depot string) bool {
	intDepot, err := strconv.ParseInt(depot, 10, 64)
	if err != nil {
		log.Error().Str("depot", depot).Msg("could not parse depot ID as int64")

		return false
	}

	_, found := slices.BinarySearch(c.Steam.Depots, intDepot)

	return found
}
