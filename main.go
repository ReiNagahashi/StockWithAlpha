package main

// import (
// 	"stock-with-alpha/app/controllers"
// 	"stock-with-alpha/config"
// 	"stock-with-alpha/utils"
// )

// func main(){
// 	utils.LoggingSettings(config.Config.LogFile)
// 	controllers.StreamIngestionData()
// 	// controllers.StartWebServer()
// }


import (
	"fmt"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
	"time"
)

func main(){
	s := models.NewSignalEvents()
	df, _ := models.GetAllCandle("IBM", config.Day, 10)
	c1 := df.Candles[0]
	c2 := df.Candles[5]

	s.Buy("IBM", c1.DateTime.UTC(), c1.Close, 1.0, true)
	s.Sell("IBM", c2.DateTime.UTC(), c2.Close, 1.0, true)

	fmt.Println(models.GetSignalEventsByCount(1))
	fmt.Println(models.GetSignalEventsAfterTime(c1.DateTime))
	fmt.Println(s.CollectAfter(time.Now().UTC()))
	fmt.Println(s.CollectAfter(c1.DateTime))
}