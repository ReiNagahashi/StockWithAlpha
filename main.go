package main

import (
	"log"
	"stock-with-alpha/app/controllers"
	"stock-with-alpha/config"
	"stock-with-alpha/utils"
)

func main(){
	utils.LoggingSettings(config.Config.LogFile)
	controllers.StreamIngestionData()
	log.Println(controllers.StartWebServer())
}


