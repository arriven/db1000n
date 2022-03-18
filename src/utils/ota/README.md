# Over-the-air updates

Lots of maintainers run their needles on a bare metal machines.
As long as this project is so frequently updated, it might be
a good idea to let them update it without the hassle.

- [x] Enabled automatic time-based version check
- [x] Enabled application self-restart after it downloaded the update

## Description

Support for the application self-update by downloading the latest release from the official repository.

```text
Stay strong, be the first in line!
```

The version should be embedded in the binary during the build, see the `build`
target in the Makefile.

## Usage

### Available flags

```bash
  -enable-self-update
        Enable the application automatic updates on the startup
  -restart-on-update
        Allows application to restart upon the successful update (ignored if auto-update is disabled) (default true)
  -self-update-check-frequency duration
        How often to run auto-update checks (default 24h0m0s)
  -skip-update-check-on-start
        Allows to skip the update check at the startup (usually set automatically by the previous version) (default false)
```

The default behavior if the self-update enabled:

```bash
* Check for the update
* If update is available - download it
   * If auto-restart is enabled
      * Notify the user that a newer version is available
      * Fork-Exec a new process (will have a different PID), add a flag to skip the version check upon startup
      * Stop the current process
   * If auto-restart is disabled - notify user that manual restart is required
* If update is NOT available - schedule the next check
```

### Examples

To update your needle, start it with a flag `-enable-self-update`

```sh
./db1000n -enable-self-update
```

#### Advanced options

Start the needle with the **self-update & self-restart**

```bash
$ ./db1000n -enable-self-update
0000/00/00 00:00:00 main.go:82: DB1000n [Version: v0.6.4][PID=75259]
0000/00/00 00:00:00 main.go:166: Running a check for a newer version...
0000/00/00 00:00:00 main.go:176: Newer version of the application is found [0.7.0]
0000/00/00 00:00:00 main.go:177: What's new:
* Added some great improvements
* Added some spectacular bugs
0000/00/00 00:00:00 main.go:180: Auto restart is enabled, restarting the application to run a new version
0000/00/00 00:00:00 restart.go:45: new process has been started successfully [old_pid=75259,new_pid=75262]

# NOTE: Process 75259 exited, Process 75262 has started with a flag to skip version check on the startup

0000/00/00 00:00:00 main.go:82: DB1000n [Version: v0.7.0][PID=75262]
0000/00/00 00:00:00 main.go:155: Version update on startup is skipped, next update check is scheduled in 24h0m0s
```

Start the needle with the self-update but do not restart the process upon update (`systemd` friendly)

```bash
$ ./db1000n -enable-self-update -self-update-check-frequency=5m -restart-on-update=false
0000/00/00 00:00:00 main.go:82: DB1000n [Version: v0.6.4][PID=75320]
0000/00/00 00:00:00 main.go:166: Running a check for a newer version...
0000/00/00 00:00:00 main.go:176: Newer version of the application is found [0.7.0]
0000/00/00 00:00:00 main.go:177: What's new:
* Added some great improvements
* Added some spectacular bugs
0000/00/00 00:00:00 main.go:191: Auto restart is disabled, restart the application manually to apply changes!
```

## References

1. Graceful restart with zero downtime for TCP connection - [https://github.com/Scalingo/go-graceful-restart-example](https://github.com/Scalingo/go-graceful-restart-example)
1. Graceful restart with zero downtime for TCP connection (two variants) [https://github.com/rcrowley/goagain](https://github.com/rcrowley/goagain)
1. Graceful restart with zero downtime for TCP connection (alternative) [https://grisha.org/blog/2014/06/03/graceful-restart-in-golang](https://grisha.org/blog/2014/06/03/graceful-restart-in-golang)
