#!/bin/sh
docker tag golang-alpine-build-jenkins:latest golang-alpine-build-jenkins:backup
docker commit golang-alpine-build-jenkins golang-alpine-build-jenkins:latest
