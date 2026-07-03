package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ym/btc/internal/pattern"
)

type Snapshot struct {
	SavedAt         time.Time         `json:"saved_at"`
	LastMonitorDate string            `json:"last_monitor_date,omitempty"`
	LastTradeDate   string            `json:"last_trade_date,omitempty"`
	TradeWeek       string            `json:"trade_week,omitempty"`
	TradesThisWeek  int               `json:"trades_this_week"`
	SentAlertSlots  []string          `json:"sent_alert_slots,omitempty"`
	LastAnalysis    *pattern.Analysis `json:"last_analysis,omitempty"`
}

type Store struct {
	path string
}

func NewStore(path string) *Store {
	if path == "" {
		path = "data/state.json"
	}
	return &Store{path: path}
}

func (s *Store) Load() (Snapshot, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return Snapshot{}, nil
		}
		return Snapshot{}, err
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, fmt.Errorf("parse state: %w", err)
	}
	return snap, nil
}

func (s *Store) Save(snap Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	snap.SavedAt = time.Now().UTC()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
