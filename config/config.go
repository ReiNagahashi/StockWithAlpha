package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/ini.v1"
)

type ConfigList struct{
	ApiKey 			string
	LogFile 		string
	Symbol 			string
	Duration 		string
	DbName 			string
	SQLDriver 		string
	Port 			int
}


var Config ConfigList


func init(){
	cfg, err := ini.Load("config.ini")
	if err != nil{
		log.Printf("Failed to read file: %v", err)
		os.Exit(1)
	}

	Config = ConfigList{
		ApiKey: cfg.Section("alpha_vantage").Key("api_key").String(),
		LogFile: cfg.Section("stockWithAlpha").Key("log_file").String(),
		Symbol: cfg.Section("stockWithAlpha").Key("symbol").String(),
		Duration: time.DateTime,
		DbName: cfg.Section("db").Key("name").String(),
		SQLDriver: cfg.Section("db").Key("driver").String(),
		Port: cfg.Section("web").Key("port").MustInt(),

	}
}