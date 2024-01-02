package controllers

import (
	"log"
	"stock-with-alpha/alpha"
	"stock-with-alpha/config"
)


func StreamIngestionData(){
	c := config.Config
	// ここでAIには日足・週足のどっちで取引させるのかを決める
	ai := NewAI(c.Symbol, config.Day, c.DataLimit, c.UsePercent, c.StopLimitPercent, c.BackTest)
	// getTicker関数で取得したtickerをCreateCandle関数を実行してデータに書き込む　
	apiClient := alpha.New(c.ApiKey)
	err := apiClient.GetWeeklyTicker(c.Symbol, "TIME_SERIES_WEEKLY", c.Durations["week"])
	if err != nil{
		log.Println("Failed to ingestion data for weekly...")
	}

	err = apiClient.GetDailyTicker(c.Symbol, "TIME_SERIES_DAILY", c.Durations["day"])
	if err != nil{
		log.Println("Failed to ingestion data for daily...")
	}
	
	ai.Trade()
}