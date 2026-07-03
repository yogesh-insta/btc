package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ym/btc/internal/config"
)

func TestLoadAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".credentials")
	content := `{
  "oanda": {
    "account_id": "123",
    "token": "abc",
    "environment": "practice"
  }
}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Bot.Instrument != "BTC_USD" {
		t.Fatalf("instrument = %q, want BTC_USD", cfg.Bot.Instrument)
	}
	if !cfg.Monitor.Enabled {
		t.Fatal("expected monitor enabled by default")
	}
	if cfg.Trade.Enabled {
		t.Fatal("expected trade disabled by default")
	}
	if cfg.Trade.MinConfidence != 0.70 {
		t.Fatalf("min confidence = %v, want 0.70", cfg.Trade.MinConfidence)
	}
}
