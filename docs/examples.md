---
title: Examples
description: Standings, trade history and trending players with go-to-sleep.
---

Each example assumes a `client := sleeper.New()`, a `ctx`, and a `leagueID` you own or can
see. The IDs are placeholders.

## Standings with team names

Rosters carry the owner's user ID; the name lives on the league's users. Wins and points are
in the roster's open-ended `Settings`, where JSON numbers decode as `float64`.

```go
rosters, err := client.GetRosters(ctx, leagueID)
if err != nil {
	return err
}
users, err := client.GetLeagueUsers(ctx, leagueID)
if err != nil {
	return err
}

names := make(map[string]string, len(users))
for _, u := range users {
	name := u.DisplayName
	if team, ok := u.Metadata["team_name"].(string); ok && team != "" {
		name = team
	}
	names[u.UserID] = name
}

num := func(m map[string]any, k string) float64 {
	v, _ := m[k].(float64)
	return v
}
sort.Slice(rosters, func(i, j int) bool {
	return num(rosters[i].Settings, "wins") > num(rosters[j].Settings, "wins")
})
for _, r := range rosters {
	fmt.Printf("%-24s %2.0f-%-2.0f %7.0f pts\n", names[r.OwnerID],
		num(r.Settings, "wins"), num(r.Settings, "losses"), num(r.Settings, "fpts"))
}
```

## A season of trades, with player names

`GetAllTransactions` fetches every week; `GetPlayers` turns IDs into names, from the cache
after the first call.

```go
txns, err := client.GetAllTransactions(ctx, leagueID, 0) // 0: through the current week
if err != nil {
	return err
}
players, err := client.GetPlayers(ctx, "nfl")
if err != nil {
	return err
}

name := func(id string) string {
	if p, ok := players[id]; ok {
		return p.FirstName + " " + p.LastName
	}
	return id // team defenses and retired players can be missing
}

for _, t := range txns {
	if t.Type != "trade" || t.Status != "complete" {
		continue
	}
	for id, roster := range t.Adds {
		fmt.Printf("roster %d gets %s\n", roster, name(id))
	}
	for _, pk := range t.DraftPicks {
		fmt.Printf("roster %d gets %s round %d pick (from roster %d)\n",
			pk.Owner, pk.Season, pk.Round, pk.PreviousOwner)
	}
	fmt.Println("---")
}
```

## Trending waiver adds

```go
trending, err := client.GetTrendingPlayers(ctx, "nfl", "add", 24, 10)
if err != nil {
	return err
}
players, err := client.GetPlayers(ctx, "nfl")
if err != nil {
	return err
}
for _, t := range trending {
	p := players[t.PlayerID]
	fmt.Printf("%-24s %-3s %-3s added %d times\n",
		p.FirstName+" "+p.LastName, p.Position, p.Team, t.Count)
}
```

## A deadline on every call

```go
client := sleeper.New(sleeper.WithHTTPClient(&http.Client{Timeout: 20 * time.Second}))

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()
txns, err := client.GetAllTransactions(ctx, leagueID, 18) // the whole fetch shares the 2 minutes
```

## Testing your code against a fake Sleeper

`WithBaseURL` points the client at an `httptest.Server`, so tests never touch the real API.
Give the test its own cache directory so it never reads a real cache.

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/league/123" {
		fmt.Fprint(w, `{"league_id":"123","name":"Test League","season":"2026"}`)
		return
	}
	fmt.Fprint(w, "null") // Sleeper's not-found answer
}))
defer srv.Close()

client := sleeper.New(sleeper.WithBaseURL(srv.URL), sleeper.WithCacheDir(t.TempDir()))
league, err := client.GetLeague(context.Background(), "123")
```
