:: SPDX-License-Identifier: Apache-2.0
:: Copyright (c) 2026 Qirashi
:: Project: dvpl_go


echo off
chcp 65001

echo Модернизация кода...
go fix ./...

echo Начинаю сборку.
set CGO_ENABLED=1
go build -o ./out/dvpl.exe -buildvcs=false -ldflags="-s -w -buildid=" -trimpath -buildmode=exe -tags=release -asmflags="-trimpath" -mod=readonly dvpl_go.go
if %ERRORLEVEL% neq 0 (
    echo Ошибка: Сборка завершилась с ошибкой. Код ошибки: %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
echo Сборка выполнена успешно.

where ResourceHacker >nul 2>nul
if %errorlevel% == 0 (
    echo Resource Hacker найден, выполняю команды...
    ResourceHacker -open ./out/dvpl.exe -save ./out/dvpl.exe -action addoverwrite -res ".\res\GO_ICO.ico" -mask ICONGROUP,MAINICON,
) else (
    echo Ошибка: Resource Hacker не найден в PATH.
    echo Иконка не установлена.
)

@pause
