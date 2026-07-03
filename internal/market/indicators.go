package market

import "math"

func SMA(values []float64, period int) float64 {
	if len(values) == 0 || period <= 0 {
		return 0
	}
	if len(values) < period {
		period = len(values)
	}
	sum := 0.0
	for i := len(values) - period; i < len(values); i++ {
		sum += values[i]
	}
	return sum / float64(period)
}

func EMA(values []float64, period int) float64 {
	if len(values) == 0 || period <= 0 {
		return 0
	}
	if len(values) < period {
		period = len(values)
	}
	k := 2.0 / float64(period+1)
	ema := values[0]
	for i := 1; i < len(values); i++ {
		ema = values[i]*k + ema*(1-k)
	}
	return ema
}

func HighLow(values []float64) (high, low float64) {
	if len(values) == 0 {
		return 0, 0
	}
	high, low = values[0], values[0]
	for _, v := range values[1:] {
		if v > high {
			high = v
		}
		if v < low {
			low = v
		}
	}
	return high, low
}

func PctChange(from, to float64) float64 {
	if from == 0 {
		return 0
	}
	return (to - from) / from * 100
}

func RSI(closes []float64, period int) float64 {
	if len(closes) < period+1 || period <= 0 {
		return 50
	}
	var gains, losses float64
	for i := len(closes) - period; i < len(closes); i++ {
		diff := closes[i] - closes[i-1]
		if diff >= 0 {
			gains += diff
		} else {
			losses -= diff
		}
	}
	if losses == 0 {
		return 100
	}
	rs := gains / losses
	return 100 - (100 / (1 + rs))
}

// LocalLows returns approximate swing lows using a simple 2-bar lookback.
func LocalLows(values []float64, lookback int) []float64 {
	if lookback < 1 || len(values) < lookback*2+1 {
		return nil
	}
	var lows []float64
	for i := lookback; i < len(values)-lookback; i++ {
		isLow := true
		for j := i - lookback; j <= i+lookback; j++ {
			if j == i {
				continue
			}
			if values[j] <= values[i] {
				isLow = false
				break
			}
		}
		if isLow {
			lows = append(lows, values[i])
		}
	}
	return lows
}

func HigherLows(lows []float64, count int) bool {
	if len(lows) < count || count < 2 {
		return false
	}
	tail := lows[len(lows)-count:]
	for i := 1; i < len(tail); i++ {
		if tail[i] <= tail[i-1] {
			return false
		}
	}
	return true
}

func MaxOf(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		m = math.Max(m, v)
	}
	return m
}
