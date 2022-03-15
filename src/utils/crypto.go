package utils

import (
	"bytes"
	"io/ioutil"
	"strings"
	"sync"

	"filippo.io/age"
)

// EncryptionKeys random 32 byte key encoded into base64 string. Used by default for configs
var EncryptionKeys = `/45pB920B6DFNwCB/n4rYUio3AVMawrdtrFnjTSIzL4=`

// decryption takes a bunch of RAM to generate scrypt identity
// we don't do decryption in hot paths so it's better to only allow one thread doing decryption at a time to avoi OOM
var decryptMutex sync.Mutex

const (
	encryptionKeyEnvName = `ENCRYPTION_KEYS`
	keySeparator         = `&`
)

// GetEncryptionKeys returns list of encryption keys from ENCRYPTION_KEYS env variable name or default value
func GetEncryptionKeys() ([]string, error) {
	keysString := GetEnvStringDefault(encryptionKeyEnvName, EncryptionKeys)
	if keysString != EncryptionKeys {
		// if user specified own keys, add default at end to be sure that it always used too
		// to avoid manual copy/join default key to new
		keysString = keysString + keySeparator + EncryptionKeys
	}

	// +1 to allocate for case if no separator and list contains key itself
	// otherwise we just allocate +1 struct for string slice that stores just 2 int fields
	// that is not a lot
	output := make([]string, 0, strings.Count(keysString, keySeparator)+1)

	for _, key := range strings.Split(keysString, keySeparator) {
		if key != "" {
			output = append(output, key)
		}
	}

	return output, nil
}

// IsEncrypted returns true if cfg encrypted with age tool (https://github.com/FiloSottile/age)
func IsEncrypted(cfg []byte) bool {
	return bytes.Contains(cfg, []byte(`age-encryption`))
}

// Decrypt decrypts config using EncryptionKeys
func Decrypt(cfg []byte) ([]byte, error) {
	keys, err := GetEncryptionKeys()
	if err != nil {
		return nil, err
	}

	decryptMutex.Lock()
	defer decryptMutex.Unlock()

	var lastErr error
	// iterate over all keys and return on first success decryption
	for _, key := range keys {
		identity, err := age.NewScryptIdentity(key)
		if err != nil {
			lastErr = err

			continue
		}

		decryptedReader, err := age.Decrypt(bytes.NewReader(cfg), identity)
		if err != nil {
			lastErr = err

			continue
		}

		return ioutil.ReadAll(decryptedReader)
	}

	return nil, lastErr
}
