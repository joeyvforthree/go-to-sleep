package sleeper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/time/rate"
)

const defaultBaseURL = "https://api.sleeper.app/v1"

// Client is a Sleeper API client with caching, rate limiting, and structured logging.
type Client struct {
	httpClient *http.Client
	baseURL    string
	cacheDir   string
	cacheTTL   time.Duration
	logger     *slog.Logger
	rateRPM    int
	limiter    *rate.Limiter
	cache      *PlayerCache
}

// New creates a new Sleeper API client with the given options.
func New(opts ...Option) *Client {
	c := &Client{
		httpClient: http.DefaultClient,
		baseURL:    defaultBaseURL,
		cacheTTL:   24 * time.Hour,
		logger:     slog.Default(),
		rateRPM:    1000,
	}
	for _, o := range opts {
		o(c)
	}

	// Default cache dir.
	if c.cacheDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			c.cacheDir = filepath.Join(home, ".cache", "sleeper")
		}
	}

	// Initialize rate limiter: rateRPM tokens per minute, burst of 10.
	rps := float64(c.rateRPM) / 60.0
	c.limiter = rate.NewLimiter(rate.Limit(rps), 10)

	return c
}

// get performs a rate-limited GET request and decodes the JSON response into dst.
// It returns ErrNotFound for null JSON bodies.
func (c *Client) get(ctx context.Context, endpoint string, dst any) error {
	c.logger.Debug("rate limit wait", "endpoint", endpoint)
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("request failed", "endpoint", endpoint, "error", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("body read failed", "endpoint", endpoint, "error", err)
		return err
	}
	dur := time.Since(start)

	c.logger.Debug("sleeper request",
		"method", "GET",
		"url", endpoint,
		"status", resp.StatusCode,
		"duration", dur,
	)

	if resp.StatusCode != http.StatusOK {
		truncated := string(body)
		if len(truncated) > 200 {
			truncated = truncated[:200]
		}
		c.logger.Warn("sleeper api error",
			"endpoint", endpoint,
			"status", resp.StatusCode,
			"body", truncated,
		)
		if resp.StatusCode == http.StatusTooManyRequests {
			c.logger.Warn("rate limit hit", "endpoint", endpoint)
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   endpoint,
			Body:       string(body),
		}
	}

	// Sleeper returns literal "null" for missing resources.
	if string(body) == "null" {
		return ErrNotFound
	}

	if err := json.Unmarshal(body, dst); err != nil {
		c.logger.Error("json decode failed", "endpoint", endpoint, "error", err)
		return err
	}
	return nil
}

// --- Users ---

func (c *Client) GetUser(ctx context.Context, identifier string) (*User, error) {
	var u User
	if err := c.get(ctx, "/user/"+identifier, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// --- Leagues ---

func (c *Client) GetUserLeagues(ctx context.Context, userID, sport, season string) ([]League, error) {
	var leagues []League
	endpoint := fmt.Sprintf("/user/%s/leagues/%s/%s", userID, sport, season)
	if err := c.get(ctx, endpoint, &leagues); err != nil {
		return nil, err
	}
	return leagues, nil
}

func (c *Client) GetLeague(ctx context.Context, leagueID string) (*League, error) {
	var l League
	if err := c.get(ctx, "/league/"+leagueID, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

func (c *Client) GetRosters(ctx context.Context, leagueID string) ([]Roster, error) {
	var rosters []Roster
	if err := c.get(ctx, "/league/"+leagueID+"/rosters", &rosters); err != nil {
		return nil, err
	}
	return rosters, nil
}

func (c *Client) GetLeagueUsers(ctx context.Context, leagueID string) ([]LeagueUser, error) {
	var users []LeagueUser
	if err := c.get(ctx, "/league/"+leagueID+"/users", &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (c *Client) GetMatchups(ctx context.Context, leagueID string, week int) ([]Matchup, error) {
	var matchups []Matchup
	endpoint := fmt.Sprintf("/league/%s/matchups/%d", leagueID, week)
	if err := c.get(ctx, endpoint, &matchups); err != nil {
		return nil, err
	}
	return matchups, nil
}

func (c *Client) GetWinnersBracket(ctx context.Context, leagueID string) ([]BracketMatch, error) {
	var bracket []BracketMatch
	if err := c.get(ctx, "/league/"+leagueID+"/winners_bracket", &bracket); err != nil {
		return nil, err
	}
	return bracket, nil
}

func (c *Client) GetLosersBracket(ctx context.Context, leagueID string) ([]BracketMatch, error) {
	var bracket []BracketMatch
	if err := c.get(ctx, "/league/"+leagueID+"/losers_bracket", &bracket); err != nil {
		return nil, err
	}
	return bracket, nil
}

func (c *Client) GetTransactions(ctx context.Context, leagueID string, week int) ([]Transaction, error) {
	var txns []Transaction
	endpoint := fmt.Sprintf("/league/%s/transactions/%d", leagueID, week)
	if err := c.get(ctx, endpoint, &txns); err != nil {
		return nil, err
	}
	return txns, nil
}

func (c *Client) GetTradedPicks(ctx context.Context, leagueID string) ([]TradedPick, error) {
	var picks []TradedPick
	if err := c.get(ctx, "/league/"+leagueID+"/traded_picks", &picks); err != nil {
		return nil, err
	}
	return picks, nil
}

// --- Convenience: Transactions ---

// GetAllTransactions fetches transactions for weeks 1..seasonWeeks.
// If seasonWeeks <= 0, it uses GetSportState("nfl") to determine the current week.
func (c *Client) GetAllTransactions(ctx context.Context, leagueID string, seasonWeeks int) ([]Transaction, error) {
	if seasonWeeks <= 0 {
		state, err := c.GetSportState(ctx, "nfl")
		if err != nil {
			return nil, err
		}
		seasonWeeks = state.Week
	}
	var all []Transaction
	for w := 1; w <= seasonWeeks; w++ {
		c.logger.Info("fetching transactions", "week", w, "total_weeks", seasonWeeks)
		txns, err := c.GetTransactions(ctx, leagueID, w)
		if err != nil {
			return nil, err
		}
		all = append(all, txns...)
	}
	return all, nil
}

// GetPlayerTransactions returns all transactions involving a specific player.
func (c *Client) GetPlayerTransactions(ctx context.Context, leagueID, playerID string, seasonWeeks int) ([]Transaction, error) {
	all, err := c.GetAllTransactions(ctx, leagueID, seasonWeeks)
	if err != nil {
		return nil, err
	}
	var filtered []Transaction
	for _, t := range all {
		if _, ok := t.Adds[playerID]; ok {
			filtered = append(filtered, t)
			continue
		}
		if _, ok := t.Drops[playerID]; ok {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

// --- Drafts ---

func (c *Client) GetUserDrafts(ctx context.Context, userID, sport, season string) ([]Draft, error) {
	var drafts []Draft
	endpoint := fmt.Sprintf("/user/%s/drafts/%s/%s", userID, sport, season)
	if err := c.get(ctx, endpoint, &drafts); err != nil {
		return nil, err
	}
	return drafts, nil
}

func (c *Client) GetLeagueDrafts(ctx context.Context, leagueID string) ([]Draft, error) {
	var drafts []Draft
	if err := c.get(ctx, "/league/"+leagueID+"/drafts", &drafts); err != nil {
		return nil, err
	}
	return drafts, nil
}

func (c *Client) GetDraft(ctx context.Context, draftID string) (*Draft, error) {
	var d Draft
	if err := c.get(ctx, "/draft/"+draftID, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (c *Client) GetDraftPicks(ctx context.Context, draftID string) ([]DraftPick, error) {
	var picks []DraftPick
	if err := c.get(ctx, "/draft/"+draftID+"/picks", &picks); err != nil {
		return nil, err
	}
	return picks, nil
}

func (c *Client) GetTradedDraftPicks(ctx context.Context, draftID string) ([]TradedPick, error) {
	var picks []TradedPick
	if err := c.get(ctx, "/draft/"+draftID+"/traded_picks", &picks); err != nil {
		return nil, err
	}
	return picks, nil
}

// --- Players ---

// GetPlayers fetches all players for a sport. Uses the SQLite cache if available and fresh.
func (c *Client) GetPlayers(ctx context.Context, sport string) (map[string]Player, error) {
	// Try cache first.
	if c.cache == nil && c.cacheDir != "" {
		pc, err := NewPlayerCache(c.cacheDir, c.cacheTTL, c.logger)
		if err != nil {
			c.logger.Error("cache init failed, proceeding without cache", "error", err)
		} else {
			c.cache = pc
		}
	}

	if c.cache != nil {
		players, err := c.cache.GetPlayers(sport)
		if err == nil && players != nil {
			return players, nil
		}
	}

	var players map[string]Player
	if err := c.get(ctx, "/players/"+sport, &players); err != nil {
		return nil, err
	}

	if c.cache != nil {
		if err := c.cache.StorePlayers(sport, players); err != nil {
			c.logger.Error("cache store failed", "error", err)
		}
	}

	return players, nil
}

func (c *Client) GetTrendingPlayers(ctx context.Context, sport, trendType string, lookbackHours, limit int) ([]TrendingPlayer, error) {
	var tp []TrendingPlayer
	endpoint := fmt.Sprintf("/players/%s/trending/%s?lookback_hours=%d&limit=%d",
		sport, trendType, lookbackHours, limit)
	if err := c.get(ctx, endpoint, &tp); err != nil {
		return nil, err
	}
	return tp, nil
}

// --- State ---

func (c *Client) GetSportState(ctx context.Context, sport string) (*SportState, error) {
	var s SportState
	if err := c.get(ctx, "/state/"+sport, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
