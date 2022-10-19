## Quick build

forget about these build script using quirky docker compose build :). If you run ubuntu or any other linux
distro you can run

```
go build -trimpath --tags "json1 fts5 secure_delete osusergo netgo sqlite_stat4 sqlite_foreign_keys" -ldflags="-X main.version=v1.0 -extldflags=-static -w -s"
```

and the binary will be produced `webnote-go`.

Run `./webnote-go -h` for quick help to setup.

If you copy the binary to other location make sure you copy the directory `asset` as well.

To see webnote apps in the folder `app` and in main route in `main.go`. To create new app just add new file in `app` and route in `main.go`. The GUI can be html file in `assets/media/html`. See an example of a onetime password share link app in [here](https://gonote.duckdns.org:6919/assets/media/html/onetime-secret.html)

If the new app needs to save data into the database you can use the table `note` and store the content as a json string. Then use the sqlite json extention in your app to select/update/query.

## Docker image to run on azure webapp

Run latest image: `docker run --rm -p 8080:8080 -v $(pwd):/home stevekieu/webnote-azure-app:v1.12`. Then access http://localhost:8080

The image is built using command
`docker build -t stevekieu/webnote-azure-app:v1.12 --build-arg APP_VERSION=v1.12 -f Dockerfile.azure-app-svc .`

You can build yourself tag different image and play around.