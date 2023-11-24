package controllers

import (
	// "fmt"
	"log"
	"stock-with-alpha/alpha"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
	"strings"
	"time"
)

func StreamIngestionData(){
	// getTicker関数で取得したtickerをCreateCandle関数を実行してデータに書き込む　
	apiClient := alpha.New(config.Config.ApiKey)
	ticker,err := apiClient.GetTicker(config.Config.Symbol, "GLOBAL_QUOTE")
	if err != nil{
		log.Println("Failed to get ticker")
	}

	log.Printf("action=StreamIngestionData, %v", ticker)

	// 月名の英語表記マップ
	monthMap := map[string]string{
		"01": "Jan", "02": "Feb", "03": "Mar", "04": "Apr",
		"05": "May", "06": "Jun", "07": "Jul", "08": "Aug",
		"09": "Sep", "10": "Oct", "11": "Nov", "12": "Dec",
	}

	// 変換対象の日付
	dateString := ticker.GlobalQuote.LatestTradingDay
	dateParts := strings.Split(dateString, "-")

	dateString = dateParts[0] + "-" + monthMap[dateParts[1]] + "-" + dateParts[2]

	t, err := time.Parse("2006-Jan-02", dateString)
	if err != nil{
		log.Fatalln("Failed to parse")
	}

	isCreated := models.CreateCandle(*ticker, ticker.GlobalQuote.Symbol, t)
	if isCreated == true{
		//TODO
	}
}