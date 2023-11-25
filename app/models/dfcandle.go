package models

import "time"

// dfの存在意義：全ての各キャンドルの特定の項目だけを取り出してリストとして扱える

type DataFrameCandle struct{
	Symbol string `json:"symbol"`
	Candles []Candle `json:"candles"`
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
