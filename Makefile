.PHONY: default static-server deps fmt server static-release-all release-all all static-all run
export GOPATH:=$(shell pwd)/vendor
# Set the GOPROXY environment variable
export GOPROXY=https://goproxy.io,direct
# export http_proxy=socks5://127.0.0.1:1080
# export https_proxy=%http_proxy%

BUILDTAGS=release
default: all

deps:
	go mod tidy -v
# 	go mod vendor -v

fmt:
	go fmt -mod=mod ./...

static-server:
	CGO_ENABLED=0 go install --tags '$(BUILDTAGS)' -ldflags '-s -w -extldflags "-static"' -mod=mod github.com/larisgo/laravel-echo-server/main/laravel-echo-server

server:
	go install --tags '$(BUILDTAGS)' -ldflags '-s -w' -mod=mod github.com/larisgo/laravel-echo-server/main/laravel-echo-server

release: BUILDTAGS=release

static-release-all: fmt release static-server
release-all: fmt release server

all: fmt server
static-all: fmt static-server

clean:
	go clean -mod=mod -r ./...

run:
	GOOS="" GOARCH="" go install -mod=mod github.com/larisgo/laravel-echo-server/main/laravel-echo-server
	vendor/bin/laravel-echo-server