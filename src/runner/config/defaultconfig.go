package config

import (
	"encoding/base64"
)

var DefaultConfig = `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`

func init() {
	decoded, err := base64.StdEncoding.DecodeString(DefaultConfig)
	if err != nil {
		panic("Can't decode base64 encoded encrypted config")
	}
	DefaultConfig = string(decoded)

}
