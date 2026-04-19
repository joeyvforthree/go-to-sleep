# go-to-sleep

Go client library for the Sleeper fantasy sports API.

## Build & Test

```bash
go build ./...
go vet ./...
go test ./...
```

## Architecture

Flat package `sleeper` in the repo root. All source files:

- `models.go` — API response structs. `Player` has custom `UnmarshalJSON` that captures unknown fields in `Extra`.
- `errors.go` — `APIError` type wrapping sentinel errors (`ErrNotFound`, `ErrRateLimited`, `ErrServerError`). Use `errors.Is` / `errors.As`.
- `options.go` — Functional options for `New()`: `WithBaseURL`, `WithHTTPClient`, `WithCacheDir`, `WithCacheTTL`, `WithLogger`, `WithRateLimit`.
- `client.go` — HTTP client with rate limiting (`golang.org/x/time/rate`), structured logging (`log/slog`), and all API methods. All methods take `context.Context` as first arg.
- `cache.go` — SQLite-backed player cache (`modernc.org/sqlite`). TTL-based expiry, default 24h.

## Dependencies

- `golang.org/x/time/rate` — token bucket rate limiter
- `modernc.org/sqlite` — pure Go SQLite (no cgo)

Do not add new dependencies without explicit justification.

## Conventions

- All public methods take `context.Context` as the first parameter.
- Errors from non-2xx API responses are `*APIError` and wrap sentinels via `Unwrap()`.
- Null JSON responses (Sleeper returns `"null"` for missing resources) produce `ErrNotFound`.
- Rate limiter defaults to 1000 req/min with burst of 10.
- Tests use `net/http/httptest` — no live API calls in tests.
