#!/usr/bin/env bash
set -e
set -x
. /etc/profile

go mod tidy
CGO_ENABLED=0 go build -a -tags netgo -installsuffix cgo -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore -extldflags '-static'" -o "/interxd" .
go mod verify
echoInfo "INFO: Sucessfully intalled INTERX $(interxd version)"
