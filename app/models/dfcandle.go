package models

import (
	"time"
	"github.com/markcheno/go-talib"
)

// dfの存在意義：全ての各キャンドルの特定の項目だけを取り出してリストとして扱える

type DataFrameCandle struct{
	Symbol 	 string 		`json:"symbol"`
	Candles  []Candle 		`json:"candles"`
	Duration time.Duration 	`json:"duration"`
	Smas 	 []Sma 			`json:"smas,omitempty"`
	Emas 	 []Ema 			`json:"emas,omitempty"`
}


type Sma struct{
	Period int `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}

type Ema struct {
	Period int `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}


// 各キャンドルの日にちを取得
func (df *DataFrameCandle) DateTimes() []time.Time {
	s := make([]time.Time, len(df.Candles))

	for i, candle := range df.Candles{
		s[i] = candle.DateTime
	}

	return s
}

// 各キャンドルの開始値を取得
func (df *DataFrameCandle) Opens() []float64{
	s := make([]float64, len(df.Candles))

	for i, candle := range df.Candles{
		s[i] = candle.Open
	}

	return s
}

// 各キャンドルの終値を取得
func (df *DataFrameCandle) Closes() []float64{
	s := make([]float64, len(df.Candles))

	for i, candle := range df.Candles{
		s[i] = candle.Close
	}

	return s
}

// 各キャンドルの高値を取得
func (df *DataFrameCandle) Highs() []float64{
	s := make([]float64, len(df.Candles))

	for i, candle := range df.Candles{
		s[i] = candle.High
	}

	return s
}

// 各キャンドルの低値を取得
func (df *DataFrameCandle) Lows() []float64{
	s := make([]float64, len(df.Candles))

	for i, candle := range df.Candles{
		s[i] = candle.Low
	}

	return s
}

// 各キャンドルの出来高を取得
func (df *DataFrameCandle) Volumes() []float64{
	s := make([]float64, len(df.Candles))

	for i, candle := range df.Candles{
		s[i] = candle.Volume
	}

	return s
}


func (df *DataFrameCandle) AddSma(period int) bool {
	if len(df.Candles) > period{
		df.Smas = append(df.Smas, Sma {
			Period: period,
			Values: talib.Sma(df.Closes(), period),
		})

		return true
	}
	
	return false
}


func (df *DataFrameCandle) AddEma(period int) bool{
	if len(df.Candles) > period{
		df.Emas = append(df.Emas, Ema{
			Period: period,
			Values: talib.Ema(df.Closes(), period),
		})
		return true
	}
	return false
}
