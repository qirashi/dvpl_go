:: SPDX-License-Identifier: Apache-2.0
:: Copyright (c) 2026 Qirashi
:: Project: installer


@echo off
chcp 65001

copy /Y "..\dvpl_go\dvpl_go.ico" "./assets\dvpl_go.ico"
copy /Y "..\dvpl_go\out\dvpl-windows-x86_64\dvpl.exe" "./assets\dvpl.exe"

echo Modernizing code...
go fix ./...

echo Starting build...
set CC=
set CXX=
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -o ./out/installer.exe -buildvcs=false -ldflags="-s -w -buildid=" -trimpath -buildmode=exe -tags=release -asmflags="-trimpath" -mod=readonly installer.go
if %ERRORLEVEL% neq 0 (
    echo Error: Build failed. Error code: %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
echo Build completed successfully.

where ResourceHacker >nul 2>nul
if %errorlevel% == 0 (
    echo Resource Hacker found, executing commands...
    ResourceHacker -open ./out/installer.exe -save ./out/installer.exe -action addoverwrite -res ".\installer.ico" -mask ICONGROUP,MAINICON,
    ResourceHacker -open ./out/installer.exe -save ./out/installer.exe -action addoverwrite -res ".\installer.manifest" -mask MANIFEST,1,

    ResourceHacker -open "./installer.rc" -save "./installer.res" -action compile -log CONSOLE
    ResourceHacker -open ./out/installer.exe -save ./out/installer.exe -action addoverwrite -res ".\installer.res" -mask VERSIONINFO,
) else (
    echo Error: Resource Hacker not found in PATH.
)

@pause
