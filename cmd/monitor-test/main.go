package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ym/btc/internal/bot"
	"github.com/ym/btc/internal/config"
	"github.com/ym/btc/internal/oanda"
)

func main() {
	credentialsPath := flag.String("credentials", ".credentials", "path to credentials JSON")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load(*credentialsPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	client := oanda.NewClient(cfg.OANDA.RESTBaseURL(), cfg.OANDA.AccountID, cfg.OANDA.Token)
	b := bot.New(cfg, client)

	if err := b.RunDailyMonitor(context.Background(), ""); err != nil {
		slog.Error("monitor failed", "error", err)
		os.Exit(1)
	}

	st := b.Status()
	out, _ := json.MarshalIndent(map[string]any{
		"close":        st.LastClosePrice,
		"bull_score":   st.LastBullScore,
		"patterns":     st.LastPatterns,
		"trade_ready":  st.TradeReady,
		"fear_greed":   st.LastFearGreed,
		"monitor_date": st.LastMonitorDate,
	}, "", "  ")
	fmt.Println(string(out))
}
