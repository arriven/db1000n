package utils

import "os"

func GetEnvStringDefault(key, value string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return value
}
