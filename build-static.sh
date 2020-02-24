#!/bin/bash

# This should build a linux amd64 binary on any modern linux system having docker installed and working.

VER=$(git rev-parse --short HEAD)
sed -i "s/const Version = .*/const Version = \"${VER}\"/" models/version.go

# Uncomment to build the apline image first
# docker build -t golang-alpine-build:latest -f Dockerfile.alpine .

# you can remove the --rm option, build it first time and then commit the container to image. Then add --rm back, so next time you dont have to install go library again.
docker run --rm --name golang-alpine-build  -v $(pwd):/work --workdir /work --entrypoint go --env-file ~/.gobuild-linux-cgo golang-alpine-build:latest build --tags "json1 secure_delete" --ldflags '-extldflags "-static" -w -s' -o webnote-linux-amd64-static main.go

ARCH=linux-amd64-static

CDIR=$(pwd)

mkdir /tmp/webnote-$$/webnote-go-bin -p
sudo cp -a assets webnote-linux-amd64-static /tmp/webnote-$$/webnote-go-bin/
cd /tmp/webnote-$$
sudo tar czf $CDIR/webnote-go-bin-${ARCH}.tgz webnote-go-bin
sudo rm -rf /tmp/webnote-$$

echo "The archive ready to be extracted and run is webnote-go-bin-linux-amd64-static.tgz."
