package sleeper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// PlayerCache is a SQLite-backed cache for player data.
type PlayerCache struct {
	db     *sql.DB
	ttl    time.Duration
	logger *slog.Logger
}

// NewPlayerCache opens (or creates) the SQLite cache database at dir/sleeper_players.db.
func NewPlayerCache(dir string, ttl time.Duration, logger *slog.Logger) (*PlayerCache, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cache mkdir: %w", err)
	}

	dbPath := filepath.Join(dir, "sleeper_players.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("cache open: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS players (
		sport      TEXT PRIMARY KEY,
		data       BLOB NOT NULL,
		fetched_at DATETIME NOT NULL
	)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("cache init table: %w", err)
	}

	return &PlayerCache{db: db, ttl: ttl, logger: logger}, nil
}

// GetPlayers returns cached player data for the sport if it exists and is within TTL.
// Returns nil, nil if no fresh data is available.
func (pc *PlayerCache) GetPlayers(sport string) (map[string]Player, error) {
	var data []byte
	var fetchedAt time.Time
	err := pc.db.QueryRow(
		"SELECT data, fetched_at FROM players WHERE sport = ?", sport,
	).Scan(&data, &fetchedAt)
	if err == sql.ErrNoRows {
		pc.logger.Debug("player cache miss", "sport", sport)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache query: %w", err)
	}

	age := time.Since(fetchedAt)
	if age > pc.ttl {
		pc.logger.Debug("player cache expired", "sport", sport, "age", age)
		return nil, nil
	}

	pc.logger.Debug("player cache hit", "sport", sport, "age", age)

	var players map[string]Player
	if err := json.Unmarshal(data, &players); err != nil {
		return nil, fmt.Errorf("cache unmarshal: %w", err)
	}
	return players, nil
}

// StorePlayers writes the full player set for a sport into the cache.
func (pc *PlayerCache) StorePlayers(sport string, players map[string]Player) error {
	data, err := json.Marshal(players)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	_, err = pc.db.Exec(
		`INSERT INTO players (sport, data, fetched_at) VALUES (?, ?, ?)
		 ON CONFLICT(sport) DO UPDATE SET data = excluded.data, fetched_at = excluded.fetched_at`,
		sport, data, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("cache store: %w", err)
	}

	pc.logger.Info("player cache write", "sport", sport, "player_count", len(players), "data_size", len(data))
	return nil
}

// GetPlayer looks up a single player by ID across all cached sports.
func (pc *PlayerCache) GetPlayer(playerID string) (*Player, error) {
	rows, err := pc.db.Query("SELECT data, fetched_at FROM players")
	if err != nil {
		return nil, fmt.Errorf("cache query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var data []byte
		var fetchedAt time.Time
		if err := rows.Scan(&data, &fetchedAt); err != nil {
			return nil, fmt.Errorf("cache scan: %w", err)
		}
		if time.Since(fetchedAt) > pc.ttl {
			continue
		}
		var players map[string]Player
		if err := json.Unmarshal(data, &players); err != nil {
			continue
		}
		if p, ok := players[playerID]; ok {
			return &p, nil
		}
	}
	return nil, nil
}

// Clear removes cached data for a sport.
func (pc *PlayerCache) Clear(sport string) error {
	_, err := pc.db.Exec("DELETE FROM players WHERE sport = ?", sport)
	if err != nil {
		return fmt.Errorf("cache clear: %w", err)
	}
	return nil
}

// Close closes the underlying database connection.
func (pc *PlayerCache) Close() error {
	return pc.db.Close()
}
