# lancache

A proof-of-concept custom LAN cache for Steam.

## Features

:white_check_mark: Allows caching of specific games

:white_check_mark: Saves cached files as-is without slicing for visibility and easier maintenance

:white_check_mark: Provides useful cache metrics in Prometheus format

:white_check_mark: Lightweight image size (~10MB)

## Usage

1. Run the app (via docker compose, for example):

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

> [!NOTE]
> The cache directory must be writable by the container.
> The container runs as the `app` user with UID 1234.

2. Set a DNS record for `lancache.steamcontent.com` pointing to the host running the app.

> [!NOTE]
> Steam will reject your LAN cache if the response TTL is too high.
> Additionally, some DNS resolvers will still return public AAAA records even if you override the A record for a given hostname.
> We suggest using the following CNAME -> A configuration for best results.

```
$ dig lancache.steamcontent.com
<snip>
;; ANSWER SECTION:
lancache.steamcontent.com.   5     IN      CNAME   lancache.lan.
lancache.lan.              600     IN      A       192.168.1.121
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

Requirements:

 - Go `=1.24`

```bash
# Build
make

# Run
bin/lancache
```

## License

Copyright (c) 2025 Tom Richards. All rights reserved.
