# btc

Go daemon that watches for a Bitcoin bull market and can open weekday long entries on **OANDA practice** (`BTC_USD`).

Two phases:

1. **Daily monitor** — fetch daily candles, detect bull patterns, log to `logs/price/`
2. **Weekday trades** — Mon–Fri long entries when bull confidence is high (disabled by default)

OANDA `BTC_USD` is a leveraged CFD long (not exchange-traded options), which gives upside exposure similar to calls.

---

## Prerequisites

| Requirement | Notes |
|-------------|--------|
| **Go 1.22+** | `go version` |
| **OANDA practice account** | Demo + API token with BTC_USD |

---

## Setup

```bash
cp .credentials.example .credentials
chmod 600 .credentials
go mod download
```

---

## Part 1: Start daily monitoring

```bash
go run ./cmd/btc
```

On startup and twice daily at **08:00** and **18:00 Sydney**, the bot:

- Fetches ~220 daily candles from OANDA
- Pulls Fear & Greed index
- Scores bull patterns (SMA stack, golden cross, higher lows, 20d breakout, 7d momentum, RSI, sentiment)
- Writes JSON lines to `logs/price/YYYY-MM-DD.jsonl`
- **Emails** the monitor summary to `email.alert_to`
- Saves state to `data/state.json`

**One-shot test:**

```bash
go run ./cmd/monitor-test
```

**Health check:**

```bash
curl http://localhost:8081/health | python3 -m json.tool
```

Look for `bull_score`, `patterns`, `trade_ready`, `last_monitor_date`.

**Email alerts** go to `email.alert_to` at **08:00** and **18:00** Sydney. Change times with `monitor.alert_hours_local` (e.g. `[8, 18]`). Requires Gmail SMTP + app password in `.credentials`.

---

## Part 2: Enable weekday long entries

When you're ready to paper-trade, set in `.credentials`:

```json
"trade": { "enabled": true }
```

Rules:

| Rule | Default |
|------|---------|
| Trading days | Mon–Fri (Australia/Sydney) |
| Trade window | 10:00 local, once per day |
| Min confidence (`bull_score`) | 0.70 |
| Full size confidence | 0.85 |
| Max trades per week | 1 |
| Position | Market long + 10% stop-loss |
| Leverage cap | ASIC 2:1 |

Half size is used between 0.70–0.85 confidence; full size at ≥ 0.85.

Trade decisions log to `logs/trades/journal.jsonl`.

---

## Patterns detected

| Pattern | Meaning |
|---------|---------|
| `above_sma20` | Price above 20-day SMA |
| `above_sma50` | Price above 50-day SMA |
| `above_sma200` | Price above 200-day SMA |
| `golden_cross` | SMA50 > SMA200 |
| `higher_lows` | Ascending swing lows |
| `breakout_20d` | Near/above 20-day high |
| `momentum_7d` | 7-day return ≥ 2% |
| `rsi_bull_zone` | RSI 50–72 |
| `sentiment_ok` | Fear & Greed ≥ 35 |

`bull_score` is the weighted sum of active patterns (0–1).

---

## Safety

| Action | Command |
|--------|---------|
| Emergency stop | `curl -X POST http://localhost:8081/kill` or `touch .halt` |
| Resume | `rm .halt` and restart |

---

## Tests

```bash
go test ./...
```

---

## Project layout

```
btc/
├── cmd/btc/           # main daemon
├── cmd/monitor-test/  # one-shot daily analysis
├── internal/
│   ├── bot/           # monitor + trade loops
│   ├── pattern/       # bull pattern scoring
│   ├── market/        # indicators + candle parsing
│   ├── journal/       # price + trade logs
│   ├── state/         # persisted monitor/trade state
│   └── schedule/      # weekday + daily timing
├── logs/price/        # daily pattern audit
├── logs/trades/       # trade journal
└── data/state.json    # last analysis + weekly trade count
```
