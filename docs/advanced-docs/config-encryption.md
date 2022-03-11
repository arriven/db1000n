# Encryption

## How application uses configuration files according to encryption

Encryption is done using [age](https://github.com/FiloSottile/age) as CLI and golang library for encryption/decryption.
Under the hood it uses ChaCha20+Poly1305 as AEAD encryption and store encrypted files with pre-defined header `age-encryption.org/v1`

When application loads config from any source, at first it check is it encrypted (has header).
If it is encrypted then tries to decrypt using keys stored in `ENCRYPTION_KEYS` env variable and iteratively try every key
until successful decryption or skip config. If it wasn't encrypted then return as is.

You can encrypt every config file with separate keys and run application specifying all keys at once as a list with separator: `&`.

Separator `&` was chosen to allow pass base64 strings as keys that can encode random keys (binary values).

## Encrypt new default config

### Prepare JSON config

Let's take next config for example:

```json
{
  "jobs": [
    {
      "type": "slow-loris",
      "args": {
        "address": "80.87.198.26:53",
        "ContentLength": 1000,
        "DialWorkersCount": 1,
        "RampUpInterval": 1,
        "SleepInterval": 1000,
        "DurationSeconds": 1000,
        "Path": "https://meduza.io"
      }
    }
  ]
}
```

and save it as `/home/user/config.json`

### Prepare key for encryption

Use some of stored in `src/utils/crypto.go.ENCRYPTION_KEYS` or generate new one. For example lets use:
`some long password to encrypt config`

### Encrypt config

#### Using make

```sh
make DEFAULT_CONFIG=/home/user/config.json encrypt_config

Enter passphrase (leave empty to autogenerate a secure one): some long password to encrypt config
Confirm passphrase: some long password to encrypt config
Saved in file: /tmp/fileMx6JFo
Save value as env variable:
export DEFAULT_CONFIG_VALUE='YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCAwS1pOUXlLM004L1NxcDRwL21CaUl3IDE4Cjc1cEZtcmZZZHJieFRvd0hhM0RVVVMxb2VMY0RmQzBkUkJQMXE2UWdyMUEKLS0tIHlSK1VFMkNOSHovbzRqUlJ3RW5VclphRy9TU0NQSG8vMzJUZ1c4RUozZncKD6zE4MONozWBfQYn9HG31DW100o2oFpn6iACQAvCDyXkgSeuQtRFjPwCIW5q2Dltq7Srkc8b81/ZynC59uqkmDJefGyNPzTk3ilRl6wcLOhCP1TD7YtCtZ/7ZpoGpNMiDD6XhKnOmz10sBSy1SXt54+zFVcuQ1ITRi4E2WmiFRjTa8T+ZMwurW+F+iwOu6+z8/0sKQaG5SrKA74GI9D6iRQnqiPg2Abr97Vq7X2Fjvz2NqFjcB0dD29XijHcLCdXQ1DcI3gx94SdMmmfeU5ub2ArsH/4nA8XlS7YE7BirUihgHD4/KIr52dc+Fst6i7SBH433d/Y3Pmhi89FHY8+sGyPFXNG+SeLLHafcR6bLLGyk0iGa2bZaBqUGovYNojni8KSrLRPXTgCyeNAOS7Gpamwi1Xco7m7nEEmAv9vpEvtOUx83pGBOkgu3oSV0t3jmp+OUvcwMMQ='
```

Encrypted config file saved to `/tmp/fileMx6JFo` file. And can be saved and distributed anywhere.

#### Using age

```sh
age --encrypt -p --output=encrypted_config.json /home/user/config.json
Enter passphrase (leave empty to autogenerate a secure one): some long password to encrypt config
Confirm passphrase: some long password to encrypt config
```

##### Converting to base64

```sh
cat encrypted_config.json | base64 | tr -d '\n'

YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCAwS1pOUXlLM004L1NxcDRwL21CaUl3IDE4Cjc1cEZtcmZZZHJieFRvd0hhM0RVVVMxb2VMY0RmQzBkUkJQMXE2UWdyMUEKLS0tIHlSK1VFMkNOSHovbzRqUlJ3RW5VclphRy9TU0NQSG8vMzJUZ1c4RUozZncKD6zE4MONozWBfQYn9HG31DW100o2oFpn6iACQAvCDyXkgSeuQtRFjPwCIW5q2Dltq7Srkc8b81/ZynC59uqkmDJefGyNPzTk3ilRl6wcLOhCP1TD7YtCtZ/7ZpoGpNMiDD6XhKnOmz10sBSy1SXt54+zFVcuQ1ITRi4E2WmiFRjTa8T+ZMwurW+F+iwOu6+z8/0sKQaG5SrKA74GI9D6iRQnqiPg2Abr97Vq7X2Fjvz2NqFjcB0dD29XijHcLCdXQ1DcI3gx94SdMmmfeU5ub2ArsH/4nA8XlS7YE7BirUihgHD4/KIr52dc+Fst6i7SBH433d/Y3Pmhi89FHY8+sGyPFXNG+SeLLHafcR6bLLGyk0iGa2bZaBqUGovYNojni8KSrLRPXTgCyeNAOS7Gpamwi1Xco7m7nEEmAv9vpEvtOUx83pGBOkgu3oSV0t3jmp+OUvcwMMQ=
```

### Embedding encrypted config as default backup config into binary

Export env variable as it printed:

```sh
export DEFAULT_CONFIG_VALUE='YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCAwS1pOUXlLM004L1NxcDRwL21CaUl3IDE4Cjc1cEZtcmZZZHJieFRvd0hhM0RVVVMxb2VMY0RmQzBkUkJQMXE2UWdyMUEKLS0tIHlSK1VFMkNOSHovbzRqUlJ3RW5VclphRy9TU0NQSG8vMzJUZ1c4RUozZncKD6zE4MONozWBfQYn9HG31DW100o2oFpn6iACQAvCDyXkgSeuQtRFjPwCIW5q2Dltq7Srkc8b81/ZynC59uqkmDJefGyNPzTk3ilRl6wcLOhCP1TD7YtCtZ/7ZpoGpNMiDD6XhKnOmz10sBSy1SXt54+zFVcuQ1ITRi4E2WmiFRjTa8T+ZMwurW+F+iwOu6+z8/0sKQaG5SrKA74GI9D6iRQnqiPg2Abr97Vq7X2Fjvz2NqFjcB0dD29XijHcLCdXQ1DcI3gx94SdMmmfeU5ub2ArsH/4nA8XlS7YE7BirUihgHD4/KIr52dc+Fst6i7SBH433d/Y3Pmhi89FHY8+sGyPFXNG+SeLLHafcR6bLLGyk0iGa2bZaBqUGovYNojni8KSrLRPXTgCyeNAOS7Gpamwi1Xco7m7nEEmAv9vpEvtOUx83pGBOkgu3oSV0t3jmp+OUvcwMMQ='
```

This base64 value is encrypted config with specified password (hashed with scrypt against brute force).
Build new binary with new config:

```sh
make build_encrypted

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X 'github.com/Arriven/db1000n/src/runner/config.DefaultConfig=YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCAwS1pOUXlLM004L1NxcDRwL21CaUl3IDE4Cjc1cEZtcmZZZHJieFRvd0hhM0RVVVMxb2VMY0RmQzBkUkJQMXE2UWdyMUEKLS0tIHlSK1VFMkNOSHovbzRqUlJ3RW5VclphRy9TU0NQSG8vMzJUZ1c4RUozZncKD6zE4MONozWBfQYn9HG31DW100o2oFpn6iACQAvCDyXkgSeuQtRFjPwCIW5q2Dltq7Srkc8b81/ZynC59uqkmDJefGyNPzTk3ilRl6wcLOhCP1TD7YtCtZ/7ZpoGpNMiDD6XhKnOmz10sBSy1SXt54+zFVcuQ1ITRi4E2WmiFRjTa8T+ZMwurW+F+iwOu6+z8/0sKQaG5SrKA74GI9D6iRQnqiPg2Abr97Vq7X2Fjvz2NqFjcB0dD29XijHcLCdXQ1DcI3gx94SdMmmfeU5ub2ArsH/4nA8XlS7YE7BirUihgHD4/KIr52dc+Fst6i7SBH433d/Y3Pmhi89FHY8+sGyPFXNG+SeLLHafcR6bLLGyk0iGa2bZaBqUGovYNojni8KSrLRPXTgCyeNAOS7Gpamwi1Xco7m7nEEmAv9vpEvtOUx83pGBOkgu3oSV0t3jmp+OUvcwMMQ='" -o main ./main.go
```

Your new binary saved as `main` with new encrypted and embedded default config.
To turn on decryption new config you should pass encryption keys as list of keys separated with `&` symbol:

```sh
export ENCRYPTION_KEYS='some long password to encrypt config&another key'
```

It will override default encryption keys

### Embedding encryption keys into binary

You can embed keys into binary in same way:

```sh
export ENCRYPTION_KEYS='some long password to encrypt config&another key'
make build_encrypted

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X 'github.com/Arriven/db1000n/src/utils.EncryptionKeys=some long password to encrypt config&another key' -X 'github.com/Arriven/db1000n/src/runner/config.DefaultConfig=YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCAwS1pOUXlLM004L1NxcDRwL21CaUl3IDE4Cjc1cEZtcmZZZHJieFRvd0hhM0RVVVMxb2VMY0RmQzBkUkJQMXE2UWdyMUEKLS0tIHlSK1VFMkNOSHovbzRqUlJ3RW5VclphRy9TU0NQSG8vMzJUZ1c4RUozZncKD6zE4MONozWBfQYn9HG31DW100o2oFpn6iACQAvCDyXkgSeuQtRFjPwCIW5q2Dltq7Srkc8b81/ZynC59uqkmDJefGyNPzTk3ilRl6wcLOhCP1TD7YtCtZ/7ZpoGpNMiDD6XhKnOmz10sBSy1SXt54+zFVcuQ1ITRi4E2WmiFRjTa8T+ZMwurW+F+iwOu6+z8/0sKQaG5SrKA74GI9D6iRQnqiPg2Abr97Vq7X2Fjvz2NqFjcB0dD29XijHcLCdXQ1DcI3gx94SdMmmfeU5ub2ArsH/4nA8XlS7YE7BirUihgHD4/KIr52dc+Fst6i7SBH433d/Y3Pmhi89FHY8+sGyPFXNG+SeLLHafcR6bLLGyk0iGa2bZaBqUGovYNojni8KSrLRPXTgCyeNAOS7Gpamwi1Xco7m7nEEmAv9vpEvtOUx83pGBOkgu3oSV0t3jmp+OUvcwMMQ='" -o main ./main.go
```
