:: SPDX-License-Identifier: Apache-2.0
:: Copyright (c) 2026 Qirashi
:: Project: dvpl_go


@echo off
chcp 65001

echo Modifying code...
go fix ./...

echo Starting build...
set CC=gcc
set CXX=g++
set CGO_ENABLED=1
set CGO_CFLAGS=-ffunction-sections -fdata-sections
set CGO_LDFLAGS=-Wl,--gc-sections
set GOOS=windows
set GOARCH=amd64
go build -o ./out/dvpl-windows-x86_64/dvpl.exe -buildvcs=false -ldflags="-s -w -buildid=" -trimpath -buildmode=exe -tags=release -asmflags="-trimpath" -mod=readonly dvpl_go.go
if %ERRORLEVEL% neq 0 (
    echo Error: Build failed. Error code: %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
echo Build completed successfully.

where ResourceHacker >nul 2>nul
if %errorlevel% == 0 (
    echo Resource Hacker found, executing commands...
    ResourceHacker -open ./out/dvpl-windows-x86_64/dvpl.exe -save ./out/dvpl-windows-x86_64/dvpl.exe -action addoverwrite -res ".\dvpl_go.ico" -mask ICONGROUP,MAINICON,
    ResourceHacker -open ./out/dvpl-windows-x86_64/dvpl.exe -save ./out/dvpl-windows-x86_64/dvpl.exe -action addoverwrite -res ".\dvpl_go.manifest" -mask MANIFEST,1,

    ResourceHacker -open "./dvpl_go.rc" -save "./dvpl_go.res" -action compile -log CONSOLE
    ResourceHacker -open ./out/dvpl-windows-x86_64/dvpl.exe -save ./out/dvpl-windows-x86_64/dvpl.exe -action addoverwrite -res ".\dvpl_go.res" -mask VERSIONINFO,
) else (
    echo Error: Resource Hacker not found in PATH.
)

@pause
