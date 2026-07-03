package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ym/btc/internal/config"
	"github.com/ym/btc/internal/feargreed"
	"github.com/ym/btc/internal/journal"
	"github.com/ym/btc/internal/market"
	"github.com/ym/btc/internal/notify"
	"github.com/ym/btc/internal/oanda"
	"github.com/ym/btc/internal/pattern"
	"github.com/ym/btc/internal/schedule"
	"github.com/ym/btc/internal/state"
)

type Status struct {
	LastCycleAt       time.Time
	CyclesCompleted   uint64
	OpenLongPositions int
	LastFearGreed     int
	LastClosePrice    float64
	LastBullScore     float64
	LastPatterns      []string
	TradeReady        bool
	LastMonitorDate   string
	LastAnalysisErr   string
	TradesThisWeek    int
	TradeEnabled      bool
}

type Bot struct {
	cfg        *config.Config
	client     *oanda.Client
	fg         *feargreed.Client
	notifier   notify.Notifier
	priceLog   *journal.PriceWriter
	tradeLog   *journal.TradeWriter
	stateStore *state.Store
	snap       state.Snapshot
	status     Status
	monitorLoc *time.Location
	tradeLoc   *time.Location
}

func New(cfg *config.Config, client *oanda.Client) *Bot {
	snap, err := state.NewStore(cfg.State.File).Load()
	if err != nil {
		slog.Warn("state load failed", "error", err)
	}
	return &Bot{
		cfg:        cfg,
		client:     client,
		fg:         feargreed.NewClient(),
		notifier:   notify.New(cfg.Email),
		priceLog:   journal.NewPriceWriter(cfg.Monitor.AuditDir),
		tradeLog:   journal.NewTradeWriter(cfg.Trade.JournalDir),
		stateStore: state.NewStore(cfg.State.File),
		snap:       snap,
		monitorLoc: schedule.LoadLocation(cfg.Monitor.Timezone),
		tradeLoc:   schedule.LoadLocation(cfg.Trade.Timezone),
	}
}

func (b *Bot) Status() Status {
	st := b.status
	st.TradeEnabled = b.cfg.Trade.Enabled
	st.TradesThisWeek = b.snap.TradesThisWeek
	st.LastMonitorDate = b.snap.LastMonitorDate
	if b.snap.LastAnalysis != nil {
		st.LastClosePrice = b.snap.LastAnalysis.Close
		st.LastBullScore = b.snap.LastAnalysis.BullScore
		st.LastPatterns = b.snap.LastAnalysis.Patterns
		st.TradeReady = b.snap.LastAnalysis.TradeReady
		st.LastFearGreed = b.snap.LastAnalysis.FearGreed
	}
	return st
}

func (b *Bot) Run(ctx context.Context) error {
	slog.Info("bot starting",
		"environment", b.cfg.OANDA.Environment,
		"instrument", b.cfg.Bot.Instrument,
		"loop_interval_seconds", b.cfg.Bot.LoopIntervalSeconds,
		"monitor_enabled", b.cfg.Monitor.Enabled,
		"trade_enabled", b.cfg.Trade.Enabled,
	)

	if b.cfg.Monitor.Enabled {
		b.maybeRunScheduledMonitor(ctx, time.Now().UTC())
	}

	for {
		if err := b.RunOnce(ctx); err != nil && ctx.Err() == nil {
			slog.Warn("cycle error", "error", err)
		}

		select {
		case <-ctx.Done():
			_ = b.stateStore.Save(b.snap)
			slog.Info("bot shutdown", "reason", ctx.Err())
			return nil
		case <-time.After(time.Duration(b.cfg.Bot.LoopIntervalSeconds) * time.Second):
		}
	}
}

func (b *Bot) RunOnce(ctx context.Context) error {
	if b.isHalted() {
		slog.Warn("halt file active — skipping cycle", "file", b.cfg.Halt.File)
		return nil
	}

	now := time.Now().UTC()
	slog.Info("cycle_begin", "at", now.Format(time.RFC3339))

	if err := b.manageOpenPositions(ctx); err != nil {
		slog.Warn("manage_open_positions_failed", "error", err)
	}

	if b.cfg.Monitor.Enabled {
		b.maybeRunScheduledMonitor(ctx, now)
	}

	if b.cfg.Trade.Enabled {
		if err := b.evaluateWeekdayTrade(ctx, now); err != nil {
			slog.Warn("evaluate_trade_failed", "error", err)
		}
	}

	b.status.LastCycleAt = time.Now().UTC()
	b.status.CyclesCompleted++
	slog.Info("cycle_complete", "at", b.status.LastCycleAt.Format(time.RFC3339))
	return nil
}

// RunDailyMonitor fetches daily candles, detects bull-market patterns, logs, and alerts.
func (b *Bot) RunDailyMonitor(ctx context.Context, alertSlot string) error {
	series, fearGreed, err := b.fetchMarketContext(ctx)
	if err != nil {
		return err
	}

	pcfg := pattern.DefaultConfig()
	pcfg.MinTradeConfidence = b.cfg.Trade.MinConfidence
	analysis := pattern.Analyze(series, fearGreed, pcfg, b.cfg.Trade.MinConfidence)
	analysis.Date = schedule.LocalDate(time.Now().UTC(), b.monitorLoc)

	if err := b.priceLog.Append(analysis); err != nil {
		slog.Warn("price journal write failed", "error", err)
	}

	b.snap.LastMonitorDate = analysis.Date
	b.snap.LastAnalysis = &analysis
	b.status.LastAnalysisErr = ""

	if alertSlot != "" {
		b.sendMonitorAlert(ctx, alertSlot, analysis)
	}

	_ = b.stateStore.Save(b.snap)

	slog.Info("daily_monitor",
		"date", analysis.Date,
		"alert_slot", alertSlot,
		"close", analysis.Close,
		"sma20", analysis.SMA20,
		"sma50", analysis.SMA50,
		"sma200", analysis.SMA200,
		"rsi14", analysis.RSI14,
		"roc_7d_pct", analysis.ROC7d,
		"fear_greed", analysis.FearGreed,
		"patterns", analysis.Patterns,
		"bull_score", analysis.BullScore,
		"trade_ready", analysis.TradeReady,
	)
	return nil
}

func (b *Bot) maybeRunScheduledMonitor(ctx context.Context, now time.Time) {
	slot, due := schedule.DueAlertSlot(now, b.monitorLoc, b.cfg.Monitor.AlertHoursLocal, b.snap.SentAlertSlots)
	if !due {
		return
	}
	if err := b.RunDailyMonitor(ctx, slot); err != nil {
		b.status.LastAnalysisErr = err.Error()
		slog.Warn("daily_monitor_failed", "error", err, "slot", slot)
	}
}

func (b *Bot) sendMonitorAlert(ctx context.Context, slot string, analysis pattern.Analysis) {
	if !b.cfg.Monitor.AlertsEnabled {
		return
	}
	if !b.notifier.Enabled() {
		slog.Warn("alert skipped", "reason", "email not configured", "slot", slot)
		return
	}

	label := schedule.AlertSlotLabel(slot, b.monitorLoc)
	subject := fmt.Sprintf("btc: monitor %s — $%.0f bull %d%%",
		label, analysis.Close, int(analysis.BullScore*100))
	body := notify.FormatMonitorAlert(b.cfg.Bot.Instrument, label, analysis)
	b.notifier.Send(ctx, subject, body)

	b.snap.SentAlertSlots = append(b.snap.SentAlertSlots, slot)
	if len(b.snap.SentAlertSlots) > 20 {
		b.snap.SentAlertSlots = b.snap.SentAlertSlots[len(b.snap.SentAlertSlots)-20:]
	}
}

func (b *Bot) evaluateWeekdayTrade(ctx context.Context, now time.Time) error {
	if b.cfg.Trade.WeekdaysOnly && !schedule.IsWeekday(now, b.tradeLoc) {
		return nil
	}
	if !schedule.DueOncePerDayHour(now, b.tradeLoc, b.cfg.Trade.TradeHourLocal, b.snap.LastTradeDate) {
		return nil
	}

	week := schedule.ISOWeek(now, b.tradeLoc)
	if b.snap.TradeWeek != week {
		b.snap.TradeWeek = week
		b.snap.TradesThisWeek = 0
	}
	if b.snap.TradesThisWeek >= b.cfg.Trade.MaxTradesPerWeek {
		slog.Info("trade_skipped", "reason", "max trades per week reached", "week", week)
		return nil
	}

	trades, err := b.fetchOpenBTCLongTrades(ctx)
	if err != nil {
		return err
	}
	b.status.OpenLongPositions = len(trades)
	if len(trades) > 0 {
		slog.Info("trade_skipped", "reason", "open BTC_USD long already exists")
		return nil
	}

	analysis, err := b.latestAnalysis(ctx)
	if err != nil {
		return err
	}
	if !analysis.TradeReady {
		_ = b.tradeLog.Append(journal.TradeEntry{
			Action:    "skip",
			Reason:    "bull score below confidence threshold",
			BullScore: analysis.BullScore,
		})
		slog.Info("trade_skipped",
			"reason", "low confidence",
			"bull_score", analysis.BullScore,
			"min_confidence", b.cfg.Trade.MinConfidence,
		)
		return nil
	}

	sizeLabel := "half"
	marginScale := 0.5
	if analysis.BullScore >= b.cfg.Trade.FullSizeConfidence {
		sizeLabel = "full"
		marginScale = 1.0
	}

	resp, err := b.executeLong(ctx, marginScale)
	if err != nil {
		return err
	}
	if resp == nil {
		return nil
	}

	b.snap.LastTradeDate = schedule.LocalDate(now, b.tradeLoc)
	b.snap.TradesThisWeek++
	_ = b.stateStore.Save(b.snap)

	_ = b.tradeLog.Append(journal.TradeEntry{
		Action:     "long",
		Reason:     "high-confidence upside pattern on weekday",
		BullScore:  analysis.BullScore,
		Confidence: sizeLabel,
		Details: map[string]any{
			"patterns": analysis.Patterns,
			"response": summarizeOrderResponse(resp),
		},
	})
	return nil
}

func (b *Bot) latestAnalysis(ctx context.Context) (pattern.Analysis, error) {
	if b.snap.LastAnalysis != nil && b.snap.LastMonitorDate == schedule.LocalDate(time.Now().UTC(), b.monitorLoc) {
		return *b.snap.LastAnalysis, nil
	}
	series, fearGreed, err := b.fetchMarketContext(ctx)
	if err != nil {
		return pattern.Analysis{}, err
	}
	pcfg := pattern.DefaultConfig()
	pcfg.MinTradeConfidence = b.cfg.Trade.MinConfidence
	return pattern.Analyze(series, fearGreed, pcfg, b.cfg.Trade.MinConfidence), nil
}

func (b *Bot) fetchMarketContext(ctx context.Context) (market.Series, int, error) {
	resp, err := b.client.Candles(ctx, b.cfg.Bot.Instrument, "D", b.cfg.Monitor.LookbackDays)
	if err != nil {
		return market.Series{}, 0, fmt.Errorf("daily_candles: %w", err)
	}
	series, err := market.ParseCandles(resp.Candles)
	if err != nil {
		return market.Series{}, 0, err
	}

	fearGreed, err := b.fg.Fetch(ctx)
	if err != nil {
		return market.Series{}, 0, fmt.Errorf("fear_greed: %w", err)
	}
	return series, fearGreed, nil
}

func (b *Bot) executeLong(ctx context.Context, marginScale float64) (*oanda.CreateOrderResponse, error) {
	marginAvailable, err := b.fetchAvailableMargin(ctx)
	if err != nil {
		return nil, err
	}
	marginAvailable *= marginScale

	midPrice, err := b.fetchCurrentMidClose(ctx)
	if err != nil {
		return nil, err
	}

	units := oanda.CalculatePositionUnits(
		marginAvailable,
		midPrice,
		b.cfg.Bot.ASICMaxLeverage,
		b.cfg.Bot.MarginUtilization,
	)
	if units <= 0 {
		slog.Info("entry_skipped", "reason", "calculated units is zero")
		return nil, nil
	}

	stopLossPrice := midPrice * (1.0 - b.cfg.Bot.StopLossPercent)
	slog.Info("placing_long_order",
		"units", oanda.FormatUnits(units),
		"mid_price", midPrice,
		"stop_loss", stopLossPrice,
		"margin_available", marginAvailable,
	)

	resp, err := b.client.CreateOrder(ctx, oanda.CreateOrderRequest{
		Order: oanda.OrderSpec{
			Type:         oanda.OrderTypeMarket,
			Instrument:   b.cfg.Bot.Instrument,
			Units:        oanda.FormatUnits(units),
			TimeInForce:  oanda.TimeInForceGTC,
			PositionFill: oanda.PositionFillDefault,
			StopLossOnFill: &oanda.OnFillStopLoss{
				Price:       oanda.FormatPrice(stopLossPrice),
				TimeInForce: oanda.TimeInForceGTC,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	slog.Info("order_submitted", "response", summarizeOrderResponse(resp))
	return resp, nil
}

func (b *Bot) manageOpenPositions(ctx context.Context) error {
	trades, err := b.fetchOpenBTCLongTrades(ctx)
	if err != nil {
		return err
	}

	b.status.OpenLongPositions = len(trades)
	now := time.Now().UTC()
	for _, trade := range trades {
		openTime, err := oanda.ParseTime(trade.OpenTime)
		if err != nil {
			slog.Warn("trade_inspection_failed", "trade_id", trade.ID, "error", err)
			continue
		}
		ageDays := now.Sub(openTime).Hours() / 24.0
		if ageDays <= float64(b.cfg.Bot.MaxHoldDays) {
			continue
		}
		resp, err := b.client.CloseTrade(ctx, trade.ID, "ALL")
		if err != nil {
			slog.Warn("close_trade_failed", "trade_id", trade.ID, "error", err)
			continue
		}
		slog.Info("trade_closed", "trade_id", trade.ID, "age_days", fmt.Sprintf("%.2f", ageDays), "response", summarizeCloseResponse(resp))
	}
	return nil
}

func (b *Bot) fetchOpenBTCLongTrades(ctx context.Context) ([]oanda.Trade, error) {
	resp, err := b.client.OpenTrades(ctx)
	if err != nil {
		return nil, fmt.Errorf("open_trades: %w", err)
	}
	var trades []oanda.Trade
	for _, trade := range resp.Trades {
		if trade.Instrument != b.cfg.Bot.Instrument {
			continue
		}
		units, err := oanda.ParsePrice(trade.CurrentUnits)
		if err != nil {
			return nil, fmt.Errorf("parse trade units: %w", err)
		}
		if units > 0 {
			trades = append(trades, trade)
		}
	}
	return trades, nil
}

func (b *Bot) fetchAvailableMargin(ctx context.Context) (float64, error) {
	summary, err := b.client.AccountSummary(ctx)
	if err != nil {
		return 0, fmt.Errorf("account_summary: %w", err)
	}
	if summary.Account.MarginAvailable == "" {
		return 0, fmt.Errorf("account summary missing marginAvailable")
	}
	return oanda.ParsePrice(summary.Account.MarginAvailable)
}

func (b *Bot) fetchCurrentMidClose(ctx context.Context) (float64, error) {
	resp, err := b.client.Candles(ctx, b.cfg.Bot.Instrument, "D", 1)
	if err != nil {
		return 0, fmt.Errorf("latest_candle: %w", err)
	}
	if len(resp.Candles) == 0 {
		return 0, fmt.Errorf("no %s candle data returned", b.cfg.Bot.Instrument)
	}
	closePrice := resp.Candles[len(resp.Candles)-1].Mid.C
	if closePrice == "" {
		return 0, fmt.Errorf("candle missing mid close price")
	}
	return oanda.ParsePrice(closePrice)
}

func (b *Bot) isHalted() bool {
	_, err := os.Stat(b.cfg.Halt.File)
	return err == nil
}

func summarizeCloseResponse(resp *oanda.CloseTradeResponse) map[string]string {
	if resp == nil {
		return map[string]string{}
	}
	out := map[string]string{"last_transaction_id": resp.LastTransactionID}
	if fill := resp.OrderFillTransaction; fill != nil {
		out["fill_id"] = fill.ID
		out["price"] = fill.Price
	}
	return out
}

func summarizeOrderResponse(resp *oanda.CreateOrderResponse) map[string]string {
	if resp == nil {
		return map[string]string{}
	}
	out := map[string]string{"last_transaction_id": resp.LastTransactionID}
	if fill := resp.OrderFillTransaction; fill != nil {
		out["fill_id"] = fill.ID
		out["price"] = fill.Price
		if fill.TradeOpened != nil {
			out["trade_id"] = fill.TradeOpened.TradeID
		}
	}
	return out
}
