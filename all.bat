@echo off
setlocal enabledelayedexpansion

for %%i in ("%~dp0.") do SET "script_dir=%%~fi"
cd "%script_dir%"
for %%i in ("%~dp0.") do SET "dirname=%%~ni"

call build.bat %*
if errorlevel 1 (
    echo ERROR BUILD 1>&2
    exit /b 1
)
if exist senv.bat (
    call senv.bat
)
call "%dirname%"