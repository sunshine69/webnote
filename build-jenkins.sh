#!/bin/sh

# Run this on android chroot env to build for arm. I ran on Ubuntu 18.04 using LinuxDeply android app.

# You may need to tweak the below if you want cross compile for different ARM
# Orange pi is 7, but x96 is 8
# export GOARM=7
# export GOARCH=arm

VER=$(git rev-parse --short HEAD)
BUILD_VERSION=${BUILD_VERSION:-$VER}

sed -i "s/const Version = .*/const Version = \"${VER}\"/" models/version.go

#go build --tags "json1 fts5 secure_delete" -ldflags='-s -w'
go build --tags "json1 secure_delete" --ldflags '-extldflags "-static" -w -s'

CDIR=$(pwd)

ARCH=$(uname -m)

mkdir /tmp/webnote-$$/webnote-go-bin -p
cp -a assets webnote-go /tmp/webnote-$$/webnote-go-bin/
cd /tmp/webnote-$$
tar czf $CDIR/webnote-go-bin-${BUILD_VERSION}-${ARCH}.tgz webnote-go-bin
rm -rf /tmp/webnote-$$
echo "The archive ready to be extracted and run is webnote-go-bin-${BUILD_VERSION}-${ARCH}.tgz"
