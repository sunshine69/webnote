#!/bin/bash

# this run on ubuntu 22.04 amd64 host

# see doc https://docs.docker.com/engine/reference/commandline/buildx/
# Install build-x plugin in. See https://docs.docker.com/buildx/working-with-buildx/ . Download the binary and save it to $HOME/.docker/cli-plugins/docker-buildx - remember to chmod +x it
# https://github.com/docker/buildx/releases

# Install qemu - apt-get install -y qemu-user-static
TAG=$(date '+Y%m%d')
APP_VERSION=$(git rev-parse --short HEAD)

# docker buildx create --name mybuilder
docker buildx use mybuilder
docker buildx build -t stevekieu/webnote-azure-app:${TAG} --build-arg APP_VERSION=${APP_VERSION}  --platform linux/amd64,linux/arm64 --push -f Dockerfile.azure-app-svc .
