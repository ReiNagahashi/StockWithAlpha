package models

import (
	"fmt"
	"log"
	"stock-with-alpha/alpha"
	"strconv"
	"time"
)



type Candle struct{
	Symbol 		string 			`json:"symbol"`
	DateTime 	time.Time 		`json:"date_time"`
	Open 		float64 		`json:"open"`
	Close 		float64 		`json:"close"`
	High		float64 		`json:"high"`
	Low 		float64 		`json:"low"`
	Volume 		float64 		`json:"volume"`
}


func NewCandle(symbol string, dateTime time.Time, open, close, high, low, volume float64) *Candle{
	return &Candle{
		symbol,
		dateTime,
		open,
		close,
		high,
		low,
		volume,
	}
}


func (c *Candle) TableName()string {
	return GetCandleTableName(c.Symbol)
}

// create, saveは各キャンドルオブジェクトが行う関数→メソッド
// getCandleは特定のキャンドルオブジェクトが実行する関数では無い→普通の関数

func (c *Candle) Create() error{
	cmd := fmt.Sprintf("INSERT INTO %s (time, open, close, high, low, volume) VALUES (?, ?, ?, ?, ?, ?)", c.TableName())
	_, err := DbConnection.Exec(cmd, c.DateTime, c.Open, c.Close, c.High, c.Low, c.Volume)
	if err != nil{
		return err
	}

	return err
}


// func (c *Candle) Save() error{
// 	cmd := fmt.Sprintf("UPDATE %s SET open = ?, close = ?, high = ?, low = ?, volume = ? WHERE time = ?", c.TableName())
// 	_, err := DbConnection.Exec(cmd, c.Open, c.Close, c.High, c.Low, c.Volume, c.Time.Format(time.RFC3339))
// 	if err != nil{
// 		return err
// 	}

// 	return err
// }

func GetCandle(symbol string, dateTime time.Time) *Candle{
	tableName := GetCandleTableName(symbol)
	cmd := fmt.Sprintf("SELECT time, open, close, high, low, volume FROM %s WHERE time = ?", tableName)
	row := DbConnection.QueryRow(cmd, dateTime)

	var candle Candle
	err := row.Scan(&candle.DateTime, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
	if err != nil{
		return nil
	}

	return NewCandle(symbol, candle.DateTime, candle.Open, candle.Close, candle.High, candle.Low, candle.Volume)
}


// とってきたティッカーをデータベースに書き込む
func CreateCandle(ticker alpha.Ticker, symbol string, dateTime time.Time) bool{
	currentCandle := GetCandle(symbol, dateTime)
	if currentCandle == nil{
		open, _ := strconv.ParseFloat(ticker.GlobalQuote.Open, 64)
		close, _ := strconv.ParseFloat(ticker.GlobalQuote.Close, 64)
		high, _ := strconv.ParseFloat(ticker.GlobalQuote.High, 64)
		low, _ := strconv.ParseFloat(ticker.GlobalQuote.Low, 64)
		volume, _ := strconv.ParseFloat(ticker.GlobalQuote.Volume, 64)

		candle := NewCandle(symbol, dateTime, open, close, high, low, volume)
		candle.Create()

		return true
	}

	return false
}

func GetAllCandle(symbol string, limit int) (dfCandle *DataFrameCandle, err error){
	tableName := GetCandleTableName(symbol)
	// 一旦ティッカーデータ群をリバースした上でリミットすることで、最新のデータをリミットして取得できるようにする
	// ただし、リバースしたらあとはもとに戻しておくこと(最後にASCにしている)
	cmd := fmt.Sprintf(`SELECT * FROM (
		SELECT time, open, close, high, low, volume FROM %s ORDER BY time DESC LIMIT ?
	) ORDER BY time ASC;`, tableName)

	rows, err := DbConnection.Query(cmd, limit)
	if err != nil{
		return
	}
	defer rows.Close()

	dfCandle = &DataFrameCandle{}
	dfCandle.Symbol = symbol
	for rows.Next(){
		var candle Candle
		candle.Symbol = symbol
		rows.Scan(&candle.DateTime, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
		dfCandle.Candles = append(dfCandle.Candles, candle)
	}
	err = rows.Err()
	if err != nil{
		log.Fatalln("Failed to get data")
		return 
	}

	return dfCandle, nil
}