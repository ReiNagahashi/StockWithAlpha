package models

import (
	"fmt"
	"log"
	"time"
)


type Candle struct{
	Symbol 		string 			`json:"symbol"`
	Name 		string 			`json:"name"`
	Duration 	time.Duration	`json:"duration"`
	DateTime 	time.Time 		`json:"date_time"`
	Open 		float64 		`json:"open"`
	Close 		float64 		`json:"close"`
	High		float64 		`json:"high"`
	Low 		float64 		`json:"low"`
	Volume 		float64 		`json:"volume"`
}


func NewCandle(symbol, name string, duration time.Duration, dateTime time.Time, open, close, high, low, volume float64) *Candle{
	return &Candle{
		symbol,
		name,
		duration,
		dateTime,
		open,
		close,
		high,
		low,
		volume,
	}
}


func (c *Candle) TableName()string {
	return GetCandleTableName(c.Symbol, c.Duration)
}

// create, saveは各キャンドルオブジェクトが行う関数→メソッド
// getCandleは特定のキャンドルオブジェクトが実行する関数では無い→普通の関数

func (c *Candle) Create(name string) error{
	cmd := fmt.Sprintf("INSERT INTO %s (time, name, open, close, high, low, volume) VALUES (?, ?, ?, ?, ?, ?, ?)", c.TableName())
	_, err := DbConnection.Exec(cmd, c.DateTime, name, c.Open, c.Close, c.High, c.Low, c.Volume)
	if err != nil{
		log.Fatalln(err)
		return err
	}

	return err
}


func GetCandle(symbol, name string, duration time.Duration, dateTime time.Time) *Candle{
	tableName := GetCandleTableName(symbol, duration)
	cmd := fmt.Sprintf("SELECT time, name, open, close, high, low, volume FROM %s WHERE time = ?", tableName)
	row := DbConnection.QueryRow(cmd, dateTime)

	var candle Candle
	err := row.Scan(&candle.DateTime, &candle.Name, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
	if err != nil{
		return nil
	}

	return NewCandle(symbol, name, duration, candle.DateTime, candle.Open, candle.Close, candle.High, candle.Low, candle.Volume)
}


// とってきたキャンドルをデータベースに書き込む
func CreateCandleWithDuration(candle Candle, symbol, name string, date time.Time, duration time.Duration) bool{
	currentCandle := GetCandle(symbol, name, duration, date)

	if currentCandle == nil{
		candle.Create(name)
		return true
	}
	
	return false
}

func GetAllCandle(symbol, name string, duration time.Duration, limit int) (dfCandle *DataFrameCandle, err error){
	tableName := GetCandleTableName(symbol, duration)
	// 一旦キャンドルデータ群をリバースした上でリミットすることで、最新のデータをリミットして取得できるようにする
	// ただし、リバースしたらあとはもとに戻しておくこと(最後にASCにしている)
	cmd := fmt.Sprintf(`SELECT * FROM (
		SELECT time, name, open, close, high, low, volume FROM %s ORDER BY time DESC LIMIT ?
	) ORDER BY time ASC;`, tableName)

	rows, err := DbConnection.Query(cmd, limit)
	if err != nil{
		return
	}
	defer rows.Close()

	dfCandle = &DataFrameCandle{}
	dfCandle.Symbol = symbol
	dfCandle.Name = name
	dfCandle.Duration = duration
	for rows.Next(){
		var candle Candle
		candle.Symbol = symbol
		candle.Duration = duration
		rows.Scan(&candle.DateTime, &candle.Name, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
		dfCandle.Candles = append(dfCandle.Candles, candle)
	}
	err = rows.Err()
	if err != nil{
		log.Fatalln("Failed to get data")
		return 
	}

	return dfCandle, nil
}