//go:build integration

package sleeper

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

const testLeagueID = "1350223992606195712" // PS & The U

func integrationClient(t *testing.T) *Client {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dir := t.TempDir()
	return New(
		WithLogger(logger),
		WithCacheDir(dir),
		WithCacheTTL(1*time.Hour),
	)
}

func TestIntegration_GetSportState(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	state, err := c.GetSportState(ctx, "nfl")
	if err != nil {
		t.Fatalf("GetSportState: %v", err)
	}
	if state.Season == "" {
		t.Error("expected non-empty season")
	}
	if state.SeasonType == "" {
		t.Error("expected non-empty season_type")
	}
	t.Logf("NFL state: season=%s week=%d type=%s", state.Season, state.Week, state.SeasonType)
}

func TestIntegration_GetUser(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	user, err := c.GetUser(ctx, "joeyvfor3")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if user.UserID == "" {
		t.Error("expected non-empty user_id")
	}
	if user.Username != "joeyvfor3" {
		t.Errorf("expected username joeyvfor3, got %s", user.Username)
	}
	t.Logf("User: id=%s username=%s display=%s", user.UserID, user.Username, user.DisplayName)
}

func TestIntegration_GetLeague(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	league, err := c.GetLeague(ctx, testLeagueID)
	if err != nil {
		t.Fatalf("GetLeague: %v", err)
	}
	if league.LeagueID != testLeagueID {
		t.Errorf("expected league_id %s, got %s", testLeagueID, league.LeagueID)
	}
	if league.Name == "" {
		t.Error("expected non-empty league name")
	}
	if league.Sport != "nfl" {
		t.Errorf("expected sport nfl, got %s", league.Sport)
	}
	t.Logf("League: name=%q season=%s status=%s prev_league=%s", league.Name, league.Season, league.Status, league.PreviousLeagueID)
}

func TestIntegration_GetRosters(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	rosters, err := c.GetRosters(ctx, testLeagueID)
	if err != nil {
		t.Fatalf("GetRosters: %v", err)
	}
	if len(rosters) != 12 {
		t.Errorf("expected 12 rosters, got %d", len(rosters))
	}

	// Verify co-owners field is populated on at least one roster
	hasCoOwner := false
	for _, r := range rosters {
		if len(r.CoOwners) > 0 {
			hasCoOwner = true
			t.Logf("Roster %d has co-owners: %v", r.RosterID, r.CoOwners)
		}
	}
	if !hasCoOwner {
		t.Log("warning: no rosters with co-owners found (expected at least 1)")
	}

	t.Logf("Rosters: count=%d", len(rosters))
	for _, r := range rosters {
		t.Logf("  roster=%d owner=%s players=%d", r.RosterID, r.OwnerID, len(r.Players))
	}
}

func TestIntegration_GetLeagueUsers(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	users, err := c.GetLeagueUsers(ctx, testLeagueID)
	if err != nil {
		t.Fatalf("GetLeagueUsers: %v", err)
	}
	if len(users) == 0 {
		t.Error("expected at least 1 league user")
	}

	// Check that metadata contains team_name for at least one user
	hasTeamName := false
	for _, u := range users {
		if tn, ok := u.Metadata["team_name"]; ok {
			hasTeamName = true
			t.Logf("User %s (%s): team_name=%v", u.DisplayName, u.Username, tn)
		}
	}
	if !hasTeamName {
		t.Log("warning: no users with team_name metadata found")
	}

	t.Logf("League users: count=%d", len(users))
}

func TestIntegration_GetPlayers(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// First call — should fetch from API and cache
	t.Log("Fetching players from API (this may take a few seconds)...")
	players, err := c.GetPlayers(ctx, "nfl")
	if err != nil {
		t.Fatalf("GetPlayers: %v", err)
	}
	if len(players) < 1000 {
		t.Errorf("expected 1000+ players, got %d", len(players))
	}
	t.Logf("Players fetched: %d", len(players))

	// Spot check a known player — Patrick Mahomes (player_id 4046)
	if pm, ok := players["4046"]; ok {
		t.Logf("Patrick Mahomes: pos=%s team=%s age=%d exp=%d", pm.Position, pm.Team, pm.Age, pm.YearsExp)
		if pm.Position != "QB" {
			t.Errorf("expected Mahomes position QB, got %s", pm.Position)
		}
	} else {
		t.Log("warning: Patrick Mahomes (4046) not found in player data")
	}

	// Second call — should hit cache
	t.Log("Fetching players again (should hit cache)...")
	players2, err := c.GetPlayers(ctx, "nfl")
	if err != nil {
		t.Fatalf("GetPlayers (cached): %v", err)
	}
	if len(players2) != len(players) {
		t.Errorf("cache returned different count: %d vs %d", len(players2), len(players))
	}
	t.Log("Cache hit confirmed")
}

func TestIntegration_GetTransactions(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	// Fetch week 1 transactions — may be empty depending on league activity
	txns, err := c.GetTransactions(ctx, testLeagueID, 1)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	t.Logf("Week 1 transactions: %d", len(txns))
	for i, tx := range txns {
		if i >= 3 {
			t.Logf("  ... and %d more", len(txns)-3)
			break
		}
		t.Logf("  type=%s status=%s adds=%d drops=%d", tx.Type, tx.Status, len(tx.Adds), len(tx.Drops))
	}
}

func TestIntegration_GetLeagueDrafts(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	drafts, err := c.GetLeagueDrafts(ctx, testLeagueID)
	if err != nil {
		t.Fatalf("GetLeagueDrafts: %v", err)
	}
	if len(drafts) == 0 {
		t.Log("no drafts found for league")
		return
	}
	t.Logf("Drafts: %d", len(drafts))

	// Fetch picks for the first draft
	draft := drafts[0]
	t.Logf("Draft: id=%s type=%s status=%s season=%s", draft.DraftID, draft.Type, draft.Status, draft.Season)

	picks, err := c.GetDraftPicks(ctx, draft.DraftID)
	if err != nil {
		t.Fatalf("GetDraftPicks: %v", err)
	}
	t.Logf("Draft picks: %d", len(picks))
	for i, p := range picks {
		if i >= 5 {
			t.Logf("  ... and %d more", len(picks)-5)
			break
		}
		t.Logf("  round=%d pick=%d player=%s roster=%d", p.Round, p.PickNo, p.PlayerID, p.RosterID)
	}
}
