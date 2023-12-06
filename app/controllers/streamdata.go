package controllers

import (
	"log"
	"stock-with-alpha/alpha"
	"stock-with-alpha/config"
)


func StreamIngestionData(){
	// getTicker関数で取得したtickerをCreateCandle関数を実行してデータに書き込む　
	apiClient := alpha.New(config.Config.ApiKey)
	err := apiClient.GetWeeklyTicker(config.Config.Symbol, "TIME_SERIES_WEEKLY", config.Config.Durations["week"])
	if err != nil{
		log.Println("Failed to ingestion data for weekly...")
	}

	err = apiClient.GetDailyTicker(config.Config.Symbol, "TIME_SERIES_DAILY", config.Config.Durations["day"])
	if err != nil{
		log.Println("Failed to ingestion data for daily...")
	}
	
	// TODO
}