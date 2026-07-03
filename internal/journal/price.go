package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ym/btc/internal/pattern"
)

type PriceWriter struct {
	dir string
}

func NewPriceWriter(dir string) *PriceWriter {
	if dir == "" {
		dir = "logs/price"
	}
	return &PriceWriter{dir: dir}
}

func (w *PriceWriter) Append(a pattern.Analysis) error {
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}
	day := a.Date
	if day == "" {
		day = time.Now().UTC().Format("2006-01-02")
	}
	path := filepath.Join(w.dir, day+".jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

type TradeWriter struct {
	dir string
}

func NewTradeWriter(dir string) *TradeWriter {
	if dir == "" {
		dir = "logs/trades"
	}
	return &TradeWriter{dir: dir}
}

type TradeEntry struct {
	At         time.Time      `json:"at"`
	Action     string         `json:"action"`
	Reason     string         `json:"reason,omitempty"`
	BullScore  float64        `json:"bull_score,omitempty"`
	Confidence string         `json:"confidence,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

func (w *TradeWriter) Append(e TradeEntry) error {
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(w.dir, "journal.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}
