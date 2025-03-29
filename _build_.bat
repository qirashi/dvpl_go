:: SPDX-License-Identifier: Apache-2.0
:: Copyright (c) 2025 Qirashi
:: Project: dvpl_go

echo off
chcp 65001

echo Начинаю сборку .exe
go build -o ./out/dvpl_go.exe -ldflags="-s -w" -trimpath -buildmode=exe -tags=release -asmflags="-trimpath" -mod=readonly dvpl_go.go
if %ERRORLEVEL% neq 0 (
    echo Ошибка: Сборка завершилась с ошибкой. Код ошибки: %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
echo .exe Успешно собран.


set ResHack="R:\Program_Files\resource_hacker\ResourceHacker.exe"
if exist "%ResHack%" (
    echo Resource Hacker найден, выполняю команды...
    "%ResHack%" -open ./out/dvpl_go.exe -save ./out/dvpl_go.exe -action addoverwrite -res ".\res\GO_ICO.ico" -mask ICONGROUP,MAINICON,
) else (
    echo Ошибка: Resource Hacker не найден по пути "%ResHack%".
	echo Иконка не установлена.
)

set TOOL1=%cd%\tools\upx.exe
if defined UPX (
    set "UPX="
)
if exist "%TOOL1%" (
    echo UPX найден, выполняю команды...
    "%TOOL1%" -9 "%cd%\out\dvpl_go.exe"
) else (
    echo Ошибка: UPX не найден по пути "%TOOL1%".
	echo Exe не сжат UPX.
)

@pause
