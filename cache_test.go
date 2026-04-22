package sleeper

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func testCache(t *testing.T, ttl time.Duration) *PlayerCache {
	t.Helper()
	dir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	pc, err := NewPlayerCache(dir, ttl, logger)
	if err != nil {
		t.Fatalf("NewPlayerCache: %v", err)
	}
	t.Cleanup(func() { pc.Close() })
	return pc
}

func TestCacheStoreAndGet(t *testing.T) {
	pc := testCache(t, time.Hour)

	players := map[string]Player{
		"p1": {PlayerID: "p1", FirstName: "Patrick", LastName: "Mahomes"},
		"p2": {PlayerID: "p2", FirstName: "Josh", LastName: "Allen"},
	}

	if err := pc.StorePlayers("nfl", players); err != nil {
		t.Fatalf("StorePlayers: %v", err)
	}

	got, err := pc.GetPlayers("nfl")
	if err != nil {
		t.Fatalf("GetPlayers: %v", err)
	}
	if got == nil {
		t.Fatal("GetPlayers returned nil, expected data")
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got["p1"].FirstName != "Patrick" {
		t.Errorf("p1 FirstName = %q, want %q", got["p1"].FirstName, "Patrick")
	}
}

func TestCacheMiss(t *testing.T) {
	pc := testCache(t, time.Hour)

	got, err := pc.GetPlayers("nfl")
	if err != nil {
		t.Fatalf("GetPlayers: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for cache miss, got %v", got)
	}
}

func TestCacheExpiry(t *testing.T) {
	// Use a TTL of 0 so data is always expired.
	pc := testCache(t, 0)

	players := map[string]Player{
		"p1": {PlayerID: "p1", FirstName: "Test"},
	}
	if err := pc.StorePlayers("nfl", players); err != nil {
		t.Fatalf("StorePlayers: %v", err)
	}

	got, err := pc.GetPlayers("nfl")
	if err != nil {
		t.Fatalf("GetPlayers: %v", err)
	}
	if got != nil {
		t.Error("expected nil for expired cache, got data")
	}
}

func TestCacheGetPlayer(t *testing.T) {
	pc := testCache(t, time.Hour)

	players := map[string]Player{
		"p1": {PlayerID: "p1", FirstName: "Patrick", LastName: "Mahomes"},
	}
	if err := pc.StorePlayers("nfl", players); err != nil {
		t.Fatalf("StorePlayers: %v", err)
	}

	p, err := pc.GetPlayer("p1")
	if err != nil {
		t.Fatalf("GetPlayer: %v", err)
	}
	if p == nil {
		t.Fatal("GetPlayer returned nil")
	}
	if p.LastName != "Mahomes" {
		t.Errorf("LastName = %q, want %q", p.LastName, "Mahomes")
	}

	// Missing player.
	p, err = pc.GetPlayer("missing")
	if err != nil {
		t.Fatalf("GetPlayer: %v", err)
	}
	if p != nil {
		t.Errorf("expected nil for missing player, got %v", p)
	}
}

func TestCacheClear(t *testing.T) {
	pc := testCache(t, time.Hour)

	players := map[string]Player{
		"p1": {PlayerID: "p1"},
	}
	if err := pc.StorePlayers("nfl", players); err != nil {
		t.Fatalf("StorePlayers: %v", err)
	}

	if err := pc.Clear("nfl"); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	got, err := pc.GetPlayers("nfl")
	if err != nil {
		t.Fatalf("GetPlayers: %v", err)
	}
	if got != nil {
		t.Error("expected nil after Clear, got data")
	}
}

func TestCacheOverwrite(t *testing.T) {
	pc := testCache(t, time.Hour)

	v1 := map[string]Player{"p1": {PlayerID: "p1", FirstName: "Old"}}
	if err := pc.StorePlayers("nfl", v1); err != nil {
		t.Fatalf("StorePlayers v1: %v", err)
	}

	v2 := map[string]Player{"p1": {PlayerID: "p1", FirstName: "New"}}
	if err := pc.StorePlayers("nfl", v2); err != nil {
		t.Fatalf("StorePlayers v2: %v", err)
	}

	got, err := pc.GetPlayers("nfl")
	if err != nil {
		t.Fatalf("GetPlayers: %v", err)
	}
	if got["p1"].FirstName != "New" {
		t.Errorf("FirstName = %q, want %q", got["p1"].FirstName, "New")
	}
}
