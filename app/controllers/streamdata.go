package controllers

import (
	"log"
	"stock-with-alpha/alpha"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
)


func StreamIngestionData(){
	c := config.Config
	// ここでAIには日足・週足のどっちで取引させるのかを決める
	tableNames := models.GetCandleTableNames()
	Ai := NewAI(config.Day, c.DataLimit, c.UsePercent, c.StopLimitPercent, c.BackTest)

	for i := 0; i < len(tableNames); i++{
		symbol := tableNames[i][0]
		product_name := tableNames[i][1]
		// getTicker関数で取得したtickerをCreateCandle関数を実行してデータに書き込む　
		apiClient := alpha.New(c.ApiKey)

		err := apiClient.GetDailyTicker(symbol, product_name, "TIME_SERIES_DAILY", c.Durations["day"])
		if err != nil{
			log.Println("Failed to ingestion data for daily...")
		}

		Ai.Trade(symbol, product_name)
	}
}