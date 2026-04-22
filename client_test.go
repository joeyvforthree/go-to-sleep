package sleeper

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(t *testing.T, routes map[string]any) (*httptest.Server, *Client) {
	t.Helper()
	mux := http.NewServeMux()
	for path, resp := range routes {
		body, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("marshal fixture: %v", err)
		}
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(body)
		})
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := New(WithBaseURL(srv.URL), WithRateLimit(10000))
	return srv, c
}

func TestGetUser(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/user/testuser": User{
			UserID:      "123",
			Username:    "testuser",
			DisplayName: "Test User",
			Avatar:      "abc",
		},
	})

	u, err := c.GetUser(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.UserID != "123" {
		t.Errorf("UserID = %q, want %q", u.UserID, "123")
	}
	if u.Username != "testuser" {
		t.Errorf("Username = %q, want %q", u.Username, "testuser")
	}
}

func TestGetUser_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/nobody", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("null"))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(WithBaseURL(srv.URL), WithRateLimit(10000))
	_, err := c.GetUser(context.Background(), "nobody")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetUser_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/badreq", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(WithBaseURL(srv.URL), WithRateLimit(10000))
	_, err := c.GetUser(context.Background(), "badreq")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
	if !errors.Is(err, ErrServerError) {
		t.Error("expected errors.Is(err, ErrServerError) to be true")
	}
}

func TestGetUser_RateLimited(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/rl", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`"rate limited"`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(WithBaseURL(srv.URL), WithRateLimit(10000))
	_, err := c.GetUser(context.Background(), "rl")
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}

func TestGetLeague(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/league/999": League{
			LeagueID: "999",
			Name:     "Test League",
			Sport:    "nfl",
			Season:   "2024",
		},
	})

	l, err := c.GetLeague(context.Background(), "999")
	if err != nil {
		t.Fatalf("GetLeague: %v", err)
	}
	if l.Name != "Test League" {
		t.Errorf("Name = %q, want %q", l.Name, "Test League")
	}
}

func TestGetRosters(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/league/999/rosters": []Roster{
			{RosterID: 1, OwnerID: "abc", Players: []string{"p1", "p2"}},
		},
	})

	rosters, err := c.GetRosters(context.Background(), "999")
	if err != nil {
		t.Fatalf("GetRosters: %v", err)
	}
	if len(rosters) != 1 {
		t.Fatalf("len = %d, want 1", len(rosters))
	}
	if rosters[0].OwnerID != "abc" {
		t.Errorf("OwnerID = %q, want %q", rosters[0].OwnerID, "abc")
	}
}

func TestGetMatchups(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/league/999/matchups/1": []Matchup{
			{RosterID: 1, MatchupID: 1, Points: 100.5},
			{RosterID: 2, MatchupID: 1, Points: 95.2},
		},
	})

	matchups, err := c.GetMatchups(context.Background(), "999", 1)
	if err != nil {
		t.Fatalf("GetMatchups: %v", err)
	}
	if len(matchups) != 2 {
		t.Fatalf("len = %d, want 2", len(matchups))
	}
}

func TestGetTransactions(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/league/999/transactions/1": []Transaction{
			{TransactionID: "tx1", Type: "free_agent", Adds: map[string]int{"p1": 1}},
		},
	})

	txns, err := c.GetTransactions(context.Background(), "999", 1)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("len = %d, want 1", len(txns))
	}
	if txns[0].Type != "free_agent" {
		t.Errorf("Type = %q, want %q", txns[0].Type, "free_agent")
	}
}

func TestGetDraft(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/draft/d1": Draft{DraftID: "d1", LeagueID: "999", Type: "snake"},
	})

	d, err := c.GetDraft(context.Background(), "d1")
	if err != nil {
		t.Fatalf("GetDraft: %v", err)
	}
	if d.Type != "snake" {
		t.Errorf("Type = %q, want %q", d.Type, "snake")
	}
}

func TestGetDraftPicks(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/draft/d1/picks": []DraftPick{
			{PlayerID: "p1", Round: 1, PickNo: 1},
		},
	})

	picks, err := c.GetDraftPicks(context.Background(), "d1")
	if err != nil {
		t.Fatalf("GetDraftPicks: %v", err)
	}
	if len(picks) != 1 {
		t.Fatalf("len = %d, want 1", len(picks))
	}
}

func TestGetSportState(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/state/nfl": SportState{Season: "2024", Week: 5, SeasonType: "regular"},
	})

	s, err := c.GetSportState(context.Background(), "nfl")
	if err != nil {
		t.Fatalf("GetSportState: %v", err)
	}
	if s.Week != 5 {
		t.Errorf("Week = %d, want 5", s.Week)
	}
}

func TestGetTrendingPlayers(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/players/nfl/trending/add": []TrendingPlayer{
			{PlayerID: "p1", Count: 100},
		},
	})

	// The query params won't match this route exactly via the default mux,
	// but ServeMux matches the path prefix before '?'.
	tp, err := c.GetTrendingPlayers(context.Background(), "nfl", "add", 24, 10)
	if err != nil {
		t.Fatalf("GetTrendingPlayers: %v", err)
	}
	if len(tp) != 1 {
		t.Fatalf("len = %d, want 1", len(tp))
	}
}

func TestGetAllTransactions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/league/999/transactions/1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Transaction{{TransactionID: "t1", Type: "trade"}})
	})
	mux.HandleFunc("/league/999/transactions/2", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Transaction{{TransactionID: "t2", Type: "free_agent"}})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(WithBaseURL(srv.URL), WithRateLimit(10000))
	txns, err := c.GetAllTransactions(context.Background(), "999", 2)
	if err != nil {
		t.Fatalf("GetAllTransactions: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("len = %d, want 2", len(txns))
	}
}

func TestGetPlayerTransactions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/league/999/transactions/1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Transaction{
			{TransactionID: "t1", Adds: map[string]int{"p1": 1}},
			{TransactionID: "t2", Drops: map[string]int{"p2": 1}},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(WithBaseURL(srv.URL), WithRateLimit(10000))
	txns, err := c.GetPlayerTransactions(context.Background(), "999", "p1", 1)
	if err != nil {
		t.Fatalf("GetPlayerTransactions: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("len = %d, want 1", len(txns))
	}
	if txns[0].TransactionID != "t1" {
		t.Errorf("TransactionID = %q, want %q", txns[0].TransactionID, "t1")
	}
}

func TestPlayerUnmarshalJSON_Extra(t *testing.T) {
	raw := `{
		"player_id": "123",
		"first_name": "Patrick",
		"last_name": "Mahomes",
		"position": "QB",
		"team": "KC",
		"nickname": "Showtime",
		"some_unknown": 42
	}`

	var p Player
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if p.PlayerID != "123" {
		t.Errorf("PlayerID = %q, want %q", p.PlayerID, "123")
	}
	if p.FirstName != "Patrick" {
		t.Errorf("FirstName = %q, want %q", p.FirstName, "Patrick")
	}
	if p.Extra == nil {
		t.Fatal("Extra is nil, expected populated map")
	}
	if p.Extra["nickname"] != "Showtime" {
		t.Errorf("Extra[nickname] = %v, want %q", p.Extra["nickname"], "Showtime")
	}
	if p.Extra["some_unknown"] != float64(42) {
		t.Errorf("Extra[some_unknown] = %v, want 42", p.Extra["some_unknown"])
	}
	// Known fields should not be in Extra.
	if _, ok := p.Extra["player_id"]; ok {
		t.Error("Extra should not contain known field player_id")
	}
}

func TestGetLeagueUsers(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/league/999/users": []LeagueUser{
			{UserID: "u1", Username: "user1"},
		},
	})

	users, err := c.GetLeagueUsers(context.Background(), "999")
	if err != nil {
		t.Fatalf("GetLeagueUsers: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("len = %d, want 1", len(users))
	}
}

func TestGetWinnersBracket(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/league/999/winners_bracket": []BracketMatch{
			{Round: 1, MatchID: 1, Team1: 1, Team2: 2, Winner: 1, Loser: 2},
		},
	})

	bracket, err := c.GetWinnersBracket(context.Background(), "999")
	if err != nil {
		t.Fatalf("GetWinnersBracket: %v", err)
	}
	if len(bracket) != 1 {
		t.Fatalf("len = %d, want 1", len(bracket))
	}
	if bracket[0].Winner != 1 {
		t.Errorf("Winner = %d, want 1", bracket[0].Winner)
	}
}

func TestGetTradedPicks(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/league/999/traded_picks": []TradedPick{
			{Season: "2024", Round: 1, RosterID: 1, PreviousOwner: 2, Owner: 1},
		},
	})

	picks, err := c.GetTradedPicks(context.Background(), "999")
	if err != nil {
		t.Fatalf("GetTradedPicks: %v", err)
	}
	if len(picks) != 1 {
		t.Fatalf("len = %d, want 1", len(picks))
	}
}

func TestContextCancellation(t *testing.T) {
	_, c := testServer(t, map[string]any{
		"/state/nfl": SportState{Season: "2024", Week: 1},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := c.GetSportState(ctx, "nfl")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
