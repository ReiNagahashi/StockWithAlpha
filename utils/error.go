package utils

import "log"

func ErrorHandler(errChan <-chan error){
	for err := range errChan{
		if err != nil{
			log.Println(err)
		}
	}
}