package oanda

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func FormatPrice(p float64) string {
	return strconv.FormatFloat(p, 'f', 5, 64)
}

func FormatUnits(units float64) string {
	return strconv.FormatFloat(units, 'f', 3, 64)
}

func ParsePrice(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse price %q: %w", s, err)
	}
	return v, nil
}

func ParseTime(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t, err = time.Parse(time.RFC3339, value)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("parse OANDA timestamp %q: %w", value, err)
	}
	return t.UTC(), nil
}

// CalculatePositionUnits sizes a BTC_USD long under ASIC 2:1 leverage (50% margin).
func CalculatePositionUnits(marginAvailable, midPrice, maxLeverage, marginUtilization float64) float64 {
	if marginAvailable <= 0 || midPrice <= 0 {
		return 0
	}

	maxNotional := marginAvailable * maxLeverage * marginUtilization
	rawUnits := maxNotional / midPrice
	return math.Floor(rawUnits*1000) / 1000
}

// PriceWithinTrendZone returns true when close is within ±pct of baseline.
func PriceWithinTrendZone(closePrice, baseline, pct float64) bool {
	if baseline <= 0 {
		return false
	}
	lower := baseline * (1.0 - pct/100.0)
	upper := baseline * (1.0 + pct/100.0)
	return closePrice >= lower && closePrice <= upper
}
