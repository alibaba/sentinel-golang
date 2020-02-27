#!/usr/bin/env bash

go test -race -v -coverprofile=cover.out ./...
go tool cover -html=cover.out -o coverage.html
rm -f cover.out
