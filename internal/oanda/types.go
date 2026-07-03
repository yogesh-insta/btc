package oanda

type AccountSummary struct {
	Account struct {
		ID              string `json:"id"`
		Balance         string `json:"balance"`
		NAV             string `json:"NAV"`
		MarginAvailable string `json:"marginAvailable"`
	} `json:"account"`
}

type CandlesResponse struct {
	Instrument  string   `json:"instrument"`
	Granularity string   `json:"granularity"`
	Candles     []Candle `json:"candles"`
}

type Candle struct {
	Complete bool      `json:"complete"`
	Time     string    `json:"time"`
	Volume   int       `json:"volume"`
	Mid      OHLCPrice `json:"mid"`
}

type OHLCPrice struct {
	O string `json:"o"`
	H string `json:"h"`
	L string `json:"l"`
	C string `json:"c"`
}

const (
	OrderTypeMarket     = "MARKET"
	TimeInForceGTC      = "GTC"
	PositionFillDefault = "DEFAULT"
)

type CreateOrderRequest struct {
	Order OrderSpec `json:"order"`
}

type OrderSpec struct {
	Type           string          `json:"type"`
	Instrument     string          `json:"instrument"`
	Units          string          `json:"units"`
	TimeInForce    string          `json:"timeInForce"`
	PositionFill   string          `json:"positionFill"`
	StopLossOnFill *OnFillStopLoss `json:"stopLossOnFill,omitempty"`
}

type OnFillStopLoss struct {
	Price       string `json:"price"`
	TimeInForce string `json:"timeInForce"`
}

type CreateOrderResponse struct {
	OrderCreateTransaction *Transaction `json:"orderCreateTransaction"`
	OrderFillTransaction   *Transaction `json:"orderFillTransaction"`
	LastTransactionID      string       `json:"lastTransactionID"`
}

type Transaction struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Instrument  string `json:"instrument"`
	Units       string `json:"units"`
	Price       string `json:"price"`
	TradeOpened *struct {
		TradeID string `json:"tradeID"`
		Units   string `json:"units"`
		Price   string `json:"price"`
	} `json:"tradeOpened"`
}

type OpenTradesResponse struct {
	Trades []Trade `json:"trades"`
}

type Trade struct {
	ID           string `json:"id"`
	Instrument   string `json:"instrument"`
	CurrentUnits string `json:"currentUnits"`
	OpenTime     string `json:"openTime"`
}

type CloseTradeRequest struct {
	Units string `json:"units"`
}

type CloseTradeResponse struct {
	OrderCreateTransaction *Transaction `json:"orderCreateTransaction"`
	OrderFillTransaction   *Transaction `json:"orderFillTransaction"`
	LastTransactionID      string       `json:"lastTransactionID"`
}
