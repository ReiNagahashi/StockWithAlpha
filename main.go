package main

import (
	"stock-with-alpha/app/controllers"
	"stock-with-alpha/config"
	"stock-with-alpha/utils"
)

func main(){
	utils.LoggingSettings(config.Config.LogFile)
	controllers.StreamIngestionData()
	controllers.StartWebServer()
}
