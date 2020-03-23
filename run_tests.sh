#!/bin/sh

set -eu

go test ./...
cd standalone_test
go build
./standalone_test
cd ..
