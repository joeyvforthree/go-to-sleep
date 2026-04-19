package sleeper

import "encoding/json"

// User represents a Sleeper user account.
type User struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
}

// League represents a fantasy league on Sleeper.
type League struct {
	LeagueID         string         `json:"league_id"`
	PreviousLeagueID string         `json:"previous_league_id"`
	Name             string         `json:"name"`
	Sport            string         `json:"sport"`
	Season           string         `json:"season"`
	Status           string         `json:"status"`
	RosterPositions  []string       `json:"roster_positions"`
	Settings         map[string]any `json:"settings"`
	ScoringSettings  map[string]any `json:"scoring_settings"`
	Metadata         map[string]any `json:"metadata"`
}

// Roster represents a team roster in a league.
type Roster struct {
	RosterID int            `json:"roster_id"`
	OwnerID  string         `json:"owner_id"`
	CoOwners []string       `json:"co_owners"`
	Players  []string       `json:"players"`
	Starters []string       `json:"starters"`
	Reserve  []string       `json:"reserve"`
	Settings map[string]any `json:"settings"`
	Metadata map[string]any `json:"metadata"`
}

// Player represents an NFL (or other sport) player.
// Extra captures any undocumented fields via custom UnmarshalJSON.
type Player struct {
	PlayerID              string   `json:"player_id"`
	FirstName             string   `json:"first_name"`
	LastName              string   `json:"last_name"`
	Position              string   `json:"position"`
	Team                  string   `json:"team"`
	Age                   int      `json:"age"`
	Number                int      `json:"number"`
	Height                string   `json:"height"`
	Weight                string   `json:"weight"`
	Status                string   `json:"status"`
	College               string   `json:"college"`
	BirthCountry          string   `json:"birth_country"`
	YearsExp              int      `json:"years_exp"`
	FantasyPositions      []string `json:"fantasy_positions"`
	InjuryStatus          string   `json:"injury_status"`
	InjuryStartDate       string   `json:"injury_start_date"`
	PracticeParticipation string   `json:"practice_participation"`
	DepthChartPosition    int      `json:"depth_chart_position"`
	DepthChartOrder       int      `json:"depth_chart_order"`
	SearchFullName        string   `json:"search_full_name"`
	SearchFirstName       string   `json:"search_first_name"`
	SearchLastName        string   `json:"search_last_name"`
	SearchRank            int      `json:"search_rank"`
	Hashtag               string   `json:"hashtag"`
	SportradarID          string   `json:"sportradar_id"`
	EspnID                int      `json:"espn_id"`
	YahooID               int      `json:"yahoo_id"`
	RotowireID            int      `json:"rotowire_id"`
	RotoworldID           int      `json:"rotoworld_id"`
	FantasyDataID         int      `json:"fantasy_data_id"`
	StatsID               int      `json:"stats_id"`
	Extra                 map[string]any `json:"-"`
}

// knownPlayerFields is the set of JSON keys that map to explicit Player struct fields.
var knownPlayerFields = map[string]bool{
	"player_id": true, "first_name": true, "last_name": true,
	"position": true, "team": true, "age": true, "number": true,
	"height": true, "weight": true, "status": true, "college": true,
	"birth_country": true, "years_exp": true, "fantasy_positions": true,
	"injury_status": true, "injury_start_date": true, "practice_participation": true,
	"depth_chart_position": true, "depth_chart_order": true,
	"search_full_name": true, "search_first_name": true, "search_last_name": true,
	"search_rank": true, "hashtag": true, "sportradar_id": true,
	"espn_id": true, "yahoo_id": true, "rotowire_id": true, "rotoworld_id": true,
	"fantasy_data_id": true, "stats_id": true,
}

// UnmarshalJSON implements custom unmarshalling for Player.
// Known fields are decoded into struct fields; unknown fields go into Extra.
func (p *Player) UnmarshalJSON(data []byte) error {
	// Alias to avoid infinite recursion.
	type PlayerAlias Player
	var alias PlayerAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*p = Player(alias)

	// Decode all fields into a raw map to find extras.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for k, v := range raw {
		if knownPlayerFields[k] {
			continue
		}
		if p.Extra == nil {
			p.Extra = make(map[string]any)
		}
		var val any
		if err := json.Unmarshal(v, &val); err != nil {
			p.Extra[k] = string(v)
			continue
		}
		p.Extra[k] = val
	}
	return nil
}

// TrendingPlayer represents a trending add/drop.
type TrendingPlayer struct {
	PlayerID string `json:"player_id"`
	Count    int    `json:"count"`
}

// Transaction represents a league transaction (trade, free agent, waiver).
type Transaction struct {
	TransactionID string         `json:"transaction_id"`
	Type          string         `json:"type"`
	Status        string         `json:"status"`
	RosterIDs     []int          `json:"roster_ids"`
	Adds          map[string]int `json:"adds"`
	Drops         map[string]int `json:"drops"`
	DraftPicks    []TradedPick   `json:"draft_picks"`
	Creator       string         `json:"creator"`
	Settings      map[string]any `json:"settings"`
	Metadata      map[string]any `json:"metadata"`
}

// Draft represents a league draft.
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

// DraftPick represents a single pick in a draft.
type DraftPick struct {
	PlayerID  string         `json:"player_id"`
	PickedBy  string         `json:"picked_by"`
	RosterID  int            `json:"roster_id"`
	Round     int            `json:"round"`
	DraftSlot int            `json:"draft_slot"`
	PickNo    int            `json:"pick_no"`
	Metadata  map[string]any `json:"metadata"`
}

// TradedPick represents a draft pick that has been traded.
type TradedPick struct {
	Season        string `json:"season"`
	Round         int    `json:"round"`
	RosterID      int    `json:"roster_id"`
	PreviousOwner int    `json:"previous_owner_id"`
	Owner         int    `json:"owner_id"`
}

// Matchup represents a weekly matchup between two rosters.
type Matchup struct {
	RosterID     int      `json:"roster_id"`
	MatchupID    int      `json:"matchup_id"`
	Points       float64  `json:"points"`
	Players      []string `json:"players"`
	Starters     []string `json:"starters"`
	CustomPoints float64  `json:"custom_points"`
}

// BracketMatch represents a single match in a playoff bracket.
type BracketMatch struct {
	Round   int `json:"r"`
	MatchID int `json:"m"`
	Team1   int `json:"t1"`
	Team2   int `json:"t2"`
	Winner  int `json:"w"`
	Loser   int `json:"l"`
}

// LeagueUser represents a user within a specific league.
type LeagueUser struct {
	UserID      string         `json:"user_id"`
	Username    string         `json:"username"`
	DisplayName string         `json:"display_name"`
	Avatar      string         `json:"avatar"`
	Metadata    map[string]any `json:"metadata"`
}

// SportState represents the current state of a sport on Sleeper.
type SportState struct {
	Season       string `json:"season"`
	Week         int    `json:"week"`
	SeasonType   string `json:"season_type"`
	LeagueSeason string `json:"league_season"`
	DisplayWeek  int    `json:"display_week"`
}
