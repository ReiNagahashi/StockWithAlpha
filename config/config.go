package config

import (
	"log"
	"os"
	"gopkg.in/ini.v1"
)

type ConfigList struct{
	ApiKey string
	LogFile string
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
	}
}