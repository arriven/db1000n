package utils

import "log"

func PanicHandler() {
	if err := recover(); err != nil {
		log.Printf("caught panic: %v", err)
	}
}
