package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	EnvPractice     = "practice"
	EnvLive         = "live"
	DefaultHaltFile = ".halt"
)

type Config struct {
	OANDA   OANDAConfig   `json:"oanda"`
	Email   EmailConfig   `json:"email"`
	Bot     BotConfig     `json:"bot"`
	Monitor MonitorConfig `json:"monitor"`
	Trade   TradeConfig   `json:"trade"`
	State   StateConfig   `json:"state"`
	Halt    HaltConfig    `json:"halt"`
}

type EmailConfig struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	AlertTo  string `json:"alert_to"`
}

func (e EmailConfig) Enabled() bool {
	return e.SMTPHost != "" && e.Username != "" && e.Password != "" && e.AlertTo != ""
}

type OANDAConfig struct {
	AccountID   string `json:"account_id"`
	Token       string `json:"token"`
	Environment string `json:"environment"`
}

type BotConfig struct {
	Instrument          string  `json:"instrument"`
	LoopIntervalSeconds int     `json:"loop_interval_seconds"`
	MarginUtilization   float64 `json:"margin_utilization"`
	StopLossPercent     float64 `json:"stop_loss_percent"`
	MaxHoldDays         int     `json:"max_hold_days"`
	ASICMaxLeverage     float64 `json:"asic_max_leverage"`
}

type MonitorConfig struct {
	Enabled         bool   `json:"enabled"`
	AlertsEnabled   bool   `json:"alerts_enabled"`
	Timezone        string `json:"timezone"`
	AlertHoursLocal []int  `json:"alert_hours_local"`
	LookbackDays    int    `json:"lookback_days"`
	AuditDir        string `json:"audit_dir"`
}

type TradeConfig struct {
	Enabled            bool    `json:"enabled"`
	Timezone           string  `json:"timezone"`
	WeekdaysOnly       bool    `json:"weekdays_only"`
	TradeHourLocal     int     `json:"trade_hour_local"`
	MinConfidence      float64 `json:"min_confidence"`
	FullSizeConfidence float64 `json:"full_size_confidence"`
	MaxTradesPerWeek   int     `json:"max_trades_per_week"`
	JournalDir         string  `json:"journal_dir"`
}

type StateConfig struct {
	File string `json:"file"`
}

type HaltConfig struct {
	File string `json:"file"`
}

func DefaultBotConfig() BotConfig {
	return BotConfig{
		Instrument:          "BTC_USD",
		LoopIntervalSeconds: 300,
		MarginUtilization:   0.95,
		StopLossPercent:     0.10,
		MaxHoldDays:         90,
		ASICMaxLeverage:     2.0,
	}
}

func DefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		Enabled:         true,
		AlertsEnabled:   true,
		Timezone:        "Australia/Sydney",
		AlertHoursLocal: []int{8, 18},
		LookbackDays:    220,
		AuditDir:        "logs/price",
	}
}

func DefaultTradeConfig() TradeConfig {
	return TradeConfig{
		Enabled:            false,
		Timezone:           "Australia/Sydney",
		WeekdaysOnly:       true,
		TradeHourLocal:     10,
		MinConfidence:      0.70,
		FullSizeConfidence: 0.85,
		MaxTradesPerWeek:   1,
		JournalDir:         "logs/trades",
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.OANDA.Environment == "" {
		c.OANDA.Environment = EnvPractice
	}

	defBot := DefaultBotConfig()
	if c.Bot == (BotConfig{}) {
		c.Bot = defBot
	} else {
		if c.Bot.Instrument == "" {
			c.Bot.Instrument = defBot.Instrument
		}
		if c.Bot.LoopIntervalSeconds == 0 {
			c.Bot.LoopIntervalSeconds = defBot.LoopIntervalSeconds
		}
		if c.Bot.MarginUtilization == 0 {
			c.Bot.MarginUtilization = defBot.MarginUtilization
		}
		if c.Bot.StopLossPercent == 0 {
			c.Bot.StopLossPercent = defBot.StopLossPercent
		}
		if c.Bot.MaxHoldDays == 0 {
			c.Bot.MaxHoldDays = defBot.MaxHoldDays
		}
		if c.Bot.ASICMaxLeverage == 0 {
			c.Bot.ASICMaxLeverage = defBot.ASICMaxLeverage
		}
	}

	if c.Email.SMTPPort == 0 && c.Email.SMTPHost != "" {
		c.Email.SMTPPort = 587
	}

	defMon := DefaultMonitorConfig()
	if c.Monitor.Timezone == "" && c.Monitor.LookbackDays == 0 && c.Monitor.AuditDir == "" && len(c.Monitor.AlertHoursLocal) == 0 {
		c.Monitor = defMon
	} else {
		if c.Monitor.Timezone == "" {
			c.Monitor.Timezone = defMon.Timezone
		}
		if len(c.Monitor.AlertHoursLocal) == 0 {
			c.Monitor.AlertHoursLocal = defMon.AlertHoursLocal
		}
		if c.Monitor.LookbackDays == 0 {
			c.Monitor.LookbackDays = defMon.LookbackDays
		}
		if c.Monitor.AuditDir == "" {
			c.Monitor.AuditDir = defMon.AuditDir
		}
	}

	defTrade := DefaultTradeConfig()
	if c.Trade == (TradeConfig{}) {
		c.Trade = defTrade
	} else {
		if c.Trade.Timezone == "" {
			c.Trade.Timezone = defTrade.Timezone
		}
		if c.Trade.TradeHourLocal == 0 {
			c.Trade.TradeHourLocal = defTrade.TradeHourLocal
		}
		if c.Trade.MinConfidence == 0 {
			c.Trade.MinConfidence = defTrade.MinConfidence
		}
		if c.Trade.FullSizeConfidence == 0 {
			c.Trade.FullSizeConfidence = defTrade.FullSizeConfidence
		}
		if c.Trade.MaxTradesPerWeek == 0 {
			c.Trade.MaxTradesPerWeek = defTrade.MaxTradesPerWeek
		}
		if c.Trade.JournalDir == "" {
			c.Trade.JournalDir = defTrade.JournalDir
		}
	}

	if c.State.File == "" {
		c.State.File = "data/state.json"
	}
	if c.Halt.File == "" {
		c.Halt.File = DefaultHaltFile
	}
}

func (c *Config) Validate() error {
	if c.OANDA.AccountID == "" {
		return fmt.Errorf("oanda.account_id is required")
	}
	if c.OANDA.Token == "" {
		return fmt.Errorf("oanda.token is required")
	}
	if c.OANDA.Environment != EnvPractice && c.OANDA.Environment != EnvLive {
		return fmt.Errorf("oanda.environment must be %q or %q", EnvPractice, EnvLive)
	}
	if c.Bot.Instrument == "" {
		return fmt.Errorf("bot.instrument is required")
	}
	return nil
}

func (o OANDAConfig) RESTBaseURL() string {
	if o.Environment == EnvLive {
		return "https://api-fxtrade.oanda.com"
	}
	return "https://api-fxpractice.oanda.com"
}
