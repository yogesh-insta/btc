package pattern

import (
	"sort"
	"time"

	"github.com/ym/btc/internal/market"
)

type Signal struct {
	Name   string  `json:"name"`
	Active bool    `json:"active"`
	Weight float64 `json:"weight"`
	Detail string  `json:"detail,omitempty"`
}

type Analysis struct {
	At          time.Time `json:"at"`
	Date        string    `json:"date"`
	Close       float64   `json:"close"`
	SMA20       float64   `json:"sma20"`
	SMA50       float64   `json:"sma50"`
	SMA200      float64   `json:"sma200"`
	RSI14       float64   `json:"rsi14"`
	ROC7d       float64   `json:"roc_7d_pct"`
	High20d     float64   `json:"high_20d"`
	FearGreed   int       `json:"fear_greed"`
	Patterns    []string  `json:"patterns"`
	Signals     []Signal  `json:"signals"`
	BullScore   float64   `json:"bull_score"`
	TradeReady  bool      `json:"trade_ready"`
	TradeReason string    `json:"trade_reason,omitempty"`
}

type Config struct {
	MinTradeConfidence  float64
	Momentum7dPct       float64
	BreakoutNearHighPct float64
	RSIBullMin          float64
	RSIBullMax          float64
	FearGreedBullMin    int
}

func DefaultConfig() Config {
	return Config{
		MinTradeConfidence:  0.70,
		Momentum7dPct:       2.0,
		BreakoutNearHighPct: 0.99,
		RSIBullMin:          50,
		RSIBullMax:          72,
		FearGreedBullMin:    35,
	}
}

func Analyze(series market.Series, fearGreed int, cfg Config, minConfidence float64) Analysis {
	closes := series.Closes
	last := closes[len(closes)-1]

	a := Analysis{
		At:        time.Now().UTC(),
		Date:      time.Now().UTC().Format("2006-01-02"),
		Close:     last,
		SMA20:     market.SMA(closes, 20),
		SMA50:     market.SMA(closes, 50),
		SMA200:    market.SMA(closes, 200),
		RSI14:     market.RSI(closes, 14),
		High20d:   market.MaxOf(market.LastN(closes, 20)),
		FearGreed: fearGreed,
	}

	if len(closes) >= 8 {
		a.ROC7d = market.PctChange(closes[len(closes)-8], last)
	}

	signals := []Signal{
		signal("above_sma20", last > a.SMA20 && a.SMA20 > 0, 0.10, "price above 20-day SMA"),
		signal("above_sma50", last > a.SMA50 && a.SMA50 > 0, 0.15, "price above 50-day SMA"),
		signal("above_sma200", last > a.SMA200 && a.SMA200 > 0, 0.15, "price above 200-day SMA"),
		signal("golden_cross", a.SMA50 > a.SMA200 && a.SMA50 > 0 && a.SMA200 > 0, 0.15, "50-day SMA above 200-day SMA"),
		signal("higher_lows", market.HigherLows(market.LocalLows(closes, 2), 3), 0.10, "ascending swing lows"),
		signal("breakout_20d", last >= a.High20d*cfg.BreakoutNearHighPct, 0.15, "near or above 20-day high"),
		signal("momentum_7d", a.ROC7d >= cfg.Momentum7dPct, 0.10, "7-day momentum >= 2%"),
		signal("rsi_bull_zone", a.RSI14 >= cfg.RSIBullMin && a.RSI14 <= cfg.RSIBullMax, 0.05, "RSI in bullish zone"),
		signal("sentiment_ok", fearGreed >= cfg.FearGreedBullMin, 0.05, "Fear & Greed not in extreme fear"),
	}

	var score float64
	for _, s := range signals {
		if s.Active {
			score += s.Weight
			a.Patterns = append(a.Patterns, s.Name)
		}
	}
	a.Signals = signals
	a.BullScore = round(score, 3)

	threshold := minConfidence
	if threshold <= 0 {
		threshold = cfg.MinTradeConfidence
	}
	a.TradeReady = a.BullScore >= threshold
	if a.TradeReady {
		a.TradeReason = "bull score meets confidence threshold for upside long"
	} else {
		a.TradeReason = "bull score below confidence threshold"
	}
	return a
}

func signal(name string, active bool, weight float64, detail string) Signal {
	return Signal{Name: name, Active: active, Weight: weight, Detail: detail}
}

func round(v float64, places int) float64 {
	p := 1.0
	for i := 0; i < places; i++ {
		p *= 10
	}
	return float64(int(v*p+0.5)) / p
}

func SortPatterns(patterns []string) []string {
	out := append([]string(nil), patterns...)
	sort.Strings(out)
	return out
}
