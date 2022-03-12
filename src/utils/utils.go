// Package utils [general utility functions for the db1000n app]
package utils

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
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

// GetEnvIntDefault returns environment variable or default value if no env varible is present
func GetEnvIntDefault(key string, defaultValue int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("GetEnvIntDefault[%s]: %v", key, err)
		return defaultValue
	}
	return v
}

// GetEnvBoolDefault returns environment variable or default value if no env varible is present
func GetEnvBoolDefault(key string, defaultValue bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	v, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("GetEnvBoolDefault[%s]: %v", key, err)
		return defaultValue
	}
	return v
}

// GetEnvDurationDefault returns environment variable or default value if no env varible is present
func GetEnvDurationDefault(key string, defaultValue time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	v, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("GetEnvBoolDefault[%s]: %v", key, err)
		return defaultValue
	}
	return v
}

// Decode is an alias to a mapstructure.NewDecoder({Squash: true}).Decode()
func Decode(input interface{}, output interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Squash: true, Result: output})
	if err != nil {
		log.Printf("Error parsing job config: %v", err)
		return err
	}
	return decoder.Decode(input)
}
