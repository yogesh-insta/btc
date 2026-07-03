package health

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ym/btc/internal/bot"
)

type Status struct {
	OK                bool      `json:"ok"`
	StartedAt         time.Time `json:"started_at"`
	CyclesCompleted   uint64    `json:"cycles_completed"`
	LastCycleAt       time.Time `json:"last_cycle_at,omitempty"`
	OpenLongPositions int       `json:"open_long_positions"`
	FearGreed         int       `json:"fear_greed,omitempty"`
	ClosePrice        float64   `json:"close_price,omitempty"`
	BullScore         float64   `json:"bull_score,omitempty"`
	Patterns          []string  `json:"patterns,omitempty"`
	TradeReady        bool      `json:"trade_ready"`
	TradeEnabled      bool      `json:"trade_enabled"`
	LastMonitorDate   string    `json:"last_monitor_date,omitempty"`
	TradesThisWeek    int       `json:"trades_this_week"`
	MonitorError      string    `json:"monitor_error,omitempty"`
	Halted            bool      `json:"halted"`
}

type Server struct {
	startedAt time.Time
	haltedFn  func() bool
	killFn    func() error
	statusFn  func() bot.Status
}

func NewServer() *Server {
	return &Server{startedAt: time.Now()}
}

func (s *Server) SetHaltedCheck(fn func() bool) {
	s.haltedFn = fn
}

func (s *Server) SetKillHandler(fn func() error) {
	s.killFn = fn
}

func (s *Server) SetBotStatus(fn func() bot.Status) {
	s.statusFn = fn
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /kill", s.handleKill)
	return mux
}

func (s *Server) handleKill(w http.ResponseWriter, r *http.Request) {
	if s.killFn == nil {
		http.Error(w, "kill switch not configured", http.StatusServiceUnavailable)
		return
	}
	if err := s.killFn(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "halted": true})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	st := Status{OK: true, StartedAt: s.startedAt}
	if s.haltedFn != nil {
		st.Halted = s.haltedFn()
	}
	if s.statusFn != nil {
		bs := s.statusFn()
		st.CyclesCompleted = bs.CyclesCompleted
		st.LastCycleAt = bs.LastCycleAt
		st.OpenLongPositions = bs.OpenLongPositions
		st.FearGreed = bs.LastFearGreed
		st.ClosePrice = bs.LastClosePrice
		st.BullScore = bs.LastBullScore
		st.Patterns = bs.LastPatterns
		st.TradeReady = bs.TradeReady
		st.TradeEnabled = bs.TradeEnabled
		st.LastMonitorDate = bs.LastMonitorDate
		st.TradesThisWeek = bs.TradesThisWeek
		st.MonitorError = bs.LastAnalysisErr
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(st)
}
