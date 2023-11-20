package main

import (
	"fmt"
	// "log"
	"stock-with-alpha/alpha"
	"stock-with-alpha/config"
	"stock-with-alpha/utils"
)

func main(){
	utils.LoggingSettings(config.Config.LogFile)
	apiClient := alpha.New(config.Config.ApiKey)
	ticker, _ := apiClient.GetTicker("BA", "GLOBAL_QUOTE")
	fmt.Println(ticker)

}

