# ctld Exporter
Prometheus exporter for [ctld](https://www.freebsd.org/cgi/man.cgi?query=ctld&sektion=8) (CAM Target Layer / iSCSI target daemon)
This exporter uses CGO to get the data (part of the C code is taken from [ctlstat](https://www.freebsd.org/cgi/man.cgi?query=ctlstat&sektion=8) and [ctladm](https://www.freebsd.org/cgi/man.cgi?query=ctladm&sektion=8)). It is planned to remove CGO use in future version.

The exported metrics are the data available using [ctlstat](https://www.freebsd.org/cgi/man.cgi?query=ctlstat&sektion=8) and matched against
target and LUN id within that target.

## Installation
You can build the latest version using Go v1.11+ on FreeBSD via `go get`:
```
go get -u github.com/Gandi/ctld_exporter
```

## Usage
```
usage: ctld_exporter [<flags>]

Flags:
  -h, --help              Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9572"
                          Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"
                          Path under which to expose metrics.
      --log.level="info"  Only log messages with the given severity or above. Valid levels: [debug, info,
                          warn, error, fatal]
      --log.format="logger:stderr"
                          Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or
                          "logger:stdout?json=true"
      --version           Show application version.
```

## Caveats
This exporter NEEDS to be build on FreeBSD with access to core source (expected location is `/usr/src/`).
The exporter must have r/w access to `/dev/cam/ctl` thus needing root privileges.

