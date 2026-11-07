---
title: API reference
description: Every Client method, the Sleeper endpoint behind it, and the types it returns.
---

All methods are on `*sleeper.Client`, take a `context.Context` first, and return an error
last. IDs are strings, as Sleeper sends them. `sport` is `"nfl"`. `season` is a year string such as `"2026"`.

## Users

| Method | Endpoint | Returns |
|--------|----------|---------|
| `GetUser(ctx, usernameOrID)` | `/user/{id}` | `*User` |

## Leagues

| Method | Endpoint | Returns |
|--------|----------|---------|
| `GetUserLeagues(ctx, userID, sport, season)` | `/user/{id}/leagues/{sport}/{season}` | `[]League` |
| `GetLeague(ctx, leagueID)` | `/league/{id}` | `*League` |
| `GetRosters(ctx, leagueID)` | `/league/{id}/rosters` | `[]Roster` |
| `GetLeagueUsers(ctx, leagueID)` | `/league/{id}/users` | `[]LeagueUser` |
| `GetMatchups(ctx, leagueID, week)` | `/league/{id}/matchups/{week}` | `[]Matchup` |
| `GetWinnersBracket(ctx, leagueID)` | `/league/{id}/winners_bracket` | `[]BracketMatch` |
| `GetLosersBracket(ctx, leagueID)` | `/league/{id}/losers_bracket` | `[]BracketMatch` |
| `GetTradedPicks(ctx, leagueID)` | `/league/{id}/traded_picks` | `[]TradedPick` |

## Transactions

| Method | Endpoint | Returns |
|--------|----------|---------|
| `GetTransactions(ctx, leagueID, week)` | `/league/{id}/transactions/{week}` | `[]Transaction` |
| `GetAllTransactions(ctx, leagueID, seasonWeeks)` | one call per week | `[]Transaction` |
| `GetPlayerTransactions(ctx, leagueID, playerID, seasonWeeks)` | one call per week | `[]Transaction` |

`GetAllTransactions` fetches weeks 1 through `seasonWeeks` in order and concatenates them. With
`seasonWeeks <= 0` it asks `GetSportState("nfl")` for the current week and stops there. It
returns the first error it meets, with no partial result.

`GetPlayerTransactions` runs `GetAllTransactions` and keeps the transactions whose `Adds` or
`Drops` contain the player.

## Drafts

| Method | Endpoint | Returns |
|--------|----------|---------|
| `GetUserDrafts(ctx, userID, sport, season)` | `/user/{id}/drafts/{sport}/{season}` | `[]Draft` |
| `GetLeagueDrafts(ctx, leagueID)` | `/league/{id}/drafts` | `[]Draft` |
| `GetDraft(ctx, draftID)` | `/draft/{id}` | `*Draft` |
| `GetDraftPicks(ctx, draftID)` | `/draft/{id}/picks` | `[]DraftPick` |
| `GetTradedDraftPicks(ctx, draftID)` | `/draft/{id}/traded_picks` | `[]TradedPick` |

## Players

| Method | Endpoint | Returns |
|--------|----------|---------|
| `GetPlayers(ctx, sport)` | `/players/{sport}` | `map[string]Player`, keyed by player ID |
| `GetTrendingPlayers(ctx, sport, trendType, lookbackHours, limit)` | `/players/{sport}/trending/{type}` | `[]TrendingPlayer` |

`GetPlayers` is served from the [player cache](../configuration/#player-cache) when it is fresh.
`trendType` is `"add"` or `"drop"`.

## Sport state

| Method | Endpoint | Returns |
|--------|----------|---------|
| `GetSportState(ctx, sport)` | `/state/{sport}` | `*SportState` |

## Errors

A non-2xx response comes back as `*sleeper.APIError`, which wraps a sentinel you can test with
`errors.Is`:

| Sentinel | When |
|----------|------|
| `ErrNotFound` | 404, or a 200 whose body is `null` (Sleeper's answer for an unknown ID) |
| `ErrRateLimited` | 429 |
| `ErrServerError` | any 5xx |

```go
league, err := client.GetLeague(ctx, leagueID)
switch {
case errors.Is(err, sleeper.ErrNotFound):
	// no such league
case errors.Is(err, sleeper.ErrRateLimited):
	// back off and retry later
case err != nil:
	var apiErr *sleeper.APIError
	if errors.As(err, &apiErr) {
		log.Printf("status %d on %s", apiErr.StatusCode, apiErr.Endpoint)
	}
	return err
}
```

`APIError` has `StatusCode`, `Endpoint` (the path, without the host) and `Body`. A `null`
body returns `ErrNotFound` itself, not an `APIError`. Network, context and JSON errors are
returned as they are. The client does not retry.

## Types

Fields map one to one onto Sleeper's JSON. Fields Sleeper leaves open-ended (`Settings`,
`ScoringSettings`, `Metadata`) are `map[string]any`.

| Type | Main fields |
|------|-------------|
| `User` | `UserID`, `Username`, `DisplayName`, `Avatar` |
| `LeagueUser` | `User` fields plus `Metadata` (team name lives here) |
| `League` | `LeagueID`, `PreviousLeagueID`, `Name`, `Sport`, `Season`, `Status`, `RosterPositions`, `Settings`, `ScoringSettings`, `Metadata` |
| `Roster` | `RosterID`, `OwnerID`, `CoOwners`, `Players`, `Starters`, `Reserve`, `Settings`, `Metadata` |
| `Matchup` | `RosterID`, `MatchupID`, `Points`, `CustomPoints`, `Players`, `Starters` |
| `BracketMatch` | `Round`, `MatchID`, `Team1`, `Team2`, `Winner`, `Loser` (roster IDs) |
| `Transaction` | `TransactionID`, `Type`, `Status`, `RosterIDs`, `Adds`, `Drops` (player ID → roster ID), `DraftPicks`, `Creator`, `Settings`, `Metadata` |
| `Draft` | `DraftID`, `LeagueID`, `Type`, `Status`, `Season`, `DraftOrder`, `SlotToRosterID`, `Settings`, `Metadata` |
| `DraftPick` | `PlayerID`, `PickedBy`, `RosterID`, `Round`, `DraftSlot`, `PickNo`, `Metadata` |
| `TradedPick` | `Season`, `Round`, `RosterID` (original owner), `PreviousOwner`, `Owner` |
| `TrendingPlayer` | `PlayerID`, `Count` |
| `SportState` | `Season`, `Week`, `SeasonType`, `LeagueSeason`, `DisplayWeek` |

### Player

`Player` has typed fields for identity (`PlayerID`, `FirstName`, `LastName`, `Position`,
`FantasyPositions`, `Team`, `Number`), biography (`Age`, `Height`, `Weight`, `College`,
`BirthCountry`, `YearsExp`), availability (`Status`, `InjuryStatus`, `InjuryStartDate`,
`PracticeParticipation`, `DepthChartPosition`, `DepthChartOrder`), search (`SearchFullName`,
`SearchRank`, and others) and cross-site IDs (`SportradarID`, `EspnID`, `YahooID`,
`RotowireID`, `RotoworldID`, `FantasyDataID`, `StatsID`).

Sleeper adds player fields without notice. Any key the struct does not model lands in
`Player.Extra` (`map[string]any`) instead of being dropped, so new data is reachable before the
library catches up. `Extra` is kept in the player cache too.
