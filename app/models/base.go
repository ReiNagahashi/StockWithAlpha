package models

import (
	"database/sql"
	"fmt"
	"log"
	"stock-with-alpha/config"
	_ "github.com/mattn/go-sqlite3"
)

const (
	tableNameSignalEvents = "signal_events"
)

var DbConnection *sql.DB


func GetCandleTableName(symbol string) string{
	return fmt.Sprintf("%s_%s", symbol, "date")
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
	DbConnection.Exec(cmd)
	
	// 日毎のキャンドルデータのテーブル
	tableName := GetCandleTableName(config.Config.Symbol)
	c := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		time DATETIME PRIMARY KEY NOT NULL,
		open FLOAT,
		close FLOAT,
		high FLOAT,
		low FLOAT,
		volume FLOAT
	)`, tableName)
	DbConnection.Exec(c)

}