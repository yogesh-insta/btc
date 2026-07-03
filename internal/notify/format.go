package notify

import (
	"fmt"
	"strings"

	"github.com/ym/btc/internal/pattern"
)

func FormatMonitorAlert(instrument, slotLabel string, a pattern.Analysis) string {
	patterns := a.Patterns
	if len(patterns) == 0 {
		patterns = []string{"(none active)"}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "BTC monitor — %s\n\n", slotLabel)
	fmt.Fprintf(&b, "Instrument: %s\n", instrument)
	fmt.Fprintf(&b, "Close: $%.0f\n", a.Close)
	fmt.Fprintf(&b, "Bull score: %.0f%% (trade ready: %s)\n", a.BullScore*100, yesNo(a.TradeReady))
	fmt.Fprintf(&b, "Fear & Greed: %d\n\n", a.FearGreed)

	b.WriteString("Active patterns:\n")
	for _, p := range patterns {
		fmt.Fprintf(&b, "  • %s\n", p)
	}

	fmt.Fprintf(&b, "\nSMA20: $%.0f  SMA50: $%.0f  SMA200: $%.0f\n", a.SMA20, a.SMA50, a.SMA200)
	fmt.Fprintf(&b, "RSI14: %.1f  7d momentum: %+.1f%%\n", a.RSI14, a.ROC7d)
	fmt.Fprintf(&b, "20d high: $%.0f\n", a.High20d)
	return b.String()
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
