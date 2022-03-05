package utils

import "github.com/Arriven/db1000n/src/logs"

func PanicHandler() {
	if err := recover(); err != nil {
		logs.Default.Warning("caught panic: %v", err)
	}
}
