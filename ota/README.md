# Over-the-air updates

Lots of maintainers run their needles on a bare metal machines.
As long as this project is so frequently updated, it might be
a good idea to let them update it without the hassle.

## TODO

_TO BE CONFIRMED WITH THE CODE OWNER!_

- [?] Enable automatic time-based version check
- [?] Enable push-based version check
- [?] Enable application self-restart after it downloaded the update
- [?] Disable OTA updates for needles running inside the Docker container

## Description

Support for the application self-update by downloading
the latest release from the official repository.

```text
Stay strong, be the first in line!
```

The version should be embedded in the binary during the build, see the `build`
target in the Makefile.

## Usage

To update your needle, start it with a flag `-enable-self-update`

```sh
./db1000n -enable-self-update
```

### Update example

```bash
$ make build
CGO_ENABLED=0 go build -ldflags="-s -w -X 'github.com/Arriven/db1000n/ota.Version=v0.6.4'" -o db1000n -a ./main.go
$ ./db1000n -enable-self-update
1970/01/01 00:00:00.00000 main.go:76: DB1000n [Version: v0.6.4]
1970/01/01 00:00:00.00000 ota.go:30: Successfully updated to version 0.6.5
1970/01/01 00:00:00.00000 ota.go:31: Release note:
 ## What's Changed
* User friendly logs by @Arriven in https://github.com/Arriven/db1000n/pull/271
...

**Full Changelog**: https://github.com/Arriven/db1000n/compare/v0.6.4...v0.6.5
1970/01/01 00:00:00.00000 config.go:36: Loading config from "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json"
1970/01/01 00:00:00.00000 config.go:97: New config received, applying
```
