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
go build -o ./out/dvpl.exe -buildvcs=false -ldflags="-s -w -buildid=" -trimpath -buildmode=exe -tags=release -asmflags="-trimpath" -mod=readonly dvpl_go.go
if %ERRORLEVEL% neq 0 (
    echo Ошибка: Сборка завершилась с ошибкой. Код ошибки: %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
echo Сборка выполнена успешно.

where ResourceHacker >nul 2>nul
if %errorlevel% == 0 (
    echo Resource Hacker найден, выполняю команды...
    ResourceHacker -open ./out/dvpl.exe -save ./out/dvpl.exe -action addoverwrite -res ".\res\dvpl_go.ico" -mask ICONGROUP,MAINICON,
) else (
    echo Ошибка: Resource Hacker не найден в PATH.
    echo Иконка не установлена.
)

echo.
echo Копирование файлов в ./out/.installer/dvpl_go/ ...

if not exist ".installer\dvpl_go\" mkdir ".installer\dvpl_go\"

if exist ".\out\dvpl.exe" (
    copy /Y ".\out\dvpl.exe" ".installer\dvpl_go\" >nul
) else (
    echo Предупреждение: .\out\dvpl.exe не найден.
)

if exist ".\res\dvpl_go.ico" (
    copy /Y ".\res\dvpl_go.ico" ".installer\dvpl_go\" >nul
) else (
    echo Предупреждение: .\res\dvpl_go.ico не найден.
)

if exist ".installer\dvpl_go\dvpl.exe" if exist ".installer\dvpl_go\dvpl_go.ico" (
    echo Файлы успешно скопированы.
) else (
    echo Ошибка: файлы не были скопированы.
)

@pause
