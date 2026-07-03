package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ym/btc/internal/bot"
	"github.com/ym/btc/internal/config"
	"github.com/ym/btc/internal/health"
	"github.com/ym/btc/internal/oanda"
)

func main() {
	credentialsPath := flag.String("credentials", ".credentials", "path to credentials JSON")
	healthAddr := flag.String("health-addr", ":8081", "health HTTP listen address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load(*credentialsPath)
	if err != nil {
		slog.Error("load config", "error", err, "hint", "copy .credentials.example to .credentials and fill OANDA account_id + token")
		os.Exit(1)
	}

	client := oanda.NewClient(cfg.OANDA.RESTBaseURL(), cfg.OANDA.AccountID, cfg.OANDA.Token)
	b := bot.New(cfg, client)

	healthSrv := health.NewServer()
	healthSrv.SetHaltedCheck(func() bool { return fileExists(cfg.Halt.File) })
	healthSrv.SetKillHandler(func() error { return touchFile(cfg.Halt.File) })
	healthSrv.SetBotStatus(b.Status)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := runStartupChecks(ctx, client, cfg); err != nil {
		slog.Error("startup checks failed", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("health server listening", "addr", *healthAddr)
		if err := http.ListenAndServe(*healthAddr, healthSrv.Handler()); err != nil {
			slog.Error("health server", "error", err)
			cancel()
		}
	}()

	if err := b.Run(ctx); err != nil {
		slog.Error("bot exited", "error", err)
		os.Exit(1)
	}
}

func runStartupChecks(ctx context.Context, client *oanda.Client, cfg *config.Config) error {
	summary, err := client.AccountSummary(ctx)
	if err != nil {
		return err
	}
	slog.Info("account connected",
		"id", summary.Account.ID,
		"balance", summary.Account.Balance,
		"nav", summary.Account.NAV,
		"margin_available", summary.Account.MarginAvailable,
	)

	candles, err := client.Candles(ctx, cfg.Bot.Instrument, "D", 1)
	if err != nil {
		return err
	}
	slog.Info("candles fetched",
		"instrument", cfg.Bot.Instrument,
		"count", len(candles.Candles),
		"granularity", candles.Granularity,
	)

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func touchFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}
