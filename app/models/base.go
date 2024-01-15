package models

import (
	"database/sql"
	"fmt"
	"log"
	"stock-with-alpha/config"
	"strings"
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


// キャンドルテーブル(日足)の名前を全て取得
func GetCandleTableNames() [][]string{
	db, err := sql.Open(config.Config.SQLDriver, config.Config.DbName)
	if err != nil{
		log.Fatalln(err)
	}

	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil{
		log.Fatalln(err)
	}

	defer rows.Close()

	var tableNames [][]string
	for rows.Next(){
		var tableName string
		
		err := rows.Scan(&tableName)
		if err != nil{
			log.Fatal(err)
		}

		symbol := strings.Split(tableName, "_")
		if len(symbol) > 1 && symbol[1] == "24h0m0s"{
			// テーブル名をもとにキャンドルデータのnameを取ってくる
			idToFetch := 1 
			// SQLクエリを構築
			query := fmt.Sprintf("SELECT name FROM %s WHERE id = ?", tableName)
		
			var name string
		
			// QueryRowを使用して1つのレコードを取得
			err = db.QueryRow(query, idToFetch).Scan(&name)
			if err != nil {
				log.Fatal(err)
			}
			newRow := []string{symbol[0], name}
			tableNames = append(tableNames, newRow)
		}
	}

	return tableNames
}

// キャンドル・ランキング・トレードパラムズテーブルを作成(日足)
func CreateTableBySymbol(symbol, name string){
		// キャンドルデータのテーブル
		tableName := GetCandleTableName(symbol, config.Day)

		c := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time DATETIME NOT NULL,
			name TEXT,
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

	// 取引アルゴリズムのインディケーターを保存するdb
	paramsTableName := fmt.Sprintf("%s_params", symbol)

	cmd := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
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

	rankingTableName := GetRankingTableName(symbol)

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


func DropTableBySymbol(symbol string){
	db, err := sql.Open(config.Config.SQLDriver, config.Config.DbName)
	if err != nil{
		log.Fatalln(err)
	}
	defer db.Close()

	tableName := GetCandleTableName(symbol, config.Day)

	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)

	statement, err := db.Prepare(dropTableSQL)
	if err != nil{
		log.Fatalln(err)
	}

	_, err = statement.Exec()
	if err != nil{
		log.Fatalln(err)
	}

	log.Printf("Table '%s' is deleted successfully", tableName)
}


func init(){
	fmt.Println("Database Initializing...")
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

}