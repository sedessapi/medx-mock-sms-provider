#!/usr/bin/env bash
# CI entry point: build, vet and test every package.
# No package is excluded. If one has to be, list it here with the reason.
#
# Dependencies are downloaded first. After that all outbound HTTP goes to a dead proxy, so a
# test cannot reach a live service or a public production URL; loopback (httptest) still works.
set -euo pipefail

export CGO_ENABLED=0
go mod download

export HTTP_PROXY=http://127.0.0.1:1 HTTPS_PROXY=http://127.0.0.1:1
export http_proxy=$HTTP_PROXY https_proxy=$HTTPS_PROXY
export NO_PROXY=127.0.0.1,localhost,::1 no_proxy=127.0.0.1,localhost,::1

go build ./...
go vet ./...
go test -count=1 ./...
