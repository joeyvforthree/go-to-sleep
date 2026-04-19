# sleeper-go

## Context
Reusable Go client for the [Sleeper API](https://docs.sleeper.com) (`https://api.sleeper.app/v1/`). The API is free, read-only, and requires no auth. Initial focus is NFL, but the design supports any sport Sleeper offers. Provides typed structs, a SQLite-backed player cache, and convenience methods for common operations.

## Project Setup
- **Language:** Go (1.22+)
- **Module:** `github.com/joeyvforthree/sleeper-go`
- **Dependencies:**
  - `modernc.org/sqlite` — pure Go SQLite driver (no cgo) for player cache
  - No other third-party deps — stdlib `net/http` + `encoding/json` for everything else

## Project Structure
```
sleeper-go/
├── go.mod
├── go.sum
├── CLAUDE.md
├── PLAN.md
├── client.go              # Client — HTTP client, all API methods
├── client_test.go
├── options.go             # functional options (WithBaseURL, WithCacheDir, etc.)
├── models.go              # All structs: User, League, Roster, Player, etc.
├── cache.go               # SQLite-backed player cache
├── cache_test.go
└── testdata/              # fixture JSON for tests
```

## Client Design (`client.go`)

```go
package sleeper

type Client struct {
    httpClient *http.Client
    baseURL    string
    cache      *PlayerCache
}

func New(opts ...Option) *Client

// Users
func (c *Client) GetUser(identifier string) (*User, error)

// Leagues
func (c *Client) GetUserLeagues(userID, sport, season string) ([]League, error)
func (c *Client) GetLeague(leagueID string) (*League, error)
func (c *Client) GetRosters(leagueID string) ([]Roster, error)
func (c *Client) GetLeagueUsers(leagueID string) ([]LeagueUser, error)
func (c *Client) GetMatchups(leagueID string, week int) ([]Matchup, error)
func (c *Client) GetWinnersBracket(leagueID string) ([]BracketMatch, error)
func (c *Client) GetLosersBracket(leagueID string) ([]BracketMatch, error)
func (c *Client) GetTransactions(leagueID string, week int) ([]Transaction, error)
func (c *Client) GetTradedPicks(leagueID string) ([]TradedPick, error)

// Convenience — transactions
func (c *Client) GetAllTransactions(leagueID string, seasonWeeks int) ([]Transaction, error)
    // Fetches transactions for weeks 1..seasonWeeks. If seasonWeeks <= 0,
    // uses GetSportState() to determine the current week.
func (c *Client) GetPlayerTransactions(leagueID, playerID string, seasonWeeks int) ([]Transaction, error)
    // GetAllTransactions filtered to transactions where playerID appears in adds or drops.

// Drafts
func (c *Client) GetUserDrafts(userID, sport, season string) ([]Draft, error)
func (c *Client) GetLeagueDrafts(leagueID string) ([]Draft, error)
func (c *Client) GetDraft(draftID string) (*Draft, error)
func (c *Client) GetDraftPicks(draftID string) ([]DraftPick, error)
func (c *Client) GetTradedDraftPicks(draftID string) ([]TradedPick, error)

// Players
func (c *Client) GetPlayers(sport string) (map[string]Player, error)
func (c *Client) GetTrendingPlayers(sport, trendType string, lookbackHours, limit int) ([]TrendingPlayer, error)

// State
func (c *Client) GetSportState(sport string) (*SportState, error)
```

### Options (`options.go`)
```go
type Option func(*Client)

func WithBaseURL(url string) Option       // for testing
func WithHTTPClient(c *http.Client) Option
func WithCacheDir(dir string) Option      // default: ~/.cache/sleeper/
func WithCacheTTL(d time.Duration) Option // default: 24h
```

## Models (`models.go`)
All structs use `json` tags. Unknown/undocumented fields captured via `Metadata map[string]any` fields where applicable.

```go
type User struct {
    UserID      string `json:"user_id"`
    Username    string `json:"username"`
    DisplayName string `json:"display_name"`
    Avatar      string `json:"avatar"`
}

type League struct {
    LeagueID        string         `json:"league_id"`
    Name            string         `json:"name"`
    Sport           string         `json:"sport"`
    Season          string         `json:"season"`
    Status          string         `json:"status"`
    RosterPositions []string       `json:"roster_positions"`
    Settings        map[string]any `json:"settings"`
    ScoringSettings map[string]any `json:"scoring_settings"`
    Metadata        map[string]any `json:"metadata"`
}

type Roster struct {
    RosterID int      `json:"roster_id"`
    OwnerID  string   `json:"owner_id"`
    Players  []string `json:"players"`
    Starters []string `json:"starters"`
    Reserve  []string `json:"reserve"`
    Settings map[string]any `json:"settings"`
    Metadata map[string]any `json:"metadata"`
}

type Player struct {
    PlayerID           string   `json:"player_id"`
    FirstName          string   `json:"first_name"`
    LastName           string   `json:"last_name"`
    Position           string   `json:"position"`
    Team               string   `json:"team"`
    Age                int      `json:"age"`
    Number             int      `json:"number"`
    Status             string   `json:"status"`
    College            string   `json:"college"`
    YearsExp           int      `json:"years_exp"`
    FantasyPositions   []string `json:"fantasy_positions"`
    InjuryStatus       string   `json:"injury_status"`
    DepthChartPosition int      `json:"depth_chart_position"`
    DepthChartOrder    int      `json:"depth_chart_order"`
    SearchFullName     string   `json:"search_full_name"`
    Extra map[string]any `json:"-"` // populated via custom UnmarshalJSON
}

type TrendingPlayer struct {
    PlayerID string `json:"player_id"`
    Count    int    `json:"count"`
}

type Transaction struct {
    TransactionID string              `json:"transaction_id"`
    Type          string              `json:"type"`          // "trade", "free_agent", "waiver"
    Status        string              `json:"status"`
    Adds          map[string]int      `json:"adds"`          // player_id → roster_id
    Drops         map[string]int      `json:"drops"`         // player_id → roster_id
    DraftPicks    []TradedPick        `json:"draft_picks"`
    Creator       string              `json:"creator"`
    Settings      map[string]any      `json:"settings"`
    Metadata      map[string]any      `json:"metadata"`
}

type Draft struct {
    DraftID        string         `json:"draft_id"`
    LeagueID       string         `json:"league_id"`
    Type           string         `json:"type"`
    Status         string         `json:"status"`
    Season         string         `json:"season"`
    Settings       map[string]any `json:"settings"`
    Metadata       map[string]any `json:"metadata"`
    DraftOrder     map[string]int `json:"draft_order"`
    SlotToRosterID map[string]int `json:"slot_to_roster_id"`
}

type DraftPick struct {
    PlayerID  string         `json:"player_id"`
    PickedBy  string         `json:"picked_by"`
    RosterID  int            `json:"roster_id"`
    Round     int            `json:"round"`
    DraftSlot int            `json:"draft_slot"`
    PickNo    int            `json:"pick_no"`
    Metadata  map[string]any `json:"metadata"`
}

type TradedPick struct {
    Season        string `json:"season"`
    Round         int    `json:"round"`
    RosterID      int    `json:"roster_id"`
    PreviousOwner int    `json:"previous_owner_id"`
    Owner         int    `json:"owner_id"`
}

type Matchup struct {
    RosterID     int      `json:"roster_id"`
    MatchupID    int      `json:"matchup_id"`
    Points       float64  `json:"points"`
    Players      []string `json:"players"`
    Starters     []string `json:"starters"`
    CustomPoints float64  `json:"custom_points"`
}

type BracketMatch struct {
    Round   int `json:"r"`
    MatchID int `json:"m"`
    Team1   int `json:"t1"`
    Team2   int `json:"t2"`
    Winner  int `json:"w"`
    Loser   int `json:"l"`
}

type LeagueUser struct {
    UserID      string         `json:"user_id"`
    Username    string         `json:"username"`
    DisplayName string         `json:"display_name"`
    Avatar      string         `json:"avatar"`
    Metadata    map[string]any `json:"metadata"`
}

type SportState struct {
    Season       string `json:"season"`
    Week         int    `json:"week"`
    SeasonType   string `json:"season_type"`
    LeagueSeason string `json:"league_season"`
    DisplayWeek  int    `json:"display_week"`
}
```

## Player Cache (`cache.go`)
- SQLite database stored at `cacheDir/sleeper_players.db` (default: `~/.cache/sleeper/`)
- Schema: `players` table storing JSON blob per sport + `fetched_at` timestamp
- TTL-based expiry: default 24 hours, configurable
- Methods:
  - `GetPlayers(sport) (map[string]Player, error)` — returns cached data if fresh, else empty
  - `StorePlayers(sport, players)` — writes full player set as JSON blob
  - `GetPlayer(playerID) (*Player, error)` — single player lookup from cache
  - `Clear(sport)` — bust cache for a sport
- Client uses cache transparently: `GetPlayers()` checks cache first, fetches from API if stale

## Implementation Order
1. `go mod init github.com/joeyvforthree/sleeper-go`
2. `models.go` — all structs
3. `options.go` — functional options
4. `client.go` — HTTP client with all API methods
5. `cache.go` — SQLite player cache
6. `client_test.go` + `cache_test.go` — using `net/http/httptest` for deterministic testing
7. `CLAUDE.md`

## Verification
- `go vet ./...` — no issues
- `go build ./...` — compiles cleanly
- `go test ./...` — all tests pass
- Manual smoke test: `GetSportState("nfl")` against live API

## Usage (by consumers)
```go
import sleeper "github.com/joeyvforthree/sleeper-go"

client := sleeper.New(sleeper.WithCacheDir("/tmp/sleeper"))
state, err := client.GetSportState("nfl")
```
