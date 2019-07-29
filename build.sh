#!/usr/bin/env bash
export GOPATH="$(pwd)/vendor"
# Set the GOPROXY environment variable
export GOPROXY=https://goproxy.io
echo ========================
echo Require packge
go mod tidy

echo ========================
echo build
# go build -ldflags "-s -w" -o bot bot.go
go build -ldflags "-s -w" -o main main.go
