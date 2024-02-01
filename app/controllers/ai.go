package controllers

import (
	"fmt"
	"log"
	"stock-with-alpha/alpha"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
	"time"

	"github.com/markcheno/go-talib"
)

type AI struct{
	API 					*alpha.APIClient
	UsePercent 				float64
	Duration 				time.Duration
	PastPeriod 				int
	SignalEvents 			map[string]*models.SignalEvents
	OptimizedTradeParams 	map[string]*models.TradeParams
	StopLimit 				float64
	StopLimitPercent 		float64
	BackTest 				bool
	StartTrade 				time.Time
}

// 今回はAiオブジェクトはグローバルに宣言して取引していく
var Ai *AI


func NewAI(duration time.Duration, pastPeriod int, usePercent, stopLimitPercent float64, backTest bool) *AI{
	apiClient := alpha.New(config.Config.ApiKey)
	// var signalEvents *models.SignalEvents

	// signalEvents = models.NewSignalEvents()

	// if backTest{
	// 	signalEvents = models.NewSignalEvents()
	// }else{
	// 	signalEvents = models.GetSignalEventsByCount(1)
	// }

	Ai = &AI{
		API: apiClient,
		UsePercent: usePercent,
		PastPeriod: pastPeriod,
		Duration: duration,
		SignalEvents: make(map[string]*models.SignalEvents),
		OptimizedTradeParams: make(map[string]*models.TradeParams),
		BackTest: backTest,
		StartTrade: time.Now().UTC(),
		StopLimitPercent: stopLimitPercent,
	}
	
	tableNames := models.GetCandleTableNames()
	for i := 0; i < len(tableNames);i++{
		var signalEvents *models.SignalEvents
		signalEvents = models.NewSignalEvents()
		symbol := tableNames[i][0]
		product_name := tableNames[i][1]
		Ai.SignalEvents[symbol] = signalEvents
		Ai.UpdateOptimizeParams(symbol, product_name)
	}

	return Ai
} 

func (ai *AI) UpdateOptimizeParams(symbol, product_name string){
	df, _ := models.GetAllCandle(symbol,product_name, ai.Duration, ai.PastPeriod)
	ai.OptimizedTradeParams[symbol] = df.OptimizeParams()
}


func (ai *AI) Buy(symbol string, candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool){
	if ai.BackTest{
		couldBuy := ai.SignalEvents[symbol].Buy(symbol, candle.DateTime, candle.Close, 1.0, false)
		return "", couldBuy
	}

	return childOrderAcceptanceID, isOrderCompleted
}


func (ai *AI) Sell(symbol string, candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool){
	if ai.BackTest{
		couldSell := ai.SignalEvents[symbol].Sell(symbol, candle.DateTime, candle.Close, 1.0, false)

		return "", couldSell
	}
	return childOrderAcceptanceID, isOrderCompleted

}


func (ai *AI) Trade(symbol, name string){

	params := ai.OptimizedTradeParams[symbol]
	df, _ := models.GetAllCandle(symbol, name, ai.Duration, ai.PastPeriod)
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
			_, isCompleted := ai.Buy(df.Symbol, df.Candles[i])
			if !isCompleted{
				continue
			}
			//stopLimitPercentの比率だけ購入時の値段(終値)にかけた値をStoplimitとする→stopLimitを下回ったら自動で売りに抱える
			ai.StopLimit = df.Candles[i].Close * ai.StopLimitPercent

			parsedDateNow := time.Now().Format("2006-01-02")
			parsedCandleDate := df.Candles[i].DateTime.Format("2006-01-02")

			if parsedDateNow == parsedCandleDate {
				body_content := fmt.Sprintf("Product name : %s(%s)", df.Symbol, df.Name)
				email := EmailTemplate{
					Subject: "Buy Signal",
					Body: body_content,
				}
				err := email.Send(config.Config.Email)

				if err != nil{
					log.Println(err)
				}
			}
		}

		if sellPoint > 0 || df.Candles[i].Close < ai.StopLimit{
			_, isCompleted := ai.Sell(df.Symbol, df.Candles[i])
			if !isCompleted{
				continue
			}
			ai.StopLimit = 0.0
			ai.OptimizedTradeParams[df.Symbol] = df.OptimizeParams()

			// subject_content := fmt.Sprintf("You may have profit by selling today!")
			// body_content := fmt.Sprintf("Product name : %s(%s)", df.Symbol, df.Name)

			// if df.Candles[i].Close < ai.StopLimit{
			// 	subject_content = fmt.Sprintf("This is a stop-limit signal")
			// }

			// parsedDateNow := time.Now().Format("2006-01-02")
			// parsedCandleDate := df.Candles[i].DateTime.Format("2006-01-02")

			// if parsedDateNow == parsedCandleDate {
			// 	email := EmailTemplate{
			// 		Subject: subject_content,
			// 		Body: body_content,
			// 	}
			// 	err := email.Send(config.Config.Email)

			// 	if err != nil{
			// 		log.Println(err)
			// 	}
			// }
		}

	}
}
