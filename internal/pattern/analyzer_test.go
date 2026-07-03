package pattern

import (
	"testing"

	"github.com/ym/btc/internal/market"
)

func TestAnalyzeBullScore(t *testing.T) {
	closes := make([]float64, 220)
	for i := range closes {
		closes[i] = 90_000 + float64(i)*120
	}
	series := market.Series{Closes: closes, Highs: closes, Lows: closes}

	a := Analyze(series, 55, DefaultConfig(), 0.70)
	if a.BullScore <= 0 {
		t.Fatalf("expected positive bull score, got %v", a.BullScore)
	}
	if len(a.Patterns) == 0 {
		t.Fatal("expected at least one active pattern")
	}
}

func TestAnalyzeLowConfidence(t *testing.T) {
	closes := []float64{100, 99, 98, 97, 96, 95, 94, 93, 92, 91, 90, 89, 88, 87, 86}
	series := market.Series{Closes: closes, Highs: closes, Lows: closes}

	a := Analyze(series, 10, DefaultConfig(), 0.70)
	if a.TradeReady {
		t.Fatalf("expected trade not ready, score=%v patterns=%v", a.BullScore, a.Patterns)
	}
}
