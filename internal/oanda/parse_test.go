package oanda_test

import (
	"testing"

	"github.com/ym/btc/internal/oanda"
)

func TestPriceWithinTrendZone(t *testing.T) {
	baseline := 100_000.0
	zone := 5.0

	tests := []struct {
		name  string
		close float64
		want  bool
	}{
		{"at baseline", 100_000, true},
		{"lower bound", 95_000, true},
		{"upper bound", 105_000, true},
		{"below zone", 94_999, false},
		{"above zone", 105_001, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := oanda.PriceWithinTrendZone(tt.close, baseline, zone)
			if got != tt.want {
				t.Fatalf("PriceWithinTrendZone(%v, %v, %v) = %v, want %v", tt.close, baseline, zone, got, tt.want)
			}
		})
	}
}

func TestCalculatePositionUnits(t *testing.T) {
	units := oanda.CalculatePositionUnits(10_000, 100_000, 2.0, 0.95)
	want := 0.190
	if units != want {
		t.Fatalf("CalculatePositionUnits() = %v, want %v", units, want)
	}

	if got := oanda.CalculatePositionUnits(0, 100_000, 2.0, 0.95); got != 0 {
		t.Fatalf("zero margin should return 0, got %v", got)
	}
}
