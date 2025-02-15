# lancache

A proof-of-concept custom LAN cache for Steam.

## Features

:white_check_mark: Allows caching of specific games

:white_check_mark: Saves cached files as-is without slicing for visibility and easier maintenance

:white_check_mark: Provides useful cache metrics in Prometheus format

:white_check_mark: Lightweight image size (~10MB)

## Usage

Via docker compose:

```yaml
services:
  lancache:
    image: ghcr.io/t-richards/lancache:latest
    env:
      # - APP_ENV=production # (optional) set to "development" for colorful console output
      # - BYPASS_CACHE=true # (optional) bypass cache for troubleshooting
    ports:
      - "80:80"
      - "9090:9090" # (optional) prometheus metrics
    volumes:
      - /path/to/cache:/opt/cache
      - /path/to/lancache.toml:/opt/lancache.toml
```

## Configuration

```toml
# File: /opt/lancache.toml

[steam]
# array(integer): Which depots to cache.
# You can find the relevant IDs for depots on https://steamdb.info/
depots = [
  123, # Leave yourself a note about which game this is
  456, # ...or not, up to you.
]

# boolean: Default false.
# When true, the `depots` setting is ignored and all depots are cached.
cache_all = false
```

See also: a complete example [lancache.toml](lancache.toml).

## Contributing

```bash
# Build
make

# Run
bin/lancache
```

## License

Copyright (c) 2025 Tom Richards. All rights reserved.
