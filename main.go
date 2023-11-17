package main

import (
	// "fmt"
	"log"
	"stock-with-alpha/config"
	"stock-with-alpha/utils"
)

func main(){
	utils.LoggingSettings(config.Config.LogFile)
	// fmt.Println(config.Config.ApiKey)
	log.Println("test")
}