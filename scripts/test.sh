#!/usr/bin/env bash
set -e
set -x
. /etc/profile

echo "INFO: Cleaning up system resources"
go test -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -mod=readonly $(go list ./gateway/...) 
