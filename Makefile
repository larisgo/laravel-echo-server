.PHONY: default install all run
export GOPATH:=$(shell pwd)/vendor
export GOPROXY=https://goproxy.io


default:
	go build -ldflags "-s -w" -o bin/laravel-echo-server main.go

install:
	go mod tidy

all:
	@echo "build darwin_386/laravel-echo-server"
	GOOS=darwin GOARCH=386 go build -ldflags "-s -w" -o "bin/darwin_386/laravel-echo-server" main.go
	@echo "build darwin_amd64/laravel-echo-server"
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o "bin/darwin_amd64/laravel-echo-server" main.go
	@echo "build linux_386/laravel-echo-server"
	GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o "bin/linux_386/laravel-echo-server" main.go
	@echo "build linux_amd64/laravel-echo-server"
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "bin/linux_amd64/laravel-echo-server" main.go
	@echo "build linux_arm/laravel-echo-server"
	GOOS=linux GOARCH=arm go build -ldflags "-s -w" -o "bin/linux_arm/laravel-echo-server" main.go
	@echo "build linux_arm64/laravel-echo-server"
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o "bin/linux_arm64/laravel-echo-server" main.go
	@echo "build windows_386/laravel-echo-server.exe"
	GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o "bin/windows_386/laravel-echo-server.exe" main.go
	@echo "build windows_amd64/laravel-echo-server.exe"
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o "bin/windows_amd64/laravel-echo-server.exe" main.go

run:
	go build -o bin/laravel-echo-server main.go
	sudo chmod +x bin/laravel-echo-server
	bin/laravel-echo-server