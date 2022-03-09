// Package utils [general utility functions for the db1000n app]
package utils

import (
	"log"
	"os"
)

// PanicHandler just stub it in the beginning of every major module invocation to prevent single module failure from crashing the whole app
func PanicHandler() {
	if err := recover(); err != nil {
		log.Printf("caught panic: %v", err)
	}
}

// GetEnvStringDefault returns environment variable or default value if no env varible is present
func GetEnvStringDefault(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	return defaultValue
}
