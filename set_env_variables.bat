@echo off
setlocal EnableExtensions EnableDelayedExpansion

set TARGET=C:\Tools\DvplGO

:START
cls

call :READ_VALUES
call :READ_USER_PATH

echo DVPL Environment Variables:
echo ---------------------------
call :PRINT_BLOCK User "%U_WORKERS%" "%U_COMPRESS%"

@REM echo.
@REM echo PATH:
@REM echo -----
@REM echo %USER_PATH%

echo.
echo [1] Set USER variables 
echo [2] Delete USER variables 
echo [ ] 
echo [3] Add DVPL to PATH 
echo [4] Remove DVPL from PATH 
echo [ ] 
echo [0] Exit 
echo.

set /p ACTION="> "

if "%ACTION%"=="0" exit /b
if "%ACTION%"=="1" goto SET_USER
if "%ACTION%"=="2" goto DEL_USER
if "%ACTION%"=="3" goto ADD_PATH
if "%ACTION%"=="4" goto REMOVE_PATH
goto START


:SET_USER
set /p WORKERS="DVPL_MAX_WORKERS (1-99): "
set /p COMPRESS="DVPL_COMPRESS_TYPE (1-2): "

for %%V in ("%WORKERS%" "%COMPRESS%") do (
    for /f "delims=0123456789" %%A in (%%V) do goto INVALID
)

if %WORKERS% LSS 1 goto INVALID
if %WORKERS% GTR 99 goto INVALID
if %COMPRESS% LSS 1 goto INVALID
if %COMPRESS% GTR 2 goto INVALID

setx DVPL_MAX_WORKERS %WORKERS% >nul
setx DVPL_COMPRESS_TYPE %COMPRESS% >nul

goto START


:INVALID
echo Invalid values.
timeout /t 1 >nul
goto START


:DEL_USER
reg delete HKCU\Environment /v DVPL_MAX_WORKERS /f >nul 2>nul
reg delete HKCU\Environment /v DVPL_COMPRESS_TYPE /f >nul 2>nul
goto START


:READ_VALUES
set U_WORKERS=
set U_COMPRESS=

for /f "tokens=2,*" %%A in ('reg query HKCU\Environment /v DVPL_MAX_WORKERS 2^>nul') do set U_WORKERS=%%B
for /f "tokens=2,*" %%A in ('reg query HKCU\Environment /v DVPL_COMPRESS_TYPE 2^>nul') do set U_COMPRESS=%%B

exit /b


:READ_USER_PATH
set USER_PATH=

for /f "tokens=2,*" %%A in ('reg query HKCU\Environment /v Path 2^>nul') do (
    set USER_PATH=%%B
)

exit /b


:ADD_PATH
set FOUND=0
set CLEAN_PATH=

for %%A in ("%USER_PATH:;=" "%") do (
    set PART=%%~A

    if not "!PART!"=="" (

        if /I "!PART!"=="%TARGET%" set FOUND=1

        if defined CLEAN_PATH (
            set CLEAN_PATH=!CLEAN_PATH!;!PART!
        ) else (
            set CLEAN_PATH=!PART!
        )
    )
)

if "!FOUND!"=="1" (
    reg add HKCU\Environment /v Path /t REG_EXPAND_SZ /d "!CLEAN_PATH!" /f >nul
    goto START
)

if defined CLEAN_PATH (
    set NEWPATH=!CLEAN_PATH!;%TARGET%
) else (
    set NEWPATH=%TARGET%
)

reg add HKCU\Environment /v Path /t REG_EXPAND_SZ /d "!NEWPATH!" /f >nul

goto START


:REMOVE_PATH
set NEWPATH=

for %%A in ("%USER_PATH:;=" "%") do (

    set PART=%%~A

    if not "!PART!"=="" (

        if /I not "!PART!"=="%TARGET%" (

            if defined NEWPATH (
                set NEWPATH=!NEWPATH!;!PART!
            ) else (
                set NEWPATH=!PART!
            )

        )
    )
)

reg add HKCU\Environment /v Path /t REG_EXPAND_SZ /d "!NEWPATH!" /f >nul

goto START


:PRINT_BLOCK
set NAME=%~1
set W=%~2
set C=%~3

echo %NAME%:
if "%W%%C%"=="" (
    echo   No variables.
) else (
    if not "%W%"=="" echo   DVPL_MAX_WORKERS   = %W%
    if not "%C%"=="" echo   DVPL_COMPRESS_TYPE = %C%
)
exit /b