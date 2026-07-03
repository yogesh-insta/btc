package market

import (
	"fmt"

	"github.com/ym/btc/internal/oanda"
)

type Series struct {
	Closes []float64
	Highs  []float64
	Lows   []float64
	Times  []string
}

func ParseCandles(candles []oanda.Candle) (Series, error) {
	var s Series
	for _, c := range candles {
		if !c.Complete && len(candles) > 1 {
			continue
		}
		closeP, err := oanda.ParsePrice(c.Mid.C)
		if err != nil {
			return Series{}, fmt.Errorf("parse close: %w", err)
		}
		highP, err := oanda.ParsePrice(c.Mid.H)
		if err != nil {
			return Series{}, fmt.Errorf("parse high: %w", err)
		}
		lowP, err := oanda.ParsePrice(c.Mid.L)
		if err != nil {
			return Series{}, fmt.Errorf("parse low: %w", err)
		}
		s.Closes = append(s.Closes, closeP)
		s.Highs = append(s.Highs, highP)
		s.Lows = append(s.Lows, lowP)
		s.Times = append(s.Times, c.Time)
	}
	if len(s.Closes) == 0 {
		return Series{}, fmt.Errorf("no complete candles")
	}
	return s, nil
}

func LastN(values []float64, n int) []float64 {
	if n <= 0 || len(values) == 0 {
		return nil
	}
	if len(values) <= n {
		return append([]float64(nil), values...)
	}
	return append([]float64(nil), values[len(values)-n:]...)
}
