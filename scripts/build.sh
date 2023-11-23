#!/usr/bin/env bash
set -e
set -x
. /etc/profile

go mod tidy
go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o "${GOBIN}/interxd"
go mod verify
echoInfo "INFO: Sucessfully intalled INTERX $(interxd version)"
