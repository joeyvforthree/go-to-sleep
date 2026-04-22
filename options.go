package sleeper

import (
	"log/slog"
	"net/http"
	"time"
)

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default Sleeper API base URL (useful for testing).
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithCacheDir sets the directory for the SQLite player cache.
// Default: ~/.cache/sleeper/
func WithCacheDir(dir string) Option {
	return func(c *Client) { c.cacheDir = dir }
}

// WithCacheTTL sets how long cached player data is considered fresh.
// Default: 24h.
func WithCacheTTL(d time.Duration) Option {
	return func(c *Client) { c.cacheTTL = d }
}

// WithLogger sets a structured logger. Default: slog.Default().
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) { c.logger = l }
}

// WithRateLimit sets the maximum requests per minute. Default: 1000.
func WithRateLimit(rpm int) Option {
	return func(c *Client) { c.rateRPM = rpm }
}
