package models

import (
	"database/sql"
	"fmt"
	"log"
	"stock-with-alpha/config"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

const (
	tableNameSignalEvents = "signal_events"
)

var DbConnection *sql.DB


func GetCandleTableName(symbol string, duration time.Duration) string{
	return fmt.Sprintf("%s_%s", symbol, duration)
}

// テーブルを作成(日足)
func CreateTableBySymbol(symbol string){
		// キャンドルデータのテーブル
		tableName := GetCandleTableName(symbol, config.Day)

		c := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			time DATETIME PRIMARY KEY NOT NULL,
			open FLOAT,
			close FLOAT,
			high FLOAT,
			low FLOAT,
			volume FLOAT
		)`, tableName)

		_, err := DbConnection.Exec(c)
		if err != nil{
			log.Fatalln(err)
		}

		fmt.Println("Table created successfully!")
}


func init(){
	var err error
	DbConnection, err = sql.Open(config.Config.SQLDriver, config.Config.DbName)
	if err != nil{
		log.Fatalln(err)
	}

	// バックテストにおける売買データを記録するテーブル
	cmd := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		time DATETIME PRIMARY KEY NOT NULL,
		symbol STRING,
		side string,
		price FLOAT,
		size FLOAT
	)`, tableNameSignalEvents)
	_, err = DbConnection.Exec(cmd)
	if err != nil{
		log.Fatalln(err)
	}
	
	// キャンドルデータのテーブル
	for _, duration := range config.Config.Durations{
		tableName := GetCandleTableName(config.Config.Symbol, duration)

		c := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			time DATETIME PRIMARY KEY NOT NULL,
			open FLOAT,
			close FLOAT,
			high FLOAT,
			low FLOAT,
			volume FLOAT
		)`, tableName)
		_, err = DbConnection.Exec(c)
		if err != nil{
			log.Fatalln(err)
		}
	}

	// 取引アルゴリズムのインディケーターを保存するdb
	paramsTableName := fmt.Sprintf("%s_params", config.Config.Symbol)

	cmd = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		time DATETIME UNIQUE,
		emaEnable BOOL,
		emaPeriod1 INTEGER,
		emaPeriod2 INTEGER,
		bbEnable BOOL,
		bbN INTEGER,
		bbK FLOAT,
		rsiEnable BOOL,
		rsiPeriod INTEGER,
		rsiBuyThread FLOAT,
		rsiSellThread FLOAT
	)`, paramsTableName)
	_, err = DbConnection.Exec(cmd)
	if err != nil{
		log.Fatalln(err)
	}

	rankingTableName := GetRankingTableName(config.Config.Symbol)

	cmd = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		time DATETIME,
		name STRING,
		ranking INTEGER,
		performance FLOAT
	)`, rankingTableName)
	_, err = DbConnection.Exec(cmd)
	if err != nil{
		log.Fatalln(err)
	}


}