@echo off
setlocal EnableExtensions EnableDelayedExpansion

:START
cls

call :READ_VALUES

echo DVPL Environment Variables:
echo ---------------------------

call :PRINT_BLOCK User "%U_WORKERS%" "%U_COMPRESS%"
call :PRINT_BLOCK System "%S_WORKERS%" "%S_COMPRESS%"

echo.
echo [1] Set user
echo [2] Set system (admin)
echo [3] Delete user
echo [4] Delete system
echo [0] Exit
echo.

set /p ACTION="> "

if "%ACTION%"=="0" exit /b
if "%ACTION%"=="1" set SETX_FLAG=
if "%ACTION%"=="2" set SETX_FLAG=/M
if "%ACTION%"=="3" goto DEL_USER
if "%ACTION%"=="4" goto DEL_SYS
if not "%ACTION%"=="1" if not "%ACTION%"=="2" goto START

:: --- Input ---
set /p WORKERS="MAX_WORKERS (1-99): "
set /p COMPRESS="COMPRESS_TYPE (1,2,4,5): "

:: --- Validation ---
for %%V in ("%WORKERS%" "%COMPRESS%") do (
    for /f "delims=0123456789" %%A in (%%V) do goto INVALID
)

if %WORKERS% LSS 1 goto INVALID
if %WORKERS% GTR 99 goto INVALID
if "%COMPRESS%"=="3" goto INVALID
if %COMPRESS% LSS 1 goto INVALID
if %COMPRESS% GTR 5 goto INVALID

setx DVPL_MAX_WORKERS %WORKERS% %SETX_FLAG% >nul
setx DVPL_COMPRESS_TYPE %COMPRESS% %SETX_FLAG% >nul
goto START

:INVALID
echo Invalid values.
timeout /t 1 >nul
goto START

:DEL_USER
reg delete HKCU\Environment /v DVPL_MAX_WORKERS /f >nul 2>nul
reg delete HKCU\Environment /v DVPL_COMPRESS_TYPE /f >nul 2>nul
goto START

:DEL_SYS
reg delete "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v DVPL_MAX_WORKERS /f >nul 2>nul
reg delete "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v DVPL_COMPRESS_TYPE /f >nul 2>nul
goto START

:: ===================== helpers =====================

:READ_VALUES
set U_WORKERS=
set U_COMPRESS=
set S_WORKERS=
set S_COMPRESS=

for /f "tokens=2,*" %%A in ('reg query HKCU\Environment /v DVPL_MAX_WORKERS 2^>nul') do set U_WORKERS=%%B
for /f "tokens=2,*" %%A in ('reg query HKCU\Environment /v DVPL_COMPRESS_TYPE 2^>nul') do set U_COMPRESS=%%B
for /f "tokens=2,*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v DVPL_MAX_WORKERS 2^>nul') do set S_WORKERS=%%B
for /f "tokens=2,*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v DVPL_COMPRESS_TYPE 2^>nul') do set S_COMPRESS=%%B

exit /b

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
