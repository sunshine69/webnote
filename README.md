## Quick build

forget about these build script using quirky docker compose build :). If you run ubuntu or any other linux
distro you can run

```
go build -trimpath -ldflags="-X main.version=v1.0 -extldflags=-static -w -s" --tags "osusergo,netgo,sqlite_stat4,sqlite_foreign_keys,sqlite_json"
```

and the binary will be produced `webnote-go`.

Run `./webnote-go -h` for quick help to setup.

If you copy the binary to other location make sure you copy the directory `asset` as well.
