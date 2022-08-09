@echo OFF

set "args=%*"
pushd "%~dp0"
setlocal ENABLEDELAYEDEXPANSION
set GOPATH="%~dp0vendor"
rem Set the GOPROXY environment variable
Set GOPROXY=https://goproxy.io,direct
rem set http_proxy=socks5://127.0.0.1:1080
rem set https_proxy=%http_proxy%

if /i "%args%"=="default" goto %args%
if /i "%args%"=="static-server" goto %args%
if /i "%args%"=="deps" goto %args%
if /i "%args%"=="fmt" goto %args%
if /i "%args%"=="server" goto %args%
if /i "%args%"=="static-release-all" goto %args%
if /i "%args%"=="release-all" goto %args%
if /i "%args%"=="all" goto %args%
if /i "%args%"=="static-all" goto %args%
if /i "%args%"=="run" goto %args%

if /i "%BUILDTAGS%"=="" (Set BUILDTAGS=release)

:default
    CALL :all
    GOTO :EOF

:deps
    CALL go mod tidy -v
    rem CALL go mod vendor -v
    GOTO :EOF

:fmt
    CALL go fmt -mod=mod ./...
    GOTO :EOF

:static-server
    Set CGO_ENABLED=0
    CALL go install --tags "%BUILDTAGS%" -ldflags "-s -w -extldflags ""-static""" -mod=mod github.com/larisgo/laravel-echo-server/main/laravel-echo-server
    GOTO :EOF

:server
    CALL go install --tags "%BUILDTAGS%" -ldflags "-s -w" -mod=mod github.com/larisgo/laravel-echo-server/main/laravel-echo-server
    GOTO :EOF

:release
    Set BUILDTAGS=release
    GOTO :EOF

:static-release-all
    CALL :fmt
    CALL :release
    CALL :static-server
    GOTO :EOF

:release-all
    CALL :fmt
    CALL :release
    CALL :server
    GOTO :EOF

:all
    CALL :fmt
    CALL :server
    GOTO :EOF

:static-all
    CALL :fmt
    CALL :static-server
    GOTO :EOF

:clean
    CALL go clean -mod=mod -r ./...

:run
    Set GOOS=
    Set GOARCH=
    CALL go install -mod=mod github.com/larisgo/laravel-echo-server/main/laravel-echo-server
    CALL "vendor\bin\laravel-echo-server.exe"
