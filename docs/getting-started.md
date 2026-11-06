---
title: Getting started
description: Install go-to-sleep and make your first Sleeper API calls.
---

go-to-sleep is a Go client for the [Sleeper](https://sleeper.com) fantasy sports API. It
returns typed structs for every endpoint, caches the large player dump in SQLite, rate limits
itself, and logs through `log/slog`.

The Sleeper API is read-only and needs no API key.

## Install

Requires Go 1.23 or later.

```bash
go get github.com/joeyvforthree/go-to-sleep
```

The package name is `sleeper`:

```go
import sleeper "github.com/joeyvforthree/go-to-sleep"
```

## First calls

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

	state, err := client.GetSportState(ctx, "nfl")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("NFL %s, week %d (%s)\n", state.Season, state.Week, state.SeasonType)

	user, err := client.GetUser(ctx, "YOUR_USERNAME")
	if err != nil {
		log.Fatal(err)
	}

	leagues, err := client.GetUserLeagues(ctx, user.UserID, "nfl", state.Season)
	if err != nil {
		log.Fatal(err)
	}
	for _, l := range leagues {
		fmt.Printf("%s (%s)\n", l.Name, l.LeagueID)
	}
}
```

Every method takes a `context.Context` first, so deadlines and cancellation work the usual
way. Create one client and reuse it: the rate limiter lives on the client, so separate
clients do not share a limit.

## Next

- [Configuration](../configuration/): cache location and TTL, rate limit, logger, HTTP client.
- [API reference](../api-reference/): every method and the types it returns.
- [Examples](../examples/): common jobs end to end.
