package main

import (
	"fmt"
	// "log"
	// "stock-with-alpha/alpha"
	"stock-with-alpha/app/controllers"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
	"stock-with-alpha/utils"
)

func main(){
	utils.LoggingSettings(config.Config.LogFile)
	fmt.Println(models.DbConnection)
	controllers.StreamIngestionData()
}

