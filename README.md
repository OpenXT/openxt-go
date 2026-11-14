# openxt-go

Go libraries and commands for OpenXT userspace. The module path is
`github.com/openxt/openxt-go`. The license is BSD 3-Clause.

## Packages

`argo` talks to the Argo character devices `/dev/argo_stream` and
`/dev/argo_dgram`. `Dial` and `Listen` need those devices. The default build
issues the Argo ioctls directly. `go build -tags libargo` uses libargo instead
and requires `libargo.h` and `-largo`.

`argo/dbus` opens a D-Bus connection over Argo. `ConnectPlatformBus` uses
`argo:domain=0,port=5555` unless `DBUS_SYSTEM_BUS_ADDRESS` is set. `domid` is
accepted as an alias for `domain`.

`db` is a client for the OpenXT database service `com.citrix.xenclient.db`.
`ReadBinary` reads a file-backed binary node. There is no `write_binary`
method.

`ioctl` encodes Linux ioctl command numbers.

## Commands

`make` builds `argo-nc`, `db-cmd`, and `dbus-send` into `./bin`. `make test`
runs `go test ./...`. `make vet` runs `go vet ./...`.

`argo-nc` copies bytes over an Argo stream. `-l` listens. `-d` and `-p` select
the destination domain and port.

`db-cmd` reads and writes the database. `-p` uses the platform bus. Without
it, the command uses the system bus.

`dbus-send` sends a signal or a method call. `--platform` uses the Argo
platform bus. `--type` is `signal` or `method_call`.

Live Argo and the platform bus require the character devices. The unit tests
do not.
