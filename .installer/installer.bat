@echo off
setlocal EnableExtensions EnableDelayedExpansion
title DVPL Installer
color 0A

:: ---------------- ADMIN CHECK ----------------

net session >nul 2>&1
if %errorlevel% neq 0 (
    echo Requesting administrator privileges...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit
)

:: ---------------- PATHS ----------------

set SCRIPT_DIR=%~dp0
set SRC=%SCRIPT_DIR%dvpl_go
set APP=C:\Tools\DvplGO

:: ---------------- INSTALL STATE ----------------

set SHOW_INSTALL=0
set SHOW_REINSTALL=0
set SHOW_UNINSTALL=0

if exist "%SRC%\dvpl.exe" if exist "%SRC%\dvpl_go.ico" (
    set SHOW_INSTALL=1
    set SHOW_REINSTALL=1
)

reg query HKCU\Software\XInstaller\DVPLGO >nul 2>nul
if %errorlevel%==0 set SHOW_UNINSTALL=1

:: ================= MAIN LOOP =================

:START
cls

call :READ_VALUES
call :READ_USER_PATH
call :DRAW_UI

set /p ACTION=Select option ^> 

if "%ACTION%"=="0" exit

if "%ACTION%"=="5" if %SHOW_INSTALL%==1 goto INSTALL
if "%ACTION%"=="6" if %SHOW_REINSTALL%==1 goto REINSTALL
if "%ACTION%"=="7" if %SHOW_UNINSTALL%==1 goto UNINSTALL

if "%ACTION%"=="1" set SETX_FLAG=
if "%ACTION%"=="2" goto DEL_USER
if "%ACTION%"=="3" goto ADD_PATH
if "%ACTION%"=="4" goto REMOVE_PATH

if not "%ACTION%"=="1" goto START

echo.
set /p WORKERS=MAX_WORKERS ^(1-99^): 
set /p COMPRESS=COMPRESS_TYPE ^(0,1,2^): 

:: ================= VALIDATION =================

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
echo.
echo Invalid values.
timeout /t 1 >nul
goto START


:: ================= DELETE ENV =================

:DEL_USER
reg delete HKCU\Environment /v DVPL_MAX_WORKERS /f >nul 2>nul
reg delete HKCU\Environment /v DVPL_COMPRESS_TYPE /f >nul 2>nul
goto START


:: ================= INSTALL =================

:INSTALL
cls
echo Installing DVPL Tools...
echo.

set ADD_COMPRESS_OPTIONS=
echo Do you want to add additional compression settings? (LZ4 and LZ4 HC options)
set /p ADD_COMPRESS_OPTIONS=[Y/N]: 

if not exist "%APP%" mkdir "%APP%"

copy "%SRC%\dvpl.exe" "%APP%" /Y >nul
copy "%SRC%\dvpl_go.ico" "%APP%" /Y >nul

reg add "HKCU\Software\XInstaller\DVPLGO" /v InstallLocation /t REG_SZ /d "%APP%" /f >nul
reg add "HKCU\Software\XInstaller\DVPLGO" /v InstallContext /t REG_SZ /d "" /f >nul

:: FILE CONTEXT MENU
reg add "HKCR\*\shell\DvplTools" /v MUIVerb /t REG_SZ /d "Dvpl Tools" /f
reg add "HKCR\*\shell\DvplTools" /v SubCommands /t REG_SZ /d "" /f
reg add "HKCR\*\shell\DvplTools" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
reg add "HKCR\*\shell\DvplTools" /v Position /t REG_SZ /d "Top" /f

reg add "HKCR\*\shell\DvplTools\shell\01Compress" /v MUIVerb /t REG_SZ /d "dvpl -c \"Compress\"" /f
reg add "HKCR\*\shell\DvplTools\shell\01Compress" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
reg add "HKCR\*\shell\DvplTools\shell\01Compress\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -c -i \"%%1\"" /f

reg add "HKCR\*\shell\DvplTools\shell\02Decompress" /v MUIVerb /t REG_SZ /d "dvpl -d \"Decompress\"" /f
reg add "HKCR\*\shell\DvplTools\shell\02Decompress" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
reg add "HKCR\*\shell\DvplTools\shell\02Decompress\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -d -i \"%%1\"" /f

if /i "%ADD_COMPRESS_OPTIONS%"=="Y" (
    reg add "HKCR\*\shell\DvplTools\shell\04CompressLZ4HC" /v CommandFlags /t REG_DWORD /d 32 /f

    reg add "HKCR\*\shell\DvplTools\shell\04CompressLZ4HC" /v MUIVerb /t REG_SZ /d "dvpl -c \"Compress LZ4 HC\"" /f
    reg add "HKCR\*\shell\DvplTools\shell\04CompressLZ4HC" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
    reg add "HKCR\*\shell\DvplTools\shell\04CompressLZ4HC\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -c -compress 1 -i \"%%1\"" /f

    reg add "HKCR\*\shell\DvplTools\shell\05CompressLZ4" /v MUIVerb /t REG_SZ /d "dvpl -c \"Compress LZ4\"" /f
    reg add "HKCR\*\shell\DvplTools\shell\05CompressLZ4" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
    reg add "HKCR\*\shell\DvplTools\shell\05CompressLZ4\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -c -compress 2 -i \"%%1\"" /f
)

:: DIRECTORY CONTEXT MENU
reg add "HKCR\Directory\shell\DvplTools" /v MUIVerb /t REG_SZ /d "Dvpl Tools" /f
reg add "HKCR\Directory\shell\DvplTools" /v SubCommands /t REG_SZ /d "" /f
reg add "HKCR\Directory\shell\DvplTools" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
reg add "HKCR\Directory\shell\DvplTools" /v Position /t REG_SZ /d "Top" /f

reg add "HKCR\Directory\shell\DvplTools\shell\01Compress" /v MUIVerb /t REG_SZ /d "dvpl -c \"Compress\"" /f
reg add "HKCR\Directory\shell\DvplTools\shell\01Compress" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
reg add "HKCR\Directory\shell\DvplTools\shell\01Compress\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -c -i \"%%1\"" /f

reg add "HKCR\Directory\shell\DvplTools\shell\02Decompress" /v MUIVerb /t REG_SZ /d "dvpl -d \"Decompress\"" /f
reg add "HKCR\Directory\shell\DvplTools\shell\02Decompress" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
reg add "HKCR\Directory\shell\DvplTools\shell\02Decompress\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -d -i \"%%1\"" /f

if /i "%ADD_COMPRESS_OPTIONS%"=="Y" (
    reg add "HKCR\Directory\shell\DvplTools\shell\04CompressLZ4HC" /v CommandFlags /t REG_DWORD /d 32 /f

    reg add "HKCR\Directory\shell\DvplTools\shell\04CompressLZ4HC" /v MUIVerb /t REG_SZ /d "dvpl -c \"Compress LZ4 HC\"" /f
    reg add "HKCR\Directory\shell\DvplTools\shell\04CompressLZ4HC" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
    reg add "HKCR\Directory\shell\DvplTools\shell\04CompressLZ4HC\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -c -compress 1 -i \"%%1\"" /f

    reg add "HKCR\Directory\shell\DvplTools\shell\05CompressLZ4" /v MUIVerb /t REG_SZ /d "dvpl -c \"Compress LZ4\"" /f
    reg add "HKCR\Directory\shell\DvplTools\shell\05CompressLZ4" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
    reg add "HKCR\Directory\shell\DvplTools\shell\05CompressLZ4\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -c -compress 2 -i \"%%1\"" /f
)

:: DIRECTORY BACKGROUND MENU
reg add "HKCR\Directory\Background\shell\DvplTools" /v MUIVerb /t REG_SZ /d "Dvpl Tools" /f
reg add "HKCR\Directory\Background\shell\DvplTools" /v SubCommands /t REG_SZ /d "" /f
reg add "HKCR\Directory\Background\shell\DvplTools" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
reg add "HKCR\Directory\Background\shell\DvplTools" /v Position /t REG_SZ /d "Top" /f

reg add "HKCR\Directory\Background\shell\DvplTools\shell\01Compress" /v MUIVerb /t REG_SZ /d "dvpl -c \"Compress\"" /f
reg add "HKCR\Directory\Background\shell\DvplTools\shell\01Compress" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
reg add "HKCR\Directory\Background\shell\DvplTools\shell\01Compress\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -c -i \"%%V\"" /f

reg add "HKCR\Directory\Background\shell\DvplTools\shell\02Decompress" /v MUIVerb /t REG_SZ /d "dvpl -d \"Decompress\"" /f
reg add "HKCR\Directory\Background\shell\DvplTools\shell\02Decompress" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
reg add "HKCR\Directory\Background\shell\DvplTools\shell\02Decompress\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -d -i \"%%V\"" /f

if /i "%ADD_COMPRESS_OPTIONS%"=="Y" (
    reg add "HKCR\Directory\Background\shell\DvplTools\shell\04CompressLZ4HC" /v CommandFlags /t REG_DWORD /d 32 /f

    reg add "HKCR\Directory\Background\shell\DvplTools\shell\04CompressLZ4HC" /v MUIVerb /t REG_SZ /d "dvpl -c \"Compress LZ4 HC\"" /f
    reg add "HKCR\Directory\Background\shell\DvplTools\shell\04CompressLZ4HC" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
    reg add "HKCR\Directory\Background\shell\DvplTools\shell\04CompressLZ4HC\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -c -compress 1 -i \"%%V\"" /f

    reg add "HKCR\Directory\Background\shell\DvplTools\shell\05CompressLZ4" /v MUIVerb /t REG_SZ /d "dvpl -c \"Compress LZ4\"" /f
    reg add "HKCR\Directory\Background\shell\DvplTools\shell\05CompressLZ4" /v Icon /t REG_SZ /d "%APP%\dvpl_go.ico" /f
    reg add "HKCR\Directory\Background\shell\DvplTools\shell\05CompressLZ4\command" /ve /t REG_SZ /d "\"%APP%\dvpl.exe\" -c -compress 2 -i \"%%V\"" /f
)

echo.
echo Installation complete.
timeout /t 2 >nul
goto RELOAD


:: ================= REINSTALL =================

:REINSTALL
set SILENT=1
call :UNINSTALL
set SILENT=
goto INSTALL


:: ================= UNINSTALL =================

:UNINSTALL
cls
echo Uninstalling DVPL Tools...
echo.

reg delete "HKCR\*\shell\DvplTools" /f >nul 2>nul
reg delete "HKCR\Directory\shell\DvplTools" /f >nul 2>nul
reg delete "HKCR\Directory\Background\shell\DvplTools" /f >nul 2>nul

reg delete HKCU\Software\XInstaller\DVPLGO /f >nul 2>nul

if exist "%APP%" rmdir /s /q "%APP%"

echo.
echo Uninstall complete.
timeout /t 2 >nul

if defined SILENT exit /b

goto RELOAD


:: ================= RELOAD =================

:RELOAD
set SHOW_INSTALL=0
set SHOW_REINSTALL=0
set SHOW_UNINSTALL=0

if exist "%SRC%\dvpl.exe" if exist "%SRC%\dvpl_go.ico" (
    set SHOW_INSTALL=1
    set SHOW_REINSTALL=1
)

reg query HKCU\Software\XInstaller\DVPLGO >nul 2>nul
if %errorlevel%==0 set SHOW_UNINSTALL=1

goto START


:: ================= UI =================

:DRAW_UI

echo =====================================================
echo                 DVPL Installer ^| CLI
echo =====================================================
echo.

call :PRINT_BLOCK User "%U_WORKERS%" "%U_COMPRESS%"

echo.
echo -----------------------------------------------------
echo  [1]  Set USER variables
echo  [2]  Del USER variables
echo.
echo  [3]  Add DVPL to PATH
echo  [4]  Remove DVPL from PATH
echo.

if %SHOW_INSTALL%==1   echo  [5]  Install
if %SHOW_REINSTALL%==1 echo  [6]  Reinstall
if %SHOW_UNINSTALL%==1 echo  [7]  Uninstall

echo.
echo  [0]  Exit
echo -----------------------------------------------------
echo.

exit /b


:: ================= HELPERS =================

:READ_VALUES
set U_WORKERS=
set U_COMPRESS=

for /f "tokens=2,*" %%A in ('reg query HKCU\Environment /v DVPL_MAX_WORKERS 2^>nul') do set U_WORKERS=%%B
for /f "tokens=2,*" %%A in ('reg query HKCU\Environment /v DVPL_COMPRESS_TYPE 2^>nul') do set U_COMPRESS=%%B

exit /b


:READ_USER_PATH
set USER_PATH=
for /f "tokens=2,*" %%A in ('reg query HKCU\Environment /v Path 2^>nul') do set USER_PATH=%%B
exit /b


:ADD_PATH
set FOUND=0
set CLEAN_PATH=

for %%A in ("%USER_PATH:;=" "%") do (
    set PART=%%~A
    if not "!PART!"=="" (
        if /I "!PART!"=="%APP%" set FOUND=1
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
    set NEWPATH=!CLEAN_PATH!;%APP%
) else (
    set NEWPATH=%APP%
)

reg add HKCU\Environment /v Path /t REG_EXPAND_SZ /d "!NEWPATH!" /f >nul

goto START


:REMOVE_PATH
set NEWPATH=

for %%A in ("%USER_PATH:;=" "%") do (
    set PART=%%~A
    if not "!PART!"=="" (
        if /I not "!PART!"=="%APP%" (
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

echo %NAME% variables:
if "%W%%C%"=="" (
    echo   No variables
) else (
    if not "%W%"=="" echo   DVPL_MAX_WORKERS   = %W%
    if not "%C%"=="" echo   DVPL_COMPRESS_TYPE = %C%
)

set PATH_STATUS=Not installed
echo %USER_PATH% | find /I "C:\Tools\DvplGO" >nul 2>nul
if %errorlevel%==0 set PATH_STATUS=Installed

echo   Path               = %PATH_STATUS%

exit /b
