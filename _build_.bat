:: SPDX-License-Identifier: Apache-2.0
:: Copyright (c) 2026 Qirashi
:: Project: dvpl_go


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
go build -o ./out/dvpl-windows-x86_64/dvpl.exe -buildvcs=false -ldflags="-s -w -buildid=" -trimpath -buildmode=exe -tags=release -asmflags="-trimpath" -mod=readonly dvpl_go.go
if %ERRORLEVEL% neq 0 (
    echo Ошибка: Сборка завершилась с ошибкой. Код ошибки: %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
echo Сборка выполнена успешно.

where ResourceHacker >nul 2>nul
if %errorlevel% == 0 (
    echo Resource Hacker найден, выполняю команды...
    ResourceHacker -open ./out/dvpl-windows-x86_64/dvpl.exe -save ./out/dvpl-windows-x86_64/dvpl.exe -action addoverwrite -res ".\res\dvpl_go.ico" -mask ICONGROUP,MAINICON,
) else (
    echo Ошибка: Resource Hacker не найден в PATH.
    echo Иконка не установлена.
)

@pause
