#!/bin/sh

# Orange pi is 7, but x96 is 8
# export GOARM=7
# export GOARCH=arm

go build --tags "json1 fts5 secure_delete" -ldflags='-s -w'

CDIR=$(pwd)

mkdir /tmp/webnote-$$/webnote-go-bin -p
cp -a assets webnote-go /tmp/webnote-$$/webnote-go-bin/
cd /tmp/webnote-$$
tar czf $CDIR/webnote-go-bin.tgz webnote-go-bin
rm -rf /tmp/webnote-$$
