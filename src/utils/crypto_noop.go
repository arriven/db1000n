//go:build !encrypted
// +build !encrypted

package utils

import (
	"bytes"
	"fmt"
)

// GetEncryptionKeys returns list of encryption keys from ENCRYPTION_KEYS env variable name or default value
func GetEncryptionKeys() ([]string, error) {
	return nil, fmt.Errorf("encryption not supported")
}

// IsEncrypted returns true if cfg encrypted with age tool (https://github.com/FiloSottile/age)
func IsEncrypted(cfg []byte) bool {
	return bytes.Contains(cfg, []byte(`age-encryption`))
}

// Decrypt decrypts config using EncryptionKeys
func Decrypt(cfg []byte) ([]byte, error) {
	return nil, fmt.Errorf("encryption not supported")
}
