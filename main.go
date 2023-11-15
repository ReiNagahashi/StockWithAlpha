package main

import (
	"fmt"
	"stock-with-alpha/config"
)

func main(){
	fmt.Println(config.Config.ApiKey)
}