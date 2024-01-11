package controllers

import (
	"stock-with-alpha/alpha"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
	"time"

	"github.com/markcheno/go-talib"
)

type AI struct{
	API 					*alpha.APIClient
	Symbol 					string
	UsePercent 				float64
	Duration 				time.Duration
	PastPeriod 				int
	SignalEvents 			*models.SignalEvents
	OptimizedTradeParams 	*models.TradeParams
	StopLimit 				float64
	StopLimitPercent 		float64
	BackTest 				bool
	StartTrade 				time.Time
}

// 今回はAiオブジェクトはグローバルに宣言して取引していく
var Ai *AI


func NewAI(symbol string, duration time.Duration, pastPeriod int, usePercent, stopLimitPercent float64, backTest bool) *AI{
	apiClient := alpha.New(config.Config.ApiKey)
	var signalEvents *models.SignalEvents

	if backTest{
		signalEvents = models.NewSignalEvents()
	}else{
		signalEvents = models.GetSignalEventsByCount(1)
	}

	Ai = &AI{
		API: apiClient,
		Symbol: symbol,
		UsePercent: usePercent,
		PastPeriod: pastPeriod,
		Duration: duration,
		SignalEvents: signalEvents,
		BackTest: backTest,
		StartTrade: time.Now().UTC(),
		StopLimitPercent: stopLimitPercent,
	}

	Ai.UpdateOptimizeParams()

	return Ai
} 

func (ai *AI) UpdateOptimizeParams(){
	df, _ := models.GetAllCandle(ai.Symbol,"", ai.Duration, ai.PastPeriod)
	ai.OptimizedTradeParams = df.OptimizeParams()
}


func (ai *AI) Buy(candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool){
	if ai.BackTest{
		couldBuy := ai.SignalEvents.Buy(ai.Symbol, candle.DateTime, candle.Close, 1.0, false)
		return "", couldBuy
	}

	return childOrderAcceptanceID, isOrderCompleted
}


func (ai *AI) Sell(candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool){
	if ai.BackTest{
		couldSell := ai.SignalEvents.Sell(ai.Symbol, candle.DateTime, candle.Close, 1.0, false)

		return "", couldSell
	}
	return childOrderAcceptanceID, isOrderCompleted

}


func (ai *AI) Trade(){
	params := ai.OptimizedTradeParams
	df, _ := models.GetAllCandle(ai.Symbol,"", ai.Duration, ai.PastPeriod)
	lenCandles := len(df.Candles)

	var emaValues1 []float64
	var emaValues2 []float64
	if params.EmaEnable{
		emaValues1 = talib.Ema(df.Closes(), params.EmaPeriod1)
		emaValues2 = talib.Ema(df.Closes(), params.EmaPeriod2)
	}

	var bbUp []float64
	var bbDown []float64
	if params.BbEnable{
		bbUp, _, bbDown = talib.BBands(df.Closes(), params.BbN, params.BbK ,params.BbK,  0)
	}

	var rsiValues []float64
	if params.RsiEnable{
		rsiValues = talib.Rsi(df.Closes(), params.RsiPeriod)
	}

	// 上で取ってきた各アルゴリズムの最適な値を使って対応した各キャンドルを売買していく
	for i := 1; i < lenCandles; i++{
		buyPoint, sellPoint := 0,0

		if params.EmaEnable && params.EmaPeriod1 <= i && params.EmaPeriod2 <= i{
			if emaValues1[i-1] < emaValues2[i-1] && emaValues1[i] >= emaValues2[i]{
				buyPoint++
			}
			if emaValues1[i-1] > emaValues2[i-1] && emaValues1[i] <= emaValues2[i]{
				sellPoint++
			} 
		}

		if params.BbEnable && params.BbN <= i{
			if bbDown[i-1] > df.Candles[i-1].Close && bbDown[i] <= df.Candles[i].Close{
				buyPoint++
			}
			if bbUp[i-1] < df.Candles[i-1].Close && bbUp[i] >= df.Candles[i].Close{
				sellPoint++
			}
		}

		if params.RsiEnable && rsiValues[i-1] != 0 && rsiValues[i-1] != 100{
			if rsiValues[i-1] < params.RsiBuyThread && rsiValues[i] >= params.RsiBuyThread{
				buyPoint++
			}
			if rsiValues[i-1] > params.RsiSellThread && rsiValues[i] <= params.RsiSellThread{
				sellPoint++
			}
		}

		if buyPoint > 0{
			_, isCompleted := ai.Buy(df.Candles[i])
			if !isCompleted{
				continue
			}
			//stopLimitPercentの比率だけ購入時の値段(終値)にかけた値をStoplimitとする→stopLimitを下回ったら自動で売りに抱える
			ai.StopLimit = df.Candles[i].Close * ai.StopLimitPercent
		}

		if sellPoint > 0 || df.Candles[i].Close < ai.StopLimit{
			_, isCompleted := ai.Sell(df.Candles[i])
			if !isCompleted{
				continue
			}
			ai.StopLimit = 0.0
			ai.UpdateOptimizeParams()
		}

	}
}