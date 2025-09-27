:: SPDX-License-Identifier: Apache-2.0
:: Copyright (c) 2025 Qirashi
:: Project: dvpl_go

echo off
chcp 65001

echo Начинаю сборку.
set CGO_ENABLED=1
go build -o ./out/dvpl.exe -buildvcs=false -ldflags="-s -w -buildid=" -trimpath -buildmode=exe -tags=release -asmflags="-trimpath" -mod=readonly dvpl_go.go
if %ERRORLEVEL% neq 0 (
    echo Ошибка: Сборка завершилась с ошибкой. Код ошибки: %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
echo Сборка выполнена успешно.


set ResHack="R:\Program_Files\resource_hacker\ResourceHacker.exe"
if exist "%ResHack%" (
    echo Resource Hacker найден, выполняю команды...
    "%ResHack%" -open ./out/dvpl.exe -save ./out/dvpl.exe -action addoverwrite -res ".\res\GO_ICO.ico" -mask ICONGROUP,MAINICON,
) else (
    echo Ошибка: Resource Hacker не найден по пути "%ResHack%".
	echo Иконка не установлена.
)

REM where upx >nul 2>&1
REM if %errorlevel% equ 0 (
    REM echo UPX найден в PATH, выполняю команды...
    REM upx -7 "%cd%\out\dvpl.exe"
REM ) else (
    REM echo UPX не найден в переменных среды PATH.
    REM echo Exe не сжат UPX.
REM )

@pause
