// Package utils [general utility functions for the db1000n app]
package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// PanicHandler just stub it in the beginning of every major module invocation to prevent single module failure from crashing the whole app
func PanicHandler(logger *zap.Logger) {
	if err := recover(); err != nil {
		logger.Error("caught panic, recovering", zap.Any("err", err))
	}
}

// GetEnvStringDefault returns environment variable or default value if no env varible is present
func GetEnvStringDefault(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return value
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

func NonNilDurationOrDefault(d *time.Duration, dflt time.Duration) time.Duration {
	if d != nil {
		return *d
	}

	return dflt
}

func NonNilIntOrDefault(i *int, dflt int) int {
	if i != nil {
		return *i
	}

	return dflt
}

// Decode is an alias to a mapstructure.NewDecoder({Squash: true}).Decode()
// with WeaklyTypedInput set to true and MatchFunc that only compares aplhanumeric sequence in field names
func Decode(input interface{}, output interface{}) error {
	filter := func(r rune) rune {
		if ('a' <= r && r <= 'z') ||
			('A' <= r && r <= 'Z') ||
			('0' <= r && r <= '9') {
			return r
		}

		return -1
	}

	matchName := func(lhs, rhs string) bool {
		return strings.EqualFold(strings.Map(filter, lhs), strings.Map(filter, rhs))
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Squash: true, WeaklyTypedInput: true, MatchName: matchName, Result: output})
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func Unmarshal(input []byte, output interface{}, format string) error {
	switch format {
	case "", "json", "yaml":
		if err := yaml.Unmarshal(input, output); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown config format: %v", format)
	}

	return nil
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	}

	log.Printf("Please open %s", url)
}
