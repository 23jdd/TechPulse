#!/usr/bin/env sh
set -eu
go run ./cmd/worker -mode=migrate
