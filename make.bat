@echo OFF

set "args=%*"
pushd "%~dp0"
setlocal ENABLEDELAYEDEXPANSION
set GOPATH="%~dp0vendor"
rem Set the GOPROXY environment variable
set GOPROXY=https://goproxy.io

if /i "%args%"=="install" goto install
if /i "%args%"=="all" goto all
if /i "%args%"=="run" goto run

goto DEFAULT_CASE
:install
    mkdir vendor
    CALL go mod tidy
    GOTO END_CASE
:all
    echo ========================
    echo build darwin_386/laravel-echo-server
    set GOOS=darwin
    set GOARCH=386
    CALL go build -ldflags "-s -w" -o "bin/darwin_386/laravel-echo-server" main.go

    echo ========================
    echo build darwin_amd64/laravel-echo-server
    set GOOS=darwin
    set GOARCH=amd64
    CALL go build -ldflags "-s -w" -o "bin/darwin_amd64/laravel-echo-server" main.go

    echo ========================
    echo build linux_386/laravel-echo-server
    set GOOS=linux
    set GOARCH=386
    CALL go build -ldflags "-s -w" -o "bin/linux_386/laravel-echo-server" main.go

    echo ========================
    echo build linux_amd64/laravel-echo-server
    set GOOS=linux
    set GOARCH=amd64
    CALL go build -ldflags "-s -w" -o "bin/linux_amd64/laravel-echo-server" main.go

    echo ========================
    echo build linux_arm/laravel-echo-server
    set GOOS=linux
    set GOARCH=arm
    CALL go build -ldflags "-s -w" -o "bin/linux_arm/laravel-echo-server" main.go

    echo ========================
    echo build linux_arm64/laravel-echo-server
    set GOOS=linux
    set GOARCH=arm64
    CALL go build -ldflags "-s -w" -o "bin/linux_arm64/laravel-echo-server" main.go

    echo ========================
    echo build windows_386/laravel-echo-server.exe
    set GOOS=windows
    set GOARCH=386
    CALL go build -ldflags "-s -w" -o "bin/windows_386/laravel-echo-server.exe" main.go

    echo ========================
    echo build windows_amd64/laravel-echo-server.exe
    set GOOS=windows
    set GOARCH=amd64
    CALL go build -ldflags "-s -w" -o "bin/windows_amd64/laravel-echo-server.exe" main.go

    GOTO END_CASE
:run
    CALL go build -o bin\main.exe main.go && CALL %~dp0\bin\main.exe
    GOTO END_CASE
:DEFAULT_CASE
    CALL go mod tidy
    CALL go build -ldflags "-s -w" -o bin\main.exe main.go
    GOTO END_CASE
:END_CASE
    GOTO :EOF