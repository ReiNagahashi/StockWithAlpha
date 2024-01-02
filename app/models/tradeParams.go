package models

import (
	"fmt"
	"log"
	"stock-with-alpha/config"
	"time"
)

type TradeParams struct{
	EmaEnable 		bool
	EmaPeriod1 		int
	EmaPeriod2 		int
	BbEnable 		bool
	BbN 			int
	BbK 			float64
	RsiEnable 		bool
	RsiPeriod 		int
	RsiBuyThread 	float64
	RsiSellThread 	float64
	Time 			time.Time
}


func GetTradeParamsTableName(symbol string) string{
	return fmt.Sprintf("%s_params", symbol)
}


func (tp *TradeParams) CreateTradeParams() error{
	currentParams:= GetTradeParams()

	if currentParams != nil{
		return nil
	}

	cmd := fmt.Sprintf("INSERT INTO %s (emaEnable, emaPeriod1, emaPeriod2, bbEnable, bbN, bbK, rsiEnable, rsiPeriod, rsiBuyThread, rsiSellThread, time) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", GetTradeParamsTableName(config.Config.Symbol))
	_, err := DbConnection.Exec(cmd, tp.EmaEnable, tp.EmaPeriod1, tp.EmaPeriod2, tp.BbEnable, tp.BbN, tp.BbK, tp.RsiEnable, tp.RsiPeriod, tp.RsiBuyThread, tp.RsiSellThread, tp.Time)
	if err != nil{
		return err
	}

	return err
}


func GetTradeParams() *TradeParams{
	tableName := GetTradeParamsTableName(config.Config.Symbol)
	cmd := fmt.Sprintf(`SELECT emaEnable, emaPeriod1, emaPeriod2, bbEnable, bbN, bbK, rsiEnable, rsiPeriod, rsiBuyThread, rsiSellThread, time FROM %s WHERE time = ?`, tableName)
	row := DbConnection.QueryRow(cmd, time.Now().Format("2008-01-01"))
	
	var params TradeParams
	err := row.Scan(&params.EmaEnable, &params.EmaPeriod1, &params.EmaPeriod2, &params.BbEnable, &params.BbN, &params.BbK, &params.RsiEnable, &params.RsiPeriod,&params.RsiBuyThread, &params.RsiSellThread, &params.Time)
	if err != nil{
		return nil
	}
	
	fmt.Println("You already have the data")

	return &params
}


type Ranking struct{
	Name string
	Performance float64
	Ranking int
	IsEnable bool
	Time time.Time
}


func GetRankingTableName(symbol string) string{
	return fmt.Sprintf("%s_ranking", symbol)
}


func (r *Ranking) CreateRanking() (bool, error){
	parsedDate, err := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	if err != nil {
		log.Fatalln("Error parsing date:", err)
	}
	currentRanking := GetRanking(parsedDate, r.Name)

	if currentRanking != nil{
		return false, nil
	}

	cmd := fmt.Sprintf("INSERT INTO %s (name, performance, ranking, time) VALUES(?, ?, ?, ?)", GetRankingTableName(config.Config.Symbol))
	_, err = DbConnection.Exec(cmd, r.Name, r.Performance, r.Ranking, parsedDate)
	if err != nil{
		return false, err
	}

	return true, err
}


func GetRanking(date time.Time, name string) *Ranking{
	tableName := GetRankingTableName(config.Config.Symbol)

	cmd := fmt.Sprintf(`SELECT name, performance, ranking, time FROM %s WHERE time = ? AND name = ?`, tableName)
	row := DbConnection.QueryRow(cmd, date, name)
	
	var ranking Ranking
	err := row.Scan(&ranking.Name, &ranking.Performance, &ranking.Ranking, &ranking.Time)
	if err != nil{
		return nil
	}
	
	fmt.Println("You already have the data")

	return &ranking
}

