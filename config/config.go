package config

import (
	"log"
	"os"
	"time"
	"gopkg.in/ini.v1"
)

type ConfigList struct{
	ApiKey 				string
	LogFile 			string
	Symbol 				string
	Durations 			map[string]time.Duration
	DbName 				string
	SQLDriver 			string
	Port 				int
	
	BackTest 			bool
	UsePercent 			float64
	DataLimit 			int
	StopLimitPercent 	float64
	NumRanking 			int

	SmtpHostName 		string
	SmtpPort 			int
	Email 				string
}


var(
	Config ConfigList
	hour = time.Duration(time.Hour)
	Day = hour * 24
	Week = Day * 7
)


func init(){
	cfg, err := ini.Load("config.ini")
	if err != nil{
		log.Printf("Failed to read file: %v", err)
		os.Exit(1)
	}


	durations := map[string]time.Duration{
		"day": Day,
		"week": Week,
	}

	Config = ConfigList{
		ApiKey: cfg.Section("alpha_vantage").Key("api_key").String(),
		LogFile: cfg.Section("stockWithAlpha").Key("log_file").String(),
		Symbol: cfg.Section("stockWithAlpha").Key("symbol").String(),
		Durations: durations,
		DbName: cfg.Section("db").Key("name").String(),
		SQLDriver: cfg.Section("db").Key("driver").String(),
		Port: cfg.Section("web").Key("port").MustInt(),
		BackTest: cfg.Section("stockWithAlpha").Key("back_test").MustBool(),
		UsePercent: cfg.Section("stockWithAlpha").Key("use_percent").MustFloat64(),
		DataLimit: cfg.Section("stockWithAlpha").Key("data_limit").MustInt(),
		StopLimitPercent: cfg.Section("stockWithAlpha").Key("stop_limit_percent").MustFloat64(),
		NumRanking: cfg.Section("stockWithAlpha").Key("num_ranking").MustInt(),
		SmtpHostName: cfg.Section("stockWithAlpha").Key("host").String(),
		SmtpPort: cfg.Section("stockWithAlpha").Key("port").MustInt(),
		Email: cfg.Section("stockWithAlpha").Key("email").String(),
	}
}