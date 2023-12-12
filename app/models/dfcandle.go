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
	// BBandsがポインタの理由：smas,emasはラインの数が必ずしも固定しなくても良い。一方でBBandsは3つのラインに固定する
	BBands 	 *BBands 		`json:"bbands,omitempty"`
	Rsi 	 *Rsi 			`json:"rsi,omitempty"`
}


type Sma struct{
	Period int `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}

type Ema struct {
	Period int `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}


type BBands struct{
	// 移動平均線
	N 		int 		`json:"n,omitempty"`
	// Kは標準偏差の倍数を示しており、通常は2が使われます。
	// Kが大きくなると標準偏差の倍数も大きくなり、バンドの幅が広がります
	K 		float64 	`json:"k,omitempty"`
	// 各ラインをそれぞれスライスで表現する
	Up 		[]float64 	`json:"up,omitempty"`
	Mid 	[]float64 	`json:"mid,omitempty"`
	Down 	[]float64 	`json:"down,omitempty"`
}


type Rsi struct {
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

func (df *DataFrameCandle) AddBBands(n int, k float64) bool{
	if n <= len(df.Closes()){
		// talibのbbandsメソッドの一番最後は移動平均線のタイプを選ぶ→今回は普通を選択
		up, mid, down := talib.BBands(df.Closes(), n, k, k, 0)
	
		df.BBands = &BBands{
			N:n,
			K:k,
			Up: up,
			Mid: mid,
			Down: down,
		}

		return true
	}
	return false
}

func (df *DataFrameCandle) AddRsi(period int) bool{
	if len(df.Candles) > period{
		values := talib.Rsi(df.Closes(), period)
		df.Rsi = &Rsi{
			Period: period,
			Values: values,
		}
		return true
	}
	return false
}