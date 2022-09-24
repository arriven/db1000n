package utils

import (
	"bytes"
	"io"
	"runtime"
	"strings"
	"sync"

	"filippo.io/age"
)

// EncryptionKeys random 32 byte key encoded into base64 string. Used by default for configs
var EncryptionKeys = `/45pB920B6DFNwCB/n4rYUio3AVMawrdtrFnjTSIzL4=`

var ProtectedKeys = ``

// decryption takes a bunch of RAM to generate scrypt identity
// we don't do decryption in hot paths so it's better to only allow one thread doing decryption at a time to avoi OOM
var decryptMutex sync.Mutex

const (
	encryptionKeyEnvName = `ENCRYPTION_KEYS`
	keySeparator         = `&`
)

type encryptionKey struct {
	key       string
	protected bool // indicates that the content encrypted by this key shouldn't be logged anywhere
}

// getEncryptionKeys returns list of encryption keys from ENCRYPTION_KEYS env variable name or default value
func getEncryptionKeys() []encryptionKey {
	keysString := GetEnvStringDefault(encryptionKeyEnvName, EncryptionKeys)
	if keysString != EncryptionKeys {
		// if user specified own keys, add default at end to be sure that it always used too
		// to avoid manual copy/join default key to new
		keysString = keysString + keySeparator + EncryptionKeys
	}

	// +1 to allocate for case if no separator and list contains key itself
	// otherwise we just allocate +1 struct for string slice that stores just 2 int fields
	// that is not a lot
	output := make([]encryptionKey, 0, strings.Count(keysString, keySeparator)+strings.Count(ProtectedKeys, keySeparator)+1)

	for _, key := range strings.Split(keysString, keySeparator) {
		if key != "" {
			output = append(output, encryptionKey{key: key})
		}
	}

	for _, key := range strings.Split(ProtectedKeys, keySeparator) {
		if key != "" {
			output = append(output, encryptionKey{key: key, protected: true})
		}
	}

	return output
}

// IsEncrypted returns true if cfg encrypted with age tool (https://github.com/FiloSottile/age)
func IsEncrypted(cfg []byte) bool {
	return bytes.Contains(cfg, []byte(`age-encryption`))
}

// Decrypt decrypts config using EncryptionKeys
func Decrypt(cfg []byte) (result []byte, protected bool, err error) {
	decryptMutex.Lock()
	defer decryptMutex.Unlock()

	// iterate over all keys and return on first success decryption
	for _, key := range getEncryptionKeys() {
		result, err = decrypt(cfg, key.key)

		runtime.GC() // force GC to decrease memory usage

		if err != nil {
			continue
		}

		return result, key.protected, nil
	}

	return nil, false, err
}

func decrypt(cfg []byte, key string) ([]byte, error) {
	identity, err := age.NewScryptIdentity(key)
	if err != nil {
		return nil, err
	}

	decryptedReader, err := age.Decrypt(bytes.NewReader(cfg), identity)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(decryptedReader)
}
