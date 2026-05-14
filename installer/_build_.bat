:: SPDX-License-Identifier: Apache-2.0
:: Copyright (c) 2026 Qirashi
:: Project: installer


echo off
chcp 65001

echo Модернизация кода...
go fix ./...

echo Начинаю сборку.
set CC=
set CXX=
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -o ./out/installer.exe -buildvcs=false -ldflags="-s -w -buildid=" -trimpath -buildmode=exe -tags=release -asmflags="-trimpath" -mod=readonly installer.go
if %ERRORLEVEL% neq 0 (
    echo Ошибка: Сборка завершилась с ошибкой. Код ошибки: %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
echo Сборка выполнена успешно.

where ResourceHacker >nul 2>nul
if %errorlevel% == 0 (
	echo Resource Hacker найден, выполняю команды...
	ResourceHacker -open ./out/installer.exe -save ./out/installer.exe -action addoverwrite -res ".\installer.ico" -mask ICONGROUP,MAINICON,
	ResourceHacker -open ./out/installer.exe -save ./out/installer.exe -action addoverwrite -res ".\installer.manifest" -mask MANIFEST,1,
) else (
    echo Ошибка: Resource Hacker не найден в PATH.
)

@pause
