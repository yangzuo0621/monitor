#!/usr/bin/env bash

docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -w /go/src/github.com/yangzuo0621/monitor -u 1001:1001 -e VERSION=dev-fun-cluster-ev2-test-47 -e BUILDING_FLAVOUR=official -e GOFLAGS=-mod=vendor -v "$PWD":/go/src/github.com/yangzuo0621/monitor golang:1.13 go build -buildmode=pie -o bin/monitor github.com/yangzuo0621/monitor/cmd/monitor
