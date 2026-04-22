# go-to-sleep

Go client library for the [Sleeper](https://sleeper.com) fantasy sports API.

Typed response models, SQLite-backed player cache, built-in rate limiting, and structured logging via `log/slog`.

## Install

```bash
go get github.com/joeyvforthree/go-to-sleep
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    sleeper "github.com/joeyvforthree/go-to-sleep"
)

func main() {
    client := sleeper.New()
    ctx := context.Background()

    // Get current NFL state
    state, err := client.GetSportState(ctx, "nfl")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("NFL %s — Week %d (%s)\n", state.Season, state.Week, state.SeasonType)

    // Look up a user
    user, err := client.GetUser(ctx, "joeyvfor3")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("User: %s (%s)\n", user.Username, user.UserID)

    // Get league rosters
    rosters, err := client.GetRosters(ctx, "1350223992606195712")
    if err != nil {
        log.Fatal(err)
    }
    for _, r := range rosters {
        fmt.Printf("Roster %d: %d players\n", r.RosterID, len(r.Players))
    }
}
```

## Configuration

```go
client := sleeper.New(
    sleeper.WithCacheDir("/tmp/sleeper"),          // player cache location (default: ~/.cache/sleeper/)
    sleeper.WithCacheTTL(12 * time.Hour),          // cache TTL (default: 24h)
    sleeper.WithLogger(myLogger),                  // custom *slog.Logger
    sleeper.WithRateLimit(500),                    // requests per minute (default: 1000)
    sleeper.WithHTTPClient(myHTTPClient),          // custom *http.Client
)
```

## API Coverage

### Users
```go
user, err := client.GetUser(ctx, "username_or_id")
```

### Leagues
```go
leagues, err := client.GetUserLeagues(ctx, userID, "nfl", "2025")
league, err  := client.GetLeague(ctx, leagueID)
rosters, err := client.GetRosters(ctx, leagueID)
users, err   := client.GetLeagueUsers(ctx, leagueID)
matchups, err := client.GetMatchups(ctx, leagueID, week)
winners, err := client.GetWinnersBracket(ctx, leagueID)
losers, err  := client.GetLosersBracket(ctx, leagueID)
picks, err   := client.GetTradedPicks(ctx, leagueID)
```

### Transactions
```go
txns, err := client.GetTransactions(ctx, leagueID, week)

// Fetch all weeks at once
allTxns, err := client.GetAllTransactions(ctx, leagueID, 18)

// Filter to a specific player
playerTxns, err := client.GetPlayerTransactions(ctx, leagueID, "4046", 18)
```

### Drafts
```go
drafts, err := client.GetUserDrafts(ctx, userID, "nfl", "2025")
drafts, err := client.GetLeagueDrafts(ctx, leagueID)
draft, err  := client.GetDraft(ctx, draftID)
picks, err  := client.GetDraftPicks(ctx, draftID)
traded, err := client.GetTradedDraftPicks(ctx, draftID)
```

### Players
```go
// Returns 10,000+ players — automatically cached in SQLite
players, err := client.GetPlayers(ctx, "nfl")

// Trending adds/drops
trending, err := client.GetTrendingPlayers(ctx, "nfl", "add", 24, 25)
```

### Sport State
```go
state, err := client.GetSportState(ctx, "nfl")
// state.Season, state.Week, state.SeasonType
```

## Error Handling

```go
import "errors"

league, err := client.GetLeague(ctx, "bad_id")
if errors.Is(err, sleeper.ErrNotFound) {
    // 404 or null response
}
if errors.Is(err, sleeper.ErrRateLimited) {
    // 429 — server-side rate limit hit
}
if errors.Is(err, sleeper.ErrServerError) {
    // 5xx
}

// Full error details
var apiErr *sleeper.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("Status %d on %s: %s\n", apiErr.StatusCode, apiErr.Endpoint, apiErr.Body)
}
```

## Player Cache

`GetPlayers()` responses are large (~7MB) and change infrequently. The client automatically caches them in a local SQLite database using [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no cgo required).

- Default location: `~/.cache/sleeper/sleeper_players.db`
- Default TTL: 24 hours
- Cache is checked before every `GetPlayers()` call
- Configure with `WithCacheDir()` and `WithCacheTTL()`

## Logging

All HTTP activity is logged via `log/slog` at appropriate levels:

| Level | Events |
|-------|--------|
| DEBUG | HTTP requests, cache hits/misses, rate limit waits |
| INFO  | Cache writes, multi-week fetch progress |
| WARN  | Non-2xx responses, rate limit hits |
| ERROR | Network failures, JSON decode errors |

Pass a custom logger with `WithLogger()` or the client uses `slog.Default()`.

## License

[MIT](LICENSE)
