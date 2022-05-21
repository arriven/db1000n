package config

import (
	"encoding/base64"
)

// DefaultConfig is the config embedded into the app that it will use if not able to fetch any other config
//nolint:lll // Makes no sense splitting this into multiple lines
var DefaultConfig = ``

func init() {
	decoded, err := base64.StdEncoding.DecodeString(DefaultConfig)
	if err != nil {
		panic("Can't decode base64 encoded encrypted config")
	}

	DefaultConfig = string(decoded)
}
