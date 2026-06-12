:: SPDX-License-Identifier: Apache-2.0
:: Copyright (c) 2026 Qirashi
:: Project: dvpl_go


@echo off
chcp 65001

echo Modifying code...
go fix ./...

echo Starting build...
set CC=zig cc -target x86_64-linux
set CXX=zig c++ -target x86_64-linux
set CGO_ENABLED=1
set GOOS=linux
set GOARCH=amd64
go build -o ./out/dvpl-linux-x86_64/dvpl -buildvcs=false -ldflags="-s -w -buildid=" -trimpath -buildmode=exe -tags=release -asmflags="-trimpath" -mod=readonly dvpl_go.go
if %ERRORLEVEL% neq 0 (
    echo The build failed with error: %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
echo The build completed successfully.

@pause
