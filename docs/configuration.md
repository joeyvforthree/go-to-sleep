---
title: Configuration
description: Options for New: player cache, rate limit, logging, HTTP client.
---

`sleeper.New` takes functional options. With none, you get the defaults in the table.

```go
client := sleeper.New(
	sleeper.WithCacheDir("/var/cache/sleeper"),
	sleeper.WithCacheTTL(12*time.Hour),
	sleeper.WithRateLimit(500),
	sleeper.WithLogger(logger),
	sleeper.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)
```

| Option | Default | What it sets |
|--------|---------|--------------|
| `WithCacheDir(dir string)` | `~/.cache/sleeper` | Directory for the SQLite player cache |
| `WithCacheTTL(d time.Duration)` | 24h | How long cached players count as fresh |
| `WithRateLimit(rpm int)` | 1000 | Requests per minute, burst of 10 |
| `WithLogger(l *slog.Logger)` | `slog.Default()` | Where log records go |
| `WithHTTPClient(hc *http.Client)` | `http.DefaultClient` | Transport, timeouts, proxies |
| `WithBaseURL(url string)` | `https://api.sleeper.app/v1` | API root; for tests against `httptest` |

## HTTP client and timeouts

`http.DefaultClient` has no timeout. Pass your own client with one, or put a deadline on the
context of each call. The player dump is several megabytes, so leave it room.

## Rate limiting

Sleeper asks clients to stay under 1000 calls a minute. The client enforces that itself with a
token bucket: `rpm / 60` requests a second, bursting to 10. When the bucket is empty a call
waits for a token, and gives up if its context is cancelled first. Lower the limit if several
processes share one IP address; the limit is per client, not per machine.

## Player cache

`GetPlayers` returns every player in a sport, about 7 MB of JSON that Sleeper asks clients to
fetch at most once a day. The client keeps the last response per sport in `sleeper_players.db` under the cache
directory, using [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no
cgo), and serves it until it is older than the TTL.

- The database is created on the first `GetPlayers` call, not in `New`.
- If the directory cannot be created or opened, the error is logged and `GetPlayers` falls
  back to the network. A cache failure never fails the call.
- Only `GetPlayers` is cached. Everything else is fetched fresh each time.

In a container, point `WithCacheDir` at a writable volume, or the cache lives and dies with the
container.

## Logging

All records go to the logger from `WithLogger`, or `slog.Default()`.

| Level | Records |
|-------|---------|
| DEBUG | Each request (method, endpoint, status, duration), rate-limit waits, cache hits, misses and expiries |
| INFO | Cache writes (player count, size), progress of multi-week transaction fetches |
| WARN | Non-2xx responses with the first 200 bytes of the body, 429s |
| ERROR | Network failures, body read and JSON decode failures, cache open and store failures |

The client sends no credentials, so there is nothing secret to redact from these records.
